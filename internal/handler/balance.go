package handler

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
	"net/http"
)

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
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
	balance, err := h.gophermart.GetBalance(user.ID)
	if err != nil {
		log.Info().Msg(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req, err := json.Marshal(&balance)
	if err != nil {
		log.Info().Msg(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(req)
}
