package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// GetBalance - обработчик, обрабатывающий запрос на проверку баланса пользователя в системе.
func (c *Controller) GetBalance(w http.ResponseWriter, r *http.Request) {
	st, err := c.auth(w, r)
	if err != nil {
		return
	}
	u, err := c.Storage.GetUser(st.UserID)
	if err != nil {
		c.error(w, r, fmt.Errorf("failed to get user by ID - %s", err.Error()), http.StatusInternalServerError)
		return
	}
	balanceProxy, err := c.Storage.GetBalance(u.ID)
	if err != nil {
		c.error(w, r, fmt.Errorf("failed to get balance for user `%s` - %s", u.Login, err.Error()), http.StatusInternalServerError)
		return
	}
	body, err := json.Marshal(&balanceProxy)
	if err != nil {
		c.error(w, r, fmt.Errorf("failed to marshal JSON - %s", err.Error()), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", ContentTypeApplicationJSON)
	w.Write(body)
}
