package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gtgaleevtimur/gofermart/internal/entity"
	"github.com/gtgaleevtimur/gofermart/internal/repository"
)

// Login - обработчик аутентификации пользователя.
func (c *Controller) Login(w http.ResponseWriter, r *http.Request) {
	var creds *entity.AccountInfo
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		c.error(w, r, err, http.StatusBadRequest)
		return
	}
	var sessionToken string
	st, err := r.Cookie("session_token")
	if err == nil {
		sessionToken = st.Value
	}
	session, err := c.Storage.Login(creds, sessionToken)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidPair) || errors.Is(err, repository.ErrUserNotFound) {
			c.error(w, r, repository.ErrInvalidPair, http.StatusUnauthorized)
			return
		}
		c.error(w, r, err, http.StatusInternalServerError)
		return
	}
	if session == nil {
		c.error(w, r, fmt.Errorf("got nil session"), http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   session.Token,
		Expires: session.Expiry,
	})
	msg := fmt.Sprintf("session for user `%s` successfully created", creds.Login)
	c.log(r, msg)
}
