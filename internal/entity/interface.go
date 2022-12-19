package entity

// Storager - сборный интерфейс бога сервиса.
type Storager interface {
	Databaser
	Querer
	Controlluser
}

// Databaser - интерфейс, отвечающий за работу с БД.
type Databaser interface {
	GetBalanceDB(userID uint64) (Balance, error)
	GetOrderDB(orderID uint64) (Order, error)
	AddOrderDB(o *Order) error
	GetOrdersDB(id uint64) ([]Order, error)
	GetPullOrders(limit uint32) (map[uint64]Order, error)
	DeleteSessionDB(token string) error
	AddSessionDB(session *Session) error
	GetSessionDB(token string) (Session, error)
	AddUserDB(u *User) (uint64, error)
	GetUserDB(byKey interface{}) (User, error)
	AddWithdrawDB(withdraw *Withdraw) error
	GetWithdrawalsDB(userID uint64) ([]Withdraw, error)
}

// Querer - интерфейс, отвечающий за работу с blackbox.
type Querer interface {
	GetPullOrders(limit uint32) (map[uint64]Order, error)
	UpdateOrder(o Order) error
}

// Controlluser - интерфейс, отвечающий за методы контроллера хэндлера.
type Controlluser interface {
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
	GetWithdrawals(userID uint64) ([]WithdrawX, error)
}
