package entity

import (
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AccountInfo struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UsersMemory struct {
	sync.RWMutex
	ByLogin map[string]User
	ByID    map[uint64]User
}

type User struct {
	ID       uint64
	Login    string
	Password []byte
}

// CheckPassword - функция, проверяющая пароль на соответствие.
func (u *User) CheckPassword(password string) bool {
	if err := bcrypt.CompareHashAndPassword(u.Password, []byte(password)); err != nil {
		return false
	}

	return true
}

// NewUsers - конструктор,хэш-таблицы пользователей.
func NewUsers() *UsersMemory {
	return &UsersMemory{
		ByLogin: make(map[string]User),
		ByID:    make(map[uint64]User),
	}
}

type Session struct {
	UserID uint64
	Token  string
	Expiry time.Time
}

// IsExpired - метод, проверяющий срок годности cookie.
func (s *Session) IsExpired() bool {
	return s.Expiry.Before(time.Now())
}

type SessionMemory struct {
	sync.RWMutex
	BySessionToken map[string]Session
}

// NewSessions - конструктор хэш-таблицы сессии пользователей.
func NewSessions() *SessionMemory {
	return &SessionMemory{
		BySessionToken: make(map[string]Session),
	}
}

type OrdersMemory struct {
	sync.RWMutex
	ByID map[uint64]Order
}

// NewOrders - конструктор хэш-таблицы заказов пользователей.
func NewOrders() *OrdersMemory {
	return &OrdersMemory{
		ByID: make(map[uint64]Order),
	}
}

type Order struct {
	ID         uint64
	UserID     uint64
	Status     string
	Accrual    uint64
	UploadedAt time.Time
}

type OrderX struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual,omitempty"`
	UploadedAt string  `json:"uploaded_at"`
}

type Balance struct {
	UserID    uint64
	Current   uint64
	Withdrawn uint64
}

type BalanceMemory struct {
	sync.RWMutex
	ByUserID map[uint64]Balance
}

// NewBalance - конструктор хэш-таблицы балансов пользователей.
func NewBalance() *BalanceMemory {
	return &BalanceMemory{
		ByUserID: make(map[uint64]Balance),
	}
}

type BalanceX struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type WithdrawX struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	UserID      uint64  `json:"-"`
	ProcessedAt string  `json:"processed_at"`
}

type Withdraw struct {
	OrderID     uint64
	UserID      uint64
	Sum         uint64
	ProcessedAt time.Time
}
