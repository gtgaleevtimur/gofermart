package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/gtgaleevtimur/gofermart/internal/entity"
	"github.com/gtgaleevtimur/gofermart/internal/service"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
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
	session, err := h.gophermart.Registration(&acc)
	if err != nil {
		if errors.Is(err, service.ErrLoginAlreadyTaken) {
			log.Info().Msg("registration user failed")
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		log.Info().Msg("registration user failed")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if session == nil {
		log.Info().Msg("nil session received")
		http.Error(w, "nil session received", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   session.Token,
		Expires: session.Expiry,
	})
	log.Info().Msg("user successfully registered and authenticated")
	w.WriteHeader(http.StatusOK)
}
