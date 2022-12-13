package service

import "sync"

type Balance struct {
	UserID    uint64
	Current   uint64
	Withdrawn uint64
}

type Balances struct {
	pointer  *Service
	byUserID map[uint64]*Balance
	sync.RWMutex
}

func NewBalance(pointer *Service) *Balances {
	return &Balances{
		pointer:  pointer,
		byUserID: make(map[uint64]*Balance),
	}
}

func (b *Balances) Get(userID uint64) (*Balance, error) {
	b.RLock()
	balance, ok := b.byUserID[userID]
	b.RUnlock()
	if !ok {
		bal, err := b.pointer.Storage.GetBalance(userID)
		if err != nil {
			return nil, err
		}
		b.Lock()
		b.byUserID[userID] = bal
		b.Unlock()
	}

	return balance, nil
}
