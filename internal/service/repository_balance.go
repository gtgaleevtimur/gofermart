package repository

import (
	"database/sql"
	"github.com/gtgaleevtimur/gofermart/internal/service"
)

func (d *Database) GetBalance(userID uint64) (*service.Balance, error) {
	b := &service.Balance{}
	tx, err := d.DB.Begin()
	if err != nil {
		return nil, service.ErrBalanceGetTransaction
	}
	defer tx.Rollback()
	get, err := tx.Prepare("SELECT * FROM balances WHERE user_id=$1")
	if err != nil {
		return nil, service.ErrBalanceGetTransaction
	}
	txGet := tx.StmtContext(d.ctx, get)
	row := txGet.QueryRowContext(d.ctx, userID)
	err = row.Scan(&b.UserID, &b.Current, &b.Withdrawn)
	if err == sql.ErrNoRows {
		return nil, service.ErrBalanceGetTransaction
	}
	if err != nil {
		return nil, service.ErrBalanceGetTransaction
	}

	return b, nil
}
