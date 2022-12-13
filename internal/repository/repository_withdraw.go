package repository

import (
	"database/sql"
	"fmt"
	"github.com/gtgaleevtimur/gofermart/internal/service"
	"time"
)

func (d *Database) AddWithdraw(withdraw *service.Withdraw) error {
	tx, err := d.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	getByID, err := tx.Prepare("SELECT * FROM withdrawals WHERE order_id=$1")
	if err != nil {
		return service.ErrWithdrawalPostTransaction
	}
	txGetByID := tx.StmtContext(d.ctx, getByID)

	insertWithdraw, err := tx.Prepare("INSERT INTO withdrawals (order_id, user_id, sum, processed_at) VALUES ($1, $2, $3, $4)")
	if err != nil {
		return service.ErrWithdrawalPostTransaction
	}
	txInsertWithdraw := tx.StmtContext(d.ctx, insertWithdraw)

	getBalance, err := tx.Prepare("SELECT * FROM balances WHERE user_id=$1")
	if err != nil {
		return service.ErrWithdrawalPostTransaction
	}
	txGetBalance := tx.StmtContext(d.ctx, getBalance)

	updateBalance, err := tx.Prepare("UPDATE balances SET current = $2, withdrawn = $3 WHERE user_id = $1")
	if err != nil {
		return service.ErrWithdrawalPostTransaction
	}
	txUpdateBalance := tx.StmtContext(d.ctx, updateBalance)

	// проверим баланс
	var balance service.Balance
	row := txGetBalance.QueryRowContext(d.ctx, withdraw.UserID)
	err = row.Scan(&balance.UserID, &balance.Current, &balance.Withdrawn)
	if err == sql.ErrNoRows {
		return service.ErrWithdrawalPostTransaction
	}
	if err != nil {
		return service.ErrWithdrawalPostTransaction
	}
	if balance.Current < withdraw.Sum {
		return service.ErrNotEnoughFunds
	}

	// средств достаточно, обновим баланс
	current := balance.Current - withdraw.Sum
	withdrawn := balance.Withdrawn + withdraw.Sum
	_, err = txUpdateBalance.ExecContext(d.ctx, withdraw.UserID, current, withdrawn)
	if err != nil {
		return service.ErrWithdrawalPostTransaction
	}

	// добавим историю списаний
	var bw service.Withdraw
	date := new(string)
	row = txGetByID.QueryRowContext(d.ctx, withdraw.OrderID)
	err = row.Scan(&bw.OrderID, &bw.UserID, &bw.Sum, date)
	if err != nil {
		if err == sql.ErrNoRows {
			// добавим новую запись в случае отсутствия результата
			_, err = txInsertWithdraw.ExecContext(d.ctx, withdraw.OrderID, withdraw.UserID, withdraw.Sum, time.Now())
			if err != nil {
				return err
			}

			// всё хорошо, выполним транзакцию
			err = tx.Commit()
			if err != nil {
				return service.ErrWithdrawalPostTransaction
			}
			return nil
		}

		return err
	}

	if withdraw.UserID == bw.UserID {
		return fmt.Errorf("withdraw already recorded by this user")
	}
	return fmt.Errorf("withdraw already recorded by another user")
}

func (d *Database) GetUserWithdrawals(userID uint64) ([]*service.Withdraw, error) {
	ws := make([]*service.Withdraw, 0)

	tx, err := d.DB.Begin()
	if err != nil {
		return nil, service.ErrWithdrawalsGetTransaction
	}
	defer tx.Rollback()

	get, err := tx.Prepare("SELECT * FROM withdrawals WHERE user_id=$1 ORDER BY processed_at desc")
	if err != nil {
		return nil, service.ErrWithdrawalsGetTransaction
	}
	txGet := tx.StmtContext(d.ctx, get)

	rows, err := txGet.QueryContext(d.ctx, userID)
	if err != nil {
		return nil, service.ErrWithdrawalsGetTransaction
	}
	if rows.Err() != nil {
		return nil, service.ErrWithdrawalsGetTransaction
	}
	defer rows.Close()

	for rows.Next() {
		var w service.Withdraw
		date := new(string)

		err = rows.Scan(&w.OrderID, &w.UserID, &w.Sum, date)
		if err != nil {
			return nil, err
		}

		if w.ProcessedAt, err = time.Parse(time.RFC3339, *date); err != nil {
			return nil, err
		}

		ws = append(ws, &w)
	}

	return ws, nil
}
