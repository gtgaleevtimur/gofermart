package handler

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/gtgaleevtimur/gofermart/internal/entity"
)

const (
	ContentTypeApplicationJSON = "application/json"
	ContentTypeTextPlain       = "text/plain"
)

// NewRouter - функция инициализирующая и настраивающая роутер сервиса.
func NewRouter(r entity.Storager) chi.Router {
	router := chi.NewRouter()
	controller := newController(r)
	router.Use(middleware.Compress(3, "gzip"))
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Route("/api/user", func(rout chi.Router) {
		rout.Post("/register", controller.Register)
		rout.Post("/login", controller.Login)

		rout.Post("/orders", controller.PostOrders)
		rout.Get("/orders", controller.GetOrders)

		rout.Get("/balance", controller.GetBalance)

		rout.Post("/balance/withdraw", controller.PostWithdraw)
		rout.Get("/withdrawals", controller.GetWithdrawals)
	})

	router.NotFound(NotFound())
	router.MethodNotAllowed(NotAllowed())

	return router
}

type Controller struct {
	Storage entity.Storager
}

// newController - функция-конструктор контролера хэндлера.
func newController(s entity.Storager) *Controller {
	return &Controller{Storage: s}
}

// NotFound - обработчик неподдерживаемых маршрутов.
func NotFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte("route does not exist"))
	}
}

// NotAllowed - обработчик неподдерживаемых методов.
func NotAllowed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte("method does not allowed"))
	}
}
