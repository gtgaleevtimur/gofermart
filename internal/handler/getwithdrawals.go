package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gtgaleevtimur/gofermart/internal/repository"
)

// GetWithdrawals - обработчик обрабатывающий запрос на  получение информации о выводе средств с накопительного счёта пользователем.
func (c *Controller) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	var err error
	st, err := c.auth(w, r)
	if err != nil {
		return
	}
	u, err := c.Storage.GetUser(st.UserID)
	if err != nil {
		c.error(w, r, fmt.Errorf("failed to get user by ID - %s", err.Error()), http.StatusInternalServerError)
		return
	}
	wdx, err := c.Storage.GetWithdrawals(u.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNoContent) {
			c.error(w, r, repository.ErrNoContent, http.StatusNoContent)
			return
		}
		c.error(w, r, err, http.StatusInternalServerError)
		return
	}
	body, err := json.Marshal(&wdx)
	if err != nil {
		c.error(w, r, fmt.Errorf("failed to marshal JSON - %s", err.Error()), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", ContentTypeApplicationJSON)
	w.Write(body)
}
