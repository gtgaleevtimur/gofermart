package repository

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"

	"github.com/gtgaleevtimur/gofermart/internal/entity"
	"github.com/gtgaleevtimur/gofermart/internal/loon"
)

// Register - общий метод ля регистрации пользователя.
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
	r.userMemory.ByLogin[u.Login] = *u
	r.userMemory.ByID[u.ID] = *u
	r.userMemory.Unlock()
	session, err := r.Login(accInfo, "")
	if err != nil {
		return nil, err
	}
	return session, nil
}

// Login - метод, обновляющий сессию при авторизации пользователя и добавляющий при регистрации.
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

// AddSession - метод добавляющий пользователя в хэш-таблицу и БД.
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
	r.sessionMemory.BySessionToken[session.Token] = *session
	r.sessionMemory.Unlock()
	return nil
}

// GetSession - метод, возвращающий сессию пользователя из хэш-памяти или БД.
func (r *Repository) GetSession(token string) (*entity.Session, error) {
	var err error
	r.sessionMemory.RLock()
	session, ok := r.sessionMemory.BySessionToken[token]
	r.sessionMemory.RUnlock()
	if !ok {
		session, err = r.GetSessionDB(token)
		if err != nil {
			return nil, fmt.Errorf("token session not found - %s", err.Error())
		}
		r.sessionMemory.Lock()
		r.sessionMemory.BySessionToken[session.Token] = session
		r.sessionMemory.Unlock()
	}
	return &session, nil
}

// DeleteSession - метод, удаляющий сессию пользователя из хэш-таблицы и БД.
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

// GetUser - метод, возвращающий информацию о пользователе из хэш-таблицы или БД.
func (r *Repository) GetUser(byKey interface{}) (*entity.User, error) {
	var err error
	var u entity.User
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
	return &u, nil
}

// PostOrders - метод, регистрирующий заказ пользователя в хэш-таблице или БД.
func (r *Repository) PostOrders(orderID, userID uint64) error {
	err := r.AddOrders(orderID, userID)
	if err != nil {
		return err
	}

	return nil
}

// AddOrders - хэлпер метода PostOrders.
func (r *Repository) AddOrders(orderID, userID uint64) error {
	strOrderID := strconv.Itoa(int(orderID))
	if !loon.IsValid(strOrderID) {
		return ErrOrderInvalidFormat
	}
	order, _ := r.GetOrder(orderID)
	if order != nil {
		if order.UserID == userID {
			return ErrOrderAlreadyLoadedByUser
		}
		return ErrOrderAlreadyLoadedByAnotherUser
	}
	order = &entity.Order{
		ID:         orderID,
		UserID:     userID,
		Status:     "NEW",
		UploadedAt: time.Now(),
	}
	err := r.AddOrderDB(order)
	if err != nil {
		return err
	}
	r.ordersMemory.Lock()
	r.ordersMemory.ByID[orderID] = *order
	r.ordersMemory.Unlock()
	return nil
}

// GetOrder - метод, возвращающий информацию о заказе по его номеру из хэш-таблицы или БД.
func (r *Repository) GetOrder(orderID uint64) (*entity.Order, error) {
	var err error
	r.ordersMemory.RLock()
	o, ok := r.ordersMemory.ByID[orderID]
	r.ordersMemory.RUnlock()
	if !ok {
		o, err = r.GetOrderDB(orderID)
		if err != nil {
			return nil, err
		}
		r.ordersMemory.Lock()
		r.ordersMemory.ByID[orderID] = o
		r.ordersMemory.Unlock()
	}
	return &o, nil
}

// GetOrders - метод, возвращающий все заказы пользователя по его ID из БД.
func (r *Repository) GetOrders(userID uint64) ([]*entity.OrderX, error) {
	ors, err := r.GetOrdersDB(userID)
	if err != nil {
		return nil, err
	}
	layout := "2006-01-02T15:04:05-07:00"
	orsPr := make([]*entity.OrderX, 0)
	for _, o := range ors {
		po := &entity.OrderX{
			Number:     fmt.Sprint(o.ID),
			Status:     strings.TrimSpace(o.Status),
			Accrual:    float64(o.Accrual) / 100,
			UploadedAt: o.UploadedAt.Format(layout),
		}
		orsPr = append(orsPr, po)
	}
	return orsPr, nil
}

// GetBalance - метод, возвращающий баланс системы лояльности пользователя из хэш-таблицы или БД по его ID.
func (r *Repository) GetBalance(userID uint64) (*entity.BalanceX, error) {
	var err error
	r.balanceMemory.RLock()
	b, ok := r.balanceMemory.ByUserID[userID]
	r.balanceMemory.RUnlock()
	if !ok {
		b, err = r.GetBalanceDB(userID)
		if err != nil {
			return nil, err
		}
		r.balanceMemory.Lock()
		r.balanceMemory.ByUserID[userID] = b
		r.balanceMemory.Unlock()
	}
	blx := &entity.BalanceX{
		Current:   float64(b.Current) / 100,
		Withdrawn: float64(b.Withdrawn) / 100,
	}
	return blx, nil
}

// PostWithdraw - метод, регистрирующий новое списание из системы лояльности пользователем.
func (r *Repository) PostWithdraw(wd *entity.WithdrawX) error {
	orderID, err := strconv.Atoi(wd.Order)
	if err != nil {
		return ErrOrderInvalidFormat
	}
	withdraw := &entity.Withdraw{
		OrderID: uint64(orderID),
		UserID:  wd.UserID,
		Sum:     uint64(wd.Sum * 100),
	}
	strOrderID := strconv.Itoa(int(withdraw.OrderID))
	if !loon.IsValid(strOrderID) {
		return ErrOrderInvalidFormat
	}
	err = r.AddWithdrawDB(withdraw)
	if err != nil {
		return err
	}
	r.balanceMemory.Lock()
	delete(r.balanceMemory.ByUserID, withdraw.UserID)
	r.balanceMemory.Unlock()
	return nil
}

// GetWithdrawals - метод, возвращающий все списания пользователем из системы по его ID.
func (r *Repository) GetWithdrawals(userID uint64) ([]entity.WithdrawX, error) {
	wds, err := r.GetWithdrawalsDB(userID)
	if err != nil {
		return nil, err
	}
	if len(wds) == 0 {
		return nil, ErrNoContent
	}
	wdx := make([]entity.WithdrawX, 0)
	for _, v := range wds {
		wpr := entity.WithdrawX{
			Order:       fmt.Sprint(v.OrderID),
			Sum:         float64(v.Sum) / 100,
			ProcessedAt: v.ProcessedAt.Format(time.RFC3339),
		}
		if wpr.Order == "" || wpr.Sum == 0 || wpr.ProcessedAt == "" {
			continue
		}
		wdx = append(wdx, wpr)
	}
	return wdx, nil
}

// HashPass - функция, хэширующая пароль пользователя.
func HashPass(password string) ([]byte, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		return nil, err
	}
	return hashedPassword, nil
}
