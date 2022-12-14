package repository

import (
	"context"
	"github.com/rs/zerolog/log"
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
		"SELECT * FROM withdrawals WHERE user_id=$1 ORDER BY processed_at desc",
	)
	if err != nil {
		return err
	}
	r.stmts["withdrawalsGetForUser"] = stmt

	return nil
}
