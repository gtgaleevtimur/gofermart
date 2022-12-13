package service

import (
	"sync"
	"time"
)

type Session struct {
	UserID uint64
	Token  string
	Expiry time.Time
}

type Sessions struct {
	sync.RWMutex
	storage        Storager
	bySessionToken map[string]*Session
}

func NewSessions(st Storager) *Sessions {
	return &Sessions{
		storage:        st,
		bySessionToken: make(map[string]*Session),
	}
}

func (s *Sessions) Delete(token string) error {
	s.Lock()
	delete(s.bySessionToken, token)
	s.Unlock()

	err := s.storage.DeleteSession(token)
	if err != nil {
		return err
	}
	return nil
}

func (s *Sessions) Add(session *Session) error {
	s.RLock()
	_, ok := s.bySessionToken[session.Token]
	s.RUnlock()
	if ok {
		return ErrSessionExistAlready
	}
	err := s.storage.AddSession(session)
	if err != nil {
		return err
	}
	s.Lock()
	s.bySessionToken[session.Token] = session
	s.Unlock()
	return nil
}

func (s *Sessions) Get(token string) (*Session, error) {
	var err error

	s.RLock()
	session, ok := s.bySessionToken[token]
	s.RUnlock()
	if !ok {
		session, err = s.storage.GetSession(token)
		if err != nil {
			return nil, ErrSessionGetTransaction
		}
		// закэшируем полученную сессию
		s.Lock()
		s.bySessionToken[session.Token] = session
		s.Unlock()
	}

	return session, nil
}

func (s *Session) IsExpired() bool {
	return s.Expiry.Before(time.Now())
}
