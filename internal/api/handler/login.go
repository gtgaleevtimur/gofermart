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

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Info().Msg("failed to read request body.")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var acc entity.Account
	err = json.Unmarshal(body, &acc)
	if err != nil {
		log.Info().Msg("failed to unmarshal request body.")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var token string
	c, err := r.Cookie("session_token")
	if err == nil {
		token = c.Value
	}
	session, err := h.gophermart.Login(&acc, token)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPair) || errors.Is(err, service.ErrUserNotFound) {
			log.Info().Err(err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		log.Info().Err(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if session == nil {
		log.Info().Msg("nil session received")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   session.Token,
		Expires: session.Expiry,
	})
	log.Info().Msg("user successfully authenticated")
	w.WriteHeader(http.StatusOK)
}
