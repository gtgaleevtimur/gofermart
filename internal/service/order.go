package service

import (
	"github.com/gtgaleevtimur/gofermart/internal/loon"
	"strconv"
	"sync"
	"time"
)

type Order struct {
	ID         uint64
	UserID     uint64
	Status     string
	Accrual    uint64
	UploadedAt time.Time
}

type Orders struct {
	pointer *Service
	byID    map[uint64]*Order
	sync.RWMutex
}

func NewOrders(pointer *Service) *Orders {
	return &Orders{
		pointer: pointer,
		byID:    make(map[uint64]*Order),
	}
}

func (o *Orders) Add(orderID, userID uint64) error {
	// Проверим номер заказа на соответствие алгоритму Луна
	strOrderID := strconv.Itoa(int(orderID))
	if !loon.IsValid(strOrderID) {
		return ErrOrderInvalidFormat
	}

	// проверим наличие заказа
	order, _ := o.Get(orderID)
	if order != nil {
		if order.UserID == userID {
			return ErrOrderAlreadyLoadedByUser
		}
		return ErrOrderAlreadyLoadedByAnotherUser
	}

	order = &Order{
		ID:         orderID,
		UserID:     userID,
		Status:     "NEW",
		UploadedAt: time.Now(),
	}
	err := o.pointer.Storage.AddOrder(order)
	if err != nil {
		return err
	}

	// закэшируем полученный заказ
	o.Lock()
	o.byID[orderID] = order
	o.Unlock()

	return nil
}

func (o *Orders) Get(orderID uint64) (*Order, error) {
	var err error

	o.RLock()
	order, ok := o.byID[orderID]
	o.RUnlock()
	if !ok {
		order, err = o.pointer.Storage.GetOrder(orderID)
		if err != nil {
			return nil, err
		}

		// закэшируем полученный заказ
		o.Lock()
		o.byID[orderID] = order
		o.Unlock()
	}

	return order, nil
}

func (o *Orders) GetUserOrders(userID uint64) ([]*Order, error) {
	ors, err := o.pointer.Storage.GetUserOrders(userID)
	if err != nil {
		return nil, err
	}

	return ors, nil
}
