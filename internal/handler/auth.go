package handler

import (
	"fmt"
	"net/http"

	"github.com/gtgaleevtimur/gofermart/internal/entity"
	"github.com/gtgaleevtimur/gofermart/internal/repository"
)

// auth - обработчик, авторизирующий пользовтаеля и его сессию.
func (c *Controller) auth(w http.ResponseWriter, r *http.Request) (*entity.Session, error) {
	st, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			c.error(w, r, repository.ErrUnauthorizedAccess, http.StatusUnauthorized)
			return nil, repository.ErrUnauthorizedAccess
		}
		c.error(w, r, err, http.StatusBadRequest)
		return nil, err
	}
	sessionToken := st.Value
	session, err := c.Storage.GetSession(sessionToken)
	if err != nil {
		err = fmt.Errorf("session token is not present")
		c.error(w, r, err, http.StatusUnauthorized)
		return nil, err
	}
	if session.IsExpired() {
		c.Storage.DeleteSession(sessionToken)
		err = fmt.Errorf("session has expired")
		c.error(w, r, err, http.StatusUnauthorized)
		return nil, err
	}
	return session, nil
}
