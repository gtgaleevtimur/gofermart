package repository

import (
	"context"
	"github.com/rs/zerolog/log"
)

func (r *Repository) initOrders(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS orders (
				id bigint PRIMARY KEY NOT NULL,
				user_id bigint NOT NULL,
				status char(256) NOT NULL, 
				accrual bigint,
				uploaded_at timestamp NOT NULL)`)
	if err != nil {
		return err
	}
	log.Debug().Msg("table orders created")

	err = r.initOrdersStatements()
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) initOrdersStatements() error {
	stmt, err := r.db.PrepareContext(
		r.ctx,
		"INSERT INTO orders (id, user_id, status, uploaded_at) VALUES ($1, $2, $3, $4)",
	)
	if err != nil {
		return err
	}
	r.stmts["ordersInsert"] = stmt

	stmt, err = r.db.PrepareContext(
		r.ctx,
		"SELECT * FROM orders WHERE id=$1",
	)
	if err != nil {
		return err
	}
	r.stmts["orderGetByID"] = stmt

	stmt, err = r.db.PrepareContext(
		r.ctx,
		"UPDATE orders SET status = $2, accrual = $3 WHERE id = $1",
	)
	if err != nil {
		return err
	}
	r.stmts["ordersUpdate"] = stmt

	stmt, err = r.db.PrepareContext(
		r.ctx,
		"SELECT * FROM orders WHERE id=$1",
	)
	if err != nil {
		return err
	}
	r.stmts["ordersGetByID"] = stmt

	stmt, err = r.db.PrepareContext(
		r.ctx,
		"SELECT * FROM orders WHERE user_id=$1 order by uploaded_at",
	)
	if err != nil {
		return err
	}
	r.stmts["ordersGetForUser"] = stmt

	stmt, err = r.db.PrepareContext(
		r.ctx,
		"SELECT * FROM orders WHERE status='NEW' or status='PROCESSING' order by uploaded_at LIMIT $1",
	)
	if err != nil {
		return err
	}
	r.stmts["ordersGetForPool"] = stmt

	return nil
}
