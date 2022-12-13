package service

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gtgaleevtimur/gofermart/internal/entity"
	"strconv"
	"strings"
	"time"
)

func (s *Service) Registration(acc *entity.Account) (*Session, error) {
	_, err := s.Users.Add(acc)
	if err != nil {
		return nil, err
	}
	session, err := s.Login(acc, "")
	if err != nil {
		return nil, err
	}
	return session, nil
}
func (s *Service) Login(acc *entity.Account, token string) (*Session, error) {
	user, err := s.Users.Get(acc.Login)
	if err != nil {
		return nil, err
	}
	flag := user.CheckPassword(acc.Password)
	if !flag {
		return nil, ErrInvalidPair
	}
	if token != "" {
		err = s.Sessions.Delete(token)
	}
	newToken := uuid.NewString()
	expires := time.Now().Add(600 * time.Second)
	session := &Session{
		UserID: user.ID,
		Token:  newToken,
		Expiry: expires,
	}
	err = s.Sessions.Add(session)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (s *Service) PostOrder(orderID, userID uint64) error {
	err := s.Orders.Add(orderID, userID)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) GetOrder(userID uint64) ([]*entity.Order, error) {
	orders, err := s.Orders.GetUserOrders(userID)
	if err != nil {
		return nil, err
	}

	layout := "2006-01-02T15:04:05-07:00"

	orderStSlice := make([]*entity.Order, 0)
	for _, o := range orders {
		orderSt := &entity.Order{
			Number:     fmt.Sprint(o.ID),
			Status:     strings.TrimSpace(o.Status),
			Accrual:    float64(o.Accrual) / 100,
			UploadedAt: o.UploadedAt.Format(layout),
		}

		orderStSlice = append(orderStSlice, orderSt)
	}

	return orderStSlice, nil
}

func (s *Service) GetBalance(userID uint64) (*entity.Balance, error) {
	balance, err := s.Balances.Get(userID)
	if err != nil {
		return nil, err
	}

	// храним баланс в копейках, отдаём в рублях: поэтому делим на 100
	b := &entity.Balance{
		Current:   float64(balance.Current) / 100,
		Withdrawn: float64(balance.Withdrawn) / 100,
	}

	return b, nil
}

func (s *Service) PostWithdraw(w *entity.Withdraw) error {
	orderID, err := strconv.Atoi(w.Order)
	if err != nil {
		return ErrOrderInvalidFormat
	}
	withdraw := &Withdraw{
		OrderID: uint64(orderID),
		UserID:  w.UserID,
		Sum:     uint64(w.Sum * 100),
	}
	err = s.Withdrawals.Add(withdraw)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetWithdrawals(userID uint64) ([]*entity.Withdraw, error) {
	wds, err := s.Withdrawals.GetWithdrawals(userID)
	if err != nil {
		return nil, err
	}

	wdsPr := make([]*entity.Withdraw, len(wds))
	for _, v := range wds {
		wpr := &entity.Withdraw{
			Order:       fmt.Sprint(v.OrderID),
			Sum:         float64(v.Sum) / 100,
			ProcessedAt: v.ProcessedAt.Format(time.RFC3339),
		}
		wdsPr = append(wdsPr, wpr)
	}

	return wdsPr, nil
}
