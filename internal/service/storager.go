package repository

import (
	"github.com/gtgaleevtimur/gofermart/internal/service"
)

type Storager interface {
	Bootstrap() (err error)
	Connect(databaseDSN string) (err error)

	GetBalance(userID uint64) (*service.Balance, error)

	GetOrder(orderID uint64) (*service.Order, error)
	AddOrder(o *service.Order) error
	GetUserOrders(id uint64) ([]*service.Order, error)

	DeleteSession(token string) error
	AddSession(session *service.Session) error
	GetSession(token string) (*service.Session, error)

	AddUser(user *service.User) (uint64, error)
	GetUser(key interface{}) (*service.User, error)

	AddWithdraw(withdraw *service.Withdraw) error
	GetUserWithdrawals(userID uint64) ([]*service.Withdraw, error)
}
