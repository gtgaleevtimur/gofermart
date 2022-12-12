package handler

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
)

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
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
	userID := user.ID
	orders, err := h.gophermart.GetOrder(userID)
	if err != nil {
		log.Info().Msg(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		log.Info().Msg("not found orders for user")
		http.Error(w, "not found orders for user", http.StatusNoContent)
		return
	}
	body, err := json.Marshal(&orders)
	if err != nil {
		msg := fmt.Sprintf("failed to marshal JSON - %s", err.Error())
		log.Info().Msg(msg)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}
