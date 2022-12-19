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

// PostWithdraw - обработчик запроса на списание баллов с накопительного счета пользователя в счет оплаты нового заказа.
func (c *Controller) PostWithdraw(w http.ResponseWriter, r *http.Request) {
	var err error
	ct := r.Header.Get("Content-Type")
	if ct != ContentTypeApplicationJSON {
		err = fmt.Errorf("wrong content type, %s needed", ContentTypeApplicationJSON)
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
	wd := &entity.WithdrawX{}
	err = json.Unmarshal(reqBody, &wd)
	if err != nil {
		c.error(w, r, fmt.Errorf("failed to unmarshal body - %s", err.Error()), http.StatusBadRequest)
		return
	}
	wd.UserID = u.ID
	err = c.Storage.PostWithdraw(wd)
	if err != nil {
		if errors.Is(err, repository.ErrNotEnoughFunds) {
			c.error(w, r, repository.ErrNotEnoughFunds, http.StatusPaymentRequired)
			return
		}
		if errors.Is(err, repository.ErrOrderInvalidFormat) {
			c.error(w, r, repository.ErrOrderInvalidFormat, http.StatusUnprocessableEntity)
			return
		}
		c.error(w, r, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	msg := fmt.Sprintf("new withdraw has been made for order ID %s", wd.Order)
	c.log(r, msg)
}
