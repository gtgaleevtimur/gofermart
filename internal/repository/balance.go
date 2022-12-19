package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/gtgaleevtimur/gofermart/internal/entity"
)

// initBalance - метод, создающий таблицу балансов пользователя, если ее нет. Подготавливает стейтменты для базы данных.
func (r *Repository) initBalance(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS balance (
				user_id bigint PRIMARY KEY NOT NULL,
				current bigint NOT NULL,
				withdrawn bigint NOT NULL)`)
	if err != nil {
		return err
	}
	log.Debug().Msg("table balance created")
	err = r.initBalanceStatements()
	if err != nil {
		return err
	}
	return nil
}

// initBalanceStatements - метод, подготавливающий стейтменты для работы с таблицей балансов пользователей.
func (r *Repository) initBalanceStatements() error {
	stmt, err := r.db.PrepareContext(
		r.ctx,
		"INSERT INTO balance (user_id, current, withdrawn) VALUES ($1, 0, 0)",
	)
	if err != nil {
		return err
	}
	r.stmts["balanceInsert"] = stmt
	stmt, err = r.db.PrepareContext(
		r.ctx,
		"SELECT * FROM balance WHERE user_id=$1",
	)
	if err != nil {
		return err
	}
	r.stmts["balanceGet"] = stmt
	stmt, err = r.db.PrepareContext(
		r.ctx,
		"UPDATE balance SET current = $2, withdrawn = $3 WHERE user_id = $1",
	)
	if err != nil {
		return err
	}
	r.stmts["balanceUpdate"] = stmt
	return nil
}

// GetBalanceDB - метод, возвращающий баланс пользователя по его ID.
func (r *Repository) GetBalanceDB(userID uint64) (entity.Balance, error) {
	b := entity.Balance{}
	row := r.stmts["balanceGet"].QueryRowContext(r.ctx, userID)
	err := row.Scan(&b.UserID, &b.Current, &b.Withdrawn)
	if err == sql.ErrNoRows {
		return b, fmt.Errorf("user balance not found - %s", err.Error())
	}
	if err != nil {
		return b, fmt.Errorf("failed to get user balance - %s", err.Error())
	}
	return b, nil
}
