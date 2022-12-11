package handler

import (
	"github.com/gtgaleevtimur/gofermart/internal/service"
	"github.com/rs/zerolog/log"
	"net/http"
)

func (h *Handler) Auth(w http.ResponseWriter, r *http.Request) (*service.Session, error) {
	// извлечём токен сессии
	cookie, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			log.Info().Err(service.ErrUnauthorizedAccess)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return nil, service.ErrUnauthorizedAccess
		}
		log.Info().Err(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, err
	}
	sessionToken := cookie.Value

	// получим сессию из хранилища по токену
	session, err := h.gophermart.Sessions.Get(sessionToken)
	if err != nil {
		log.Info().Err(err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return nil, err
	}

	// Удаляем сессию и выходим, если прошёл срок годности
	if session.IsExpired() {
		err = h.gophermart.Sessions.Delete(sessionToken)
		if err != nil {
			log.Info().Err(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil, err
		}
		log.Info().Err(service.ErrSessionExpired)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return nil, service.ErrSessionExpired
	}

	return session, nil
}
