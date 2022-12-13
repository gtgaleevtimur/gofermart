package service

import (
	"github.com/gtgaleevtimur/gofermart/internal/loon"
	"strconv"
	"time"
)

type Withdraw struct {
	OrderID     uint64
	UserID      uint64
	Sum         uint64
	ProcessedAt time.Time
}

type Withdrawals struct {
	pointer *Service
}

func NewWithdrawals(pointer *Service) *Withdrawals {
	return &Withdrawals{
		pointer: pointer,
	}
}

func (w *Withdrawals) Add(withdraw *Withdraw) error {
	// Проверим номер заказа на соответствие алгоритму Луна
	orderID := strconv.Itoa(int(withdraw.OrderID))
	if !loon.IsValid(orderID) {
		return ErrOrderInvalidFormat
	}

	err := w.pointer.Storage.AddWithdraw(withdraw)
	if err != nil {
		return err
	}

	// баланс изменился, удалим запись из кэша баланса
	w.pointer.Balances.Lock()
	delete(w.pointer.Balances.byUserID, withdraw.UserID)
	w.pointer.Balances.Unlock()

	return nil
}

func (w *Withdrawals) GetWithdrawals(userID uint64) ([]*Withdraw, error) {
	withdrawals, err := w.pointer.Storage.GetUserWithdrawals(userID)
	if err != nil {
		return nil, err
	}

	if len(withdrawals) == 0 {
		return nil, ErrNoContent
	}

	return withdrawals, nil
}
