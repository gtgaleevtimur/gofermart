package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gtgaleevtimur/gofermart/internal/service"
)

var ErrBalanceGetTransaction = errors.New("get balance transaction failed")

// GetBalance - метод возвращающий баланс пользователя из БД.
func (d *Database) GetBalance(userID uint64) (*service.Balance, error) {
	b := &service.Balance{}
	tx, err := d.DB.Begin()
	if err != nil {
		return nil, ErrBalanceGetTransaction
	}
	defer tx.Rollback()
	get, err := tx.Prepare("SELECT * FROM balances WHERE user_id=$1")
	if err != nil {
		return nil, ErrBalanceGetTransaction
	}
	txGet := tx.StmtContext(d.ctx, get)
	row := txGet.QueryRowContext(d.ctx, userID)
	err = row.Scan(&b.UserID, &b.Current, &b.Withdrawn)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user balance not found - %w", err)
	}
	if err != nil {
		return nil, ErrBalanceGetTransaction
	}
	if err = tx.Commit(); err != nil {
		return nil, ErrBalanceGetTransaction
	}
	return b, nil
}
