package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gtgaleevtimur/gofermart/internal/entity"
	"github.com/gtgaleevtimur/gofermart/internal/repository"
)

// Register - обработчик регистрации нового пользователя.
func (c *Controller) Register(w http.ResponseWriter, r *http.Request) {
	content := r.Header.Get("Content-Type")
	if content != "application/json" {
		err := fmt.Errorf("wrong content type, JSON needed")
		c.error(w, r, err, http.StatusBadRequest)
		return
	}
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		c.error(w, r, fmt.Errorf("failed to read request body - %s", err.Error()), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	var accInfo entity.AccountInfo
	err = json.Unmarshal(reqBody, &accInfo)
	if err != nil {
		c.error(w, r, fmt.Errorf("failed to unmarshal body - %s", err.Error()), http.StatusBadRequest)
		return
	}
	session, err := c.Storage.Register(&accInfo)
	if err != nil {
		msg := "failed to register new user"
		if errors.Is(err, repository.ErrLoginAlreadyTaken) {
			c.error(w, r, fmt.Errorf("%s - %s", msg, err.Error()), http.StatusConflict)
			return
		}
		c.error(w, r, fmt.Errorf("%s - %s", msg, err.Error()), http.StatusInternalServerError)
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
	msg := fmt.Sprintf("session for user `%s` successfully created", accInfo.Login)
	c.log(r, msg)
}
