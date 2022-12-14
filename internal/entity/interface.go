package entity

type Storager interface {
	AddWithdrawDB(withdraw *Withdraw) error
	GetWithdrawalsDB(userID uint64) ([]*Withdraw, error)

	AddUserDB(user *User) (uint64, error)
	GetUserDB(byKey interface{}) (*User, error)

	DeleteSessionDB(token string) error
	AddSessionDB(session *Session) error
	GetSessionDB(token string) (*Session, error)

	GetOrderDB(orderID uint64) (*Order, error)
	AddOrderDB(order *Order) error
	GetOrdersDB(id uint64) ([]*Order, error)

	GetBalanceDB(userID uint64) (*Balance, error)

	Register(accInfo *AccountInfo) (*Session, error)
	Login(accInfo *AccountInfo, oldToken string) (*Session, error)
	AddSession(session *Session) error
	GetSession(token string) (*Session, error)
	DeleteSession(token string) error
	GetUser(byKey interface{}) (*User, error)
	PostOrders(orderID, userID uint64) error
	AddOrders(orderID, userID uint64) error
	GetOrder(orderID uint64) (*Order, error)
	GetOrders(userID uint64) ([]*OrderX, error)
	GetBalance(userID uint64) (*BalanceX, error)
	PostWithdraw(wd *WithdrawX) error
	GetWithdrawals(userID uint64) ([]*WithdrawX, error)

	GetPullOrders(limit uint32) (map[uint64]*Order, error)
	UpdateOrder(order *Order) error
}
