package app

import (
	"context"
	"flag"
	"github.com/caarlos0/env"
	"github.com/gtgaleevtimur/gofermart/internal/handler"
	"github.com/gtgaleevtimur/gofermart/internal/repository"
	s "github.com/gtgaleevtimur/gofermart/internal/service"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type config struct {
	Addr                 string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func Run() {
	cfg := new(config)
	flag.StringVar(&cfg.Addr, "a", ":8080", "Service run address")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "Postgres URI")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "http://localhost:8081", "Accrual system address")
	flag.Parse()

	err := env.Parse(cfg)
	if err != nil {
		log.Fatal().Err(err)
	}
	// Инициализация канала Grace-ful Shutdown.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	// Инициализация хранилища сервиса.
	db, err := repository.NewDatabaseDSN(cfg.DatabaseURI)
	if err != nil {
		log.Info().Msg("Postgres init failed.")
		log.Fatal().Err(err)
	}
	// Инициализация сервиса.
	gophermart := s.NewService(db)
	// Инициализация роутера.
	hand := handler.NewHandler(gophermart)
	// Инициализация сервера.
	server := http.Server{
		Addr:    cfg.Addr,
		Handler: hand.R(),
	}
	// Инициализация канала для ошибки сервера.
	serverErrors := make(chan error, 1)
	// Запуск сервера.
	go func() {
		log.Info().Msg("Service started.")
		serverErrors <- server.ListenAndServe()
	}()
	// Grace-ful Shutdown.
	go func() {
		select {
		case err = <-serverErrors:
			if err != http.ErrServerClosed {
				log.Info().Msg("Server error.")
				log.Fatal().Err(err)
			}
		case <-shutdown:
			log.Info().Msg(" Grace-ful Shutdown started.")
			defer log.Info().Msg("Grace-ful Shutdown completed.")

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Принудительно завершаем работу по истечении времени контекста.
			go func() {
				<-ctx.Done()
				if ctx.Err() == context.DeadlineExceeded {
					log.Fatal().Msg(" Grace-ful shutdown timed out. Forcing exit.")
				}
			}()

			// Штатно завершаем работу сервера.
			if err = server.Shutdown(ctx); err != nil {
				log.Info().Msg("could not stop server gracefully")
				log.Fatal().Err(err)
			}
		}
	}()
	blackbox := s.NewAccrl(db, cfg.AccrualSystemAddress)
	blackbox.Run()
}
