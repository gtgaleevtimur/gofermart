package entity

import (
	"golang.org/x/crypto/bcrypt"
	"sync"
	"time"
)

type AccountInfo struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UsersMemory struct {
	sync.RWMutex
	ByLogin map[string]*User
	ByID    map[uint64]*User
}

type User struct {
	ID       uint64
	Login    string
	Password []byte
}

func (u *User) CheckPassword(password string) bool {
	if err := bcrypt.CompareHashAndPassword(u.Password, []byte(password)); err != nil {
		return false
	}

	return true
}

func NewUsers() *UsersMemory {
	return &UsersMemory{
		ByLogin: make(map[string]*User),
		ByID:    make(map[uint64]*User),
	}
}

type Session struct {
	UserID uint64
	Token  string
	Expiry time.Time
}

func (s *Session) IsExpired() bool {
	return s.Expiry.Before(time.Now())
}

type SessionMemory struct {
	sync.RWMutex
	BySessionToken map[string]*Session
}

func NewSessions() *SessionMemory {
	return &SessionMemory{
		BySessionToken: make(map[string]*Session),
	}
}

type OrdersMemory struct {
	sync.RWMutex
	ByID map[uint64]*Order
}

func NewOrders() *OrdersMemory {
	return &OrdersMemory{
		ByID: make(map[uint64]*Order),
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
