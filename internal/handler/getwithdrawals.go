package handler

import (
	"encoding/json"
	"errors"
	"github.com/gtgaleevtimur/gofermart/internal/service"
	"github.com/rs/zerolog/log"
	"net/http"
)

func (h *Handler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
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
	witdrawals, err := h.gophermart.GetWithdrawals(user.ID)
	if err != nil {
		// 204 — нет ни одного списания
		if errors.Is(err, service.ErrNoContent) {
			log.Info().Msg(err.Error())
			http.Error(w, err.Error(), http.StatusNoContent)
			return
		}

		// 500 — внутренняя ошибка сервера
		log.Info().Msg(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(&witdrawals)
	if err != nil {
		log.Info().Msg(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}
