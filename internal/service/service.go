package service

import "github.com/gtgaleevtimur/gofermart/internal/repository"

type Service struct {
	Storage     repository.Storager
	Users       *Users
	Sessions    *Sessions
	Orders      *Orders
	Balances    *Balances
	Withdrawals *Withdrawals
}

func NewService(s repository.Storager) *Service {
	service := &Service{
		Storage:  s,
		Users:    NewUsers(s),
		Sessions: NewSessions(s),
	}
	service.initPointer()
	return service
}

func (s *Service) initPointer() {
	s.Orders = NewOrders(s)
	s.Balances = NewBalance(s)
	s.Withdrawals = NewWithdrawals(s)
}
