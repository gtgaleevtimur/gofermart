package server

import (
	"context"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gtgaleevtimur/gofermart/internal/config"
)

type Server struct {
	server *http.Server
}

func NewServer(h http.Handler, c *config.Config) *Server {
	s := &Server{
		server: &http.Server{
			Addr:    c.Address,
			Handler: h,
		}}
	return s
}

func (s *Server) ListAndServ() {
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-sig

		// определяем время для штатного завершения работы сервера
		// необходимо на случай вечного ожидания закрытия всех подключений к серверу
		shutdownCtx, shutdownCtxCancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer shutdownCtxCancel()
		// принудительно завершаем работу по истечении срока s.graceTimeout
		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Info().Msg("Graceful shutdown timed out! Forcing exit.")
				log.Fatal()
			}
		}()

		err := s.server.Shutdown(context.Background())
		if err != nil {
			log.Info().Msg("Server shutdown error.")
			log.Fatal().Err(err)
		}
	}()
	log.Info().Msg("Service started.")
	err := s.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Info().Msg("Failed to run HTTP-server")
		log.Fatal().Err(err)
	}
}
