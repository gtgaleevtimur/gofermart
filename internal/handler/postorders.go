package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gtgaleevtimur/gofermart/internal/repository"
)

// PostOrders - обработчик загрузки пользователем номера заказа для расчета баллов лояльности.
func (c *Controller) PostOrders(w http.ResponseWriter, r *http.Request) {
	var err error
	ct := r.Header.Get("Content-Type")
	if ct != ContentTypeTextPlain {
		err = fmt.Errorf("wrong content type, %s needed", ContentTypeTextPlain)
		c.error(w, r, err, http.StatusBadRequest)
		return
	}
	st, err := c.auth(w, r)
	if err != nil {
		return
	}
	u, err := c.Storage.GetUser(st.UserID)
	if err != nil {
		c.error(w, r, fmt.Errorf("failed to get user by ID - %s", err.Error()), http.StatusInternalServerError)
		return
	}
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		c.error(w, r, fmt.Errorf("failed to read request body - %s", err.Error()), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	orderID, err := strconv.Atoi(string(reqBody))
	if err != nil {
		c.error(w, r, fmt.Errorf("%s - %s", repository.ErrOrderInvalidFormat, err.Error()), http.StatusUnprocessableEntity)
		return
	}
	err = c.Storage.PostOrders(uint64(orderID), u.ID)
	if err != nil {
		if errors.Is(err, repository.ErrOrderAlreadyLoadedByUser) {
			w.WriteHeader(http.StatusOK)
			msg := fmt.Sprintf("order %d has already been uploaded by this user", orderID)
			c.log(r, msg)
			return
		}
		if errors.Is(err, repository.ErrOrderAlreadyLoadedByAnotherUser) {
			c.error(w, r, repository.ErrOrderAlreadyLoadedByAnotherUser, http.StatusConflict)
			return
		}
		if errors.Is(err, repository.ErrOrderInvalidFormat) {
			c.error(w, r, repository.ErrOrderInvalidFormat, http.StatusUnprocessableEntity)
			return
		}
		c.error(w, r, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	msg := fmt.Sprintf("new order %d has been accepted for processing", orderID)
	c.log(r, msg)
}
