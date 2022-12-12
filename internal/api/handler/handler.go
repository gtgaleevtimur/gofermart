package handler

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gtgaleevtimur/gofermart/internal/service"
)

type Handler struct {
	router     chi.Router
	gophermart *service.Service
}

func NewHandler(g *service.Service) *Handler {
	handler := &Handler{
		router:     chi.NewRouter(),
		gophermart: g,
	}
	handler.router.Use(middleware.Compress(3, "gzip"))
	handler.router.Use(middleware.RequestID)
	handler.router.Use(middleware.RealIP)
	handler.router.Use(middleware.Logger)
	handler.router.Use(middleware.Recoverer)

	handler.router.Route("/api/user", func(r chi.Router) {
		r.Post("/register", handler.Register)
		r.Post("/login", handler.Login)
		r.Post("/orders", handler.PostOrder)
		r.Get("/orders", handler.GetOrder)
		r.Get("/balance", handler.GetBalance)
		r.Post("/balance/withdraw", handler.PostWithdraw)
		r.Get("/balance/withdrawals", handler.GetWithdrawals)
	})

	return handler
}

func (h *Handler) R() chi.Router {
	return h.router
}
