package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// GetOrders - обработчик запроса на сделанный заказы пользователем.
func (c *Controller) GetOrders(w http.ResponseWriter, r *http.Request) {
	st, err := c.auth(w, r)
	if err != nil {
		return
	}
	u, err := c.Storage.GetUser(st.UserID)
	if err != nil {
		c.error(w, r, fmt.Errorf("failed to get user by ID - %s", err.Error()), http.StatusInternalServerError)
		return
	}
	userID := u.ID
	ordersX, err := c.Storage.GetOrders(userID)
	if err != nil {
		c.error(w, r, fmt.Errorf("failed to get all orders - %s", err.Error()), http.StatusInternalServerError)
		return
	}
	if len(ordersX) == 0 {
		c.error(w, r, fmt.Errorf("orders not found for this user"), http.StatusNoContent)
		return
	}
	body, err := json.Marshal(&ordersX)
	if err != nil {
		c.error(w, r, fmt.Errorf("failed to marshal JSON - %w", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", ContentTypeApplicationJSON)
	w.Write(body)
}
