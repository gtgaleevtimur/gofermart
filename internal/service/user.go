package service

import (
	"github.com/gtgaleevtimur/gofermart/internal/entity"
	"github.com/gtgaleevtimur/gofermart/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"sync"
)

type User struct {
	ID       uint64
	Login    string
	Password []byte
}

type Users struct {
	sync.RWMutex
	storage repository.Storager
	byLog   map[string]*User
	byID    map[uint64]*User
}

func NewUsers(st repository.Storager) *Users {
	return &Users{
		storage: st,
		byLog:   make(map[string]*User),
		byID:    make(map[uint64]*User),
	}
}
func (u *Users) Add(acc *entity.Account) (uint64, error) {
	u.RLock()
	_, ok := u.byLog[acc.Login]
	u.RUnlock()
	if ok {
		return 0, ErrLoginAlreadyTaken
	}
	hash, err := hashedPassword(acc.Password)
	if err != nil {
		return 0, err
	}
	user := &User{Login: acc.Login, Password: hash}
	id, err := u.storage.AddUser(user)
	if err != nil {
		return 0, err
	}
	user.ID = id
	u.Lock()
	u.byLog[user.Login] = user
	u.byID[user.ID] = user
	u.Unlock()
	return user.ID, nil
}

func (u *Users) Get(key interface{}) (*User, error) {
	var user *User
	var ok bool
	var err error
	u.RLock()
	switch k := key.(type) {
	case string:
		user, ok = u.byLog[k]
	case uint64:
		user, ok = u.byID[k]
	default:
		u.RUnlock()
		return nil, ErrTypeNotAllowed
	}
	u.RUnlock()
	if !ok {
		user, err = u.storage.GetUser(key)
		if err != nil {
			return nil, err
		}
		u.Lock()
		u.byLog[user.Login] = user
		u.byID[user.ID] = user
		u.Unlock()
	}
	return user, nil
}
func (u *User) CheckPassword(password string) bool {
	if err := bcrypt.CompareHashAndPassword(u.Password, []byte(password)); err != nil {
		return false
	}
	return true
}

func hashedPassword(password string) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		return nil, err
	}
	return hash, nil
}
