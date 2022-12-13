package app

import (
	"context"
	"github.com/gtgaleevtimur/gofermart/internal/config"
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

func Run() {
	log.Info().Msg("Start service `Gofermart`.")
	// Конфигурация приложения через считывание флагов и переменных окружения.
	conf := config.NewConfig()
	log.Info().Msg("The service will be started with the following configuration:")
	log.Info().Str("RUN_ADDRESS", conf.Address).
		Str("ACCRUAL_SYSTEM_ADDRESS", conf.AccrualAddress).
		Str("DATABASE_URI", conf.DatabaseDSN)
	// Инициализация канала Grace-ful Shutdown.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	// Инициализация хранилища сервиса.
	db, err := repository.NewDatabaseDSN(conf.DatabaseDSN)
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
		Addr:    conf.Address,
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
	blackbox := s.NewAccrl(db, conf.AccrualAddress)
	blackbox.Run()
}
