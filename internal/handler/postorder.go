package handler

import (
	"errors"
	"fmt"
	"github.com/gtgaleevtimur/gofermart/internal/service"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"strconv"
)

func (h *Handler) PostOrder(w http.ResponseWriter, r *http.Request) {
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
	orderID, err := strconv.Atoi(string(body))
	if err != nil {
		log.Info().Err(service.ErrOrderInvalidFormat)
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	err = h.gophermart.PostOrder(uint64(orderID), user.ID)
	if err != nil {
		if errors.Is(err, service.ErrOrderAlreadyLoadedByUser) {
			w.WriteHeader(http.StatusOK)
			msg := fmt.Sprintf("order %d has already been uploaded by this user", orderID)
			log.Info().Msg(msg)
			return
		}
		if errors.Is(err, service.ErrOrderAlreadyLoadedByAnotherUser) {
			msg := fmt.Sprintf("order %d has already been uploaded by another user", orderID)
			log.Info().Msg(msg)
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		if errors.Is(err, service.ErrOrderInvalidFormat) {
			log.Info().Msg(err.Error())
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		log.Info().Err(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	msg := fmt.Sprintf("new order %d has been accepted", orderID)
	log.Info().Msg(msg)
}
