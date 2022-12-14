package repository

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gtgaleevtimur/gofermart/internal/entity"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"time"
)

func (r *Repository) Register(accInfo *entity.AccountInfo) (*entity.Session, error) {
	r.userMemory.RLock()
	_, ok := r.userMemory.ByLogin[accInfo.Login]
	r.userMemory.RUnlock()
	if ok {
		return nil, ErrLoginAlreadyTaken
	}
	hashedPassword, err := HashPass(accInfo.Password)
	if err != nil {
		return nil, err
	}
	u := &entity.User{
		Login:    accInfo.Login,
		Password: hashedPassword,
	}

	id, err := r.AddUserDB(u)
	if err != nil {
		return nil, err
	}
	u.ID = id
	r.userMemory.Lock()
	r.userMemory.ByLogin[u.Login] = u
	r.userMemory.ByID[u.ID] = u
	r.userMemory.Unlock()

	session, err := r.Login(accInfo, "")
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (r *Repository) Login(accInfo *entity.AccountInfo, oldToken string) (*entity.Session, error) {
	user, err := r.GetUser(accInfo.Login)
	if err != nil {
		return nil, err
	}
	check := user.CheckPassword(accInfo.Password)
	if !check {
		return nil, ErrInvalidPair
	}

	if oldToken != "" {
		err = r.DeleteSession(oldToken)
		if err != nil {
			log.Error().Err(err)
		}
	}

	newToken := uuid.NewString()
	expiresAt := time.Now().Add(600 * time.Second)

	s := &entity.Session{
		UserID: user.ID,
		Token:  newToken,
		Expiry: expiresAt,
	}
	err = r.AddSession(s)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (r *Repository) AddSession(session *entity.Session) error {
	r.sessionMemory.RLock()
	_, ok := r.sessionMemory.BySessionToken[session.Token]
	r.sessionMemory.RUnlock()
	if ok {
		return fmt.Errorf("session already exists")
	}

	err := r.AddSessionDB(session)
	if err != nil {
		return err
	}

	r.sessionMemory.Lock()
	r.sessionMemory.BySessionToken[session.Token] = session
	r.sessionMemory.Unlock()

	return nil
}

func (r *Repository) DeleteSession(token string) error {
	r.sessionMemory.Lock()
	delete(r.sessionMemory.BySessionToken, token)
	r.sessionMemory.Unlock()

	err := r.DeleteSessionDB(token)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetUser(byKey interface{}) (*entity.User, error) {
	var err error
	var u *entity.User
	var ok bool

	r.userMemory.RLock()
	switch key := byKey.(type) {
	case string:
		u, ok = r.userMemory.ByLogin[key]
	case uint64:
		u, ok = r.userMemory.ByID[key]
	default:
		r.userMemory.RUnlock()
		return nil, fmt.Errorf("given type not implemented")
	}
	r.userMemory.RUnlock()

	if !ok {
		u, err = r.GetUserDB(byKey)
		if err != nil {
			return nil, err
		}
		// закэшируем полученного пользователя
		r.userMemory.Lock()
		r.userMemory.ByLogin[u.Login] = u
		r.userMemory.ByID[u.ID] = u
		r.userMemory.Unlock()
	}

	return u, nil
}

func HashPass(password string) ([]byte, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		return nil, err
	}

	return hashedPassword, nil
}