package service

type Storager interface {
	Bootstrap() (err error)
	Connect(databaseDSN string) (err error)

	GetBalance(userID uint64) (*Balance, error)

	GetOrder(orderID uint64) (*Order, error)
	AddOrder(o *Order) error
	GetUserOrders(id uint64) ([]*Order, error)

	DeleteSession(token string) error
	AddSession(session *Session) error
	GetSession(token string) (*Session, error)

	AddUser(user *User) (uint64, error)
	GetUser(key interface{}) (*User, error)

	AddWithdraw(withdraw *Withdraw) error
	GetUserWithdrawals(userID uint64) ([]*Withdraw, error)
}
