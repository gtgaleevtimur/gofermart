package handler

import (
	"encoding/json"
	"errors"
	"github.com/gtgaleevtimur/gofermart/internal/entity"
	"github.com/gtgaleevtimur/gofermart/internal/service"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

func (h *Handler) PostWithdraw(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	session, err := h.Auth(w, r)
	if err != nil {
		return
	}
	user, err := h.gophermart.Users.Get(session.UserID)
	if err != nil {
		log.Info().Err(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Info().Msg("failed to read request body.")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	withdraw := &entity.Withdraw{}
	err = json.Unmarshal(body, &withdraw)
	if err != nil {
		log.Info().Msg(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	withdraw.UserID = user.ID
	err = h.gophermart.PostWithdraw(withdraw)
	if err != nil {
		// 402 — на счету недостаточно средств
		if errors.Is(err, service.ErrNotEnoughFunds) {
			log.Info().Msg(err.Error())
			http.Error(w, err.Error(), http.StatusPaymentRequired)
			return
		}

		// 422 — неверный формат номера заказа
		if errors.Is(err, service.ErrOrderInvalidFormat) {
			log.Info().Msg(err.Error())
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		// 500 — внутренняя ошибка сервера
		log.Info().Msg(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 200 — успешная обработка запроса
	w.WriteHeader(http.StatusOK)
	log.Info().Msg("new withdraw has been made for order")
}
