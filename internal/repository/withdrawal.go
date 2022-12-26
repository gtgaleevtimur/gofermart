package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/gtgaleevtimur/gofermart/internal/entity"
)

// initWithdrawals - метод, создающий таблицу со средствами пользователей. Подготавливает стейтменты для базы данных.
func (r *Repository) initWithdrawals(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS withdrawals (
				order_id bigint PRIMARY KEY NOT NULL,
				user_id bigint NOT NULL,
				sum bigint NOT NULL,
				processed_at timestamp NOT NULL)`)
	if err != nil {
		return err
	}
	log.Debug().Msg("table withdrawals created")
	err = r.initWithdrawalsStatements()
	if err != nil {
		return err
	}
	return nil
}

// initWithdrawalsStatements - метод, подготавливающий стейтменты БД для работы с таблицей списаний пользователей.
func (r *Repository) initWithdrawalsStatements() error {
	stmt, err := r.db.PrepareContext(
		r.ctx,
		"INSERT INTO withdrawals (order_id, user_id, sum, processed_at) VALUES ($1, $2, $3, $4)",
	)
	if err != nil {
		return err
	}
	r.stmts["withdrawalsInsert"] = stmt
	stmt, err = r.db.PrepareContext(
		r.ctx,
		"SELECT * FROM withdrawals WHERE order_id=$1",
	)
	if err != nil {
		return err
	}
	r.stmts["withdrawalsGetByID"] = stmt
	stmt, err = r.db.PrepareContext(
		r.ctx,
		"SELECT * FROM withdrawals WHERE user_id=$1 ORDER BY processed_at DESC",
	)
	if err != nil {
		return err
	}
	r.stmts["withdrawalsGetForUser"] = stmt
	return nil
}

// AddWithdrawDB - метод, добавляющий списание баллов лояльности пользователя в БД.
func (r *Repository) AddWithdrawDB(withdraw *entity.Withdraw) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	txGetByID := tx.StmtContext(r.ctx, r.stmts["withdrawalsGetByID"])
	txInsertWithdrawal := tx.StmtContext(r.ctx, r.stmts["withdrawalsInsert"])
	txGetBalance := tx.StmtContext(r.ctx, r.stmts["balanceGet"])
	txUpdateBalance := tx.StmtContext(r.ctx, r.stmts["balanceUpdate"])
	var balance entity.Balance
	row := txGetBalance.QueryRowContext(r.ctx, withdraw.UserID)
	err = row.Scan(&balance.UserID, &balance.Current, &balance.Withdrawn)
	if err == sql.ErrNoRows {
		return fmt.Errorf("user balance not found - %s", err.Error())
	}
	if err != nil {
		return fmt.Errorf("failed to get user balance - %s", err.Error())
	}
	if balance.Current < withdraw.Sum {
		return ErrNotEnoughFunds
	}
	current := balance.Current - withdraw.Sum
	withdrawn := balance.Withdrawn + withdraw.Sum
	_, err = txUpdateBalance.ExecContext(r.ctx, withdraw.UserID, current, withdrawn)
	if err != nil {
		return fmt.Errorf("failed to update user balance - %s", err.Error())
	}

	var bw entity.Withdraw
	date := new(string)
	row = txGetByID.QueryRowContext(r.ctx, withdraw.OrderID)
	err = row.Scan(&bw.OrderID, &bw.UserID, &bw.Sum, date)
	if err != nil {
		if err == sql.ErrNoRows {
			_, err = txInsertWithdrawal.ExecContext(r.ctx, withdraw.OrderID, withdraw.UserID, withdraw.Sum, time.Now())
			if err != nil {
				return err
			}
			err = tx.Commit()
			if err != nil {
				return fmt.Errorf("add order transaction failed - %s", err.Error())
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

// GetWithdrawalsDB - метод, возвращающий сделанные пользователем списания с баланса системы лояльности из БД по его ID.
func (r *Repository) GetWithdrawalsDB(userID uint64) ([]entity.Withdraw, error) {
	ws := make([]entity.Withdraw, 0)
	rows, err := r.stmts["withdrawalsGetForUser"].QueryContext(r.ctx, userID)
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var w entity.Withdraw
		date := new(string)
		err = rows.Scan(&w.OrderID, &w.UserID, &w.Sum, date)
		if err != nil {
			return nil, err
		}
		if w.ProcessedAt, err = time.Parse(time.RFC3339, *date); err != nil {
			return nil, err
		}
		ws = append(ws, w)
	}
	return ws, nil
}
