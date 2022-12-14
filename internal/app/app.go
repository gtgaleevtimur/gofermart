package app

import (
	"context"
	"github.com/gtgaleevtimur/gofermart/internal/config"
	"github.com/gtgaleevtimur/gofermart/internal/handler"
	r "github.com/gtgaleevtimur/gofermart/internal/repository"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run() {
	conf := config.NewConfig()
	log.Debug().Str("RUN_ADDRESS", conf.Address).
		Str("DATABASE_URI", conf.DatabaseURI).
		Str("ACCRUAL_SYSTEM_ADDRESS", conf.AccrualSystemAddress).
		Msg("Receive config")
	repository, err := r.NewRepository(conf.DatabaseURI)
	if err != nil {
		log.Fatal().Err(err).Msg("Repository initialization failed")
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	server := &http.Server{
		Addr:    conf.Address,
		Handler: handler.NewRouter(repository),
	}
	go func() {
		shutdownCtx, shutdownCtxCancel := context.WithTimeout(context.Background(), time.Second*20)
		defer shutdownCtxCancel()
		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal().Msg("graceful shutdown timed out and forcing exit.")
			}
		}()
		err = server.Shutdown(context.Background())
		if err != nil {
			log.Fatal().Err(err).Msg("server shutdown error")
		}
	}()
	go func() {
		log.Info().Str("starting server at", server.Addr)
		err = server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("failed to run server")
		}
	}()
	blackbox := r.NewBlackbox(repository, conf.AccrualSystemAddress)
	blackbox.Start()
}
