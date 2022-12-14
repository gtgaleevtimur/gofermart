package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gtgaleevtimur/gofermart/internal/entity"
	"github.com/rs/zerolog/log"
	"time"
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

func (r *Repository) GetOrderDB(orderID uint64) (*entity.Order, error) {
	o := &entity.Order{}
	accrual := new(sql.NullInt64)
	date := new(string)

	row := r.stmts["orderGetByID"].QueryRowContext(r.ctx, orderID)
	err := row.Scan(&o.ID, &o.UserID, &o.Status, accrual, date)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("order not found - %s", err.Error())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get order - %s", err.Error())
	}

	if accrual.Valid {
		o.Accrual = uint64(accrual.Int64)
	}

	if o.UploadedAt, err = time.Parse(time.RFC3339, *date); err != nil {
		return nil, err
	}

	return o, nil
}

func (r *Repository) AddOrderDB(o *entity.Order) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txInsert := tx.StmtContext(r.ctx, r.stmts["ordersInsert"])
	txGetByID := tx.StmtContext(r.ctx, r.stmts["ordersGetByID"])
	var bo entity.Order
	date := new(string)
	accrual := new(sql.NullInt64)

	row := txGetByID.QueryRowContext(r.ctx, o.ID)
	err = row.Scan(&bo.ID, &bo.UserID, &bo.Status, accrual, date)
	if err != nil {
		if err == sql.ErrNoRows {
			// добавим новую запись в случае отсутствия результата
			_, err = txInsert.ExecContext(r.ctx, o.ID, o.UserID, o.Status, o.UploadedAt)
			if err != nil {
				return err
			}

			// всё хорошо, выполним транзакцию
			err = tx.Commit()
			if err != nil {
				return fmt.Errorf("add order transaction failed - %s", err.Error())
			}
			return nil
		}

		return err
	}
	if bo.UserID == o.UserID {
		return ErrOrderAlreadyLoadedByUser
	}
	return ErrOrderAlreadyLoadedByAnotherUser
}

func (r *Repository) GetOrdersDB(id uint64) ([]*entity.Order, error) {
	orders := make([]*entity.Order, 0)
	rows, err := r.stmts["ordersGetForUser"].QueryContext(r.ctx, id)
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var bo entity.Order
		accrual := new(sql.NullInt64)
		date := new(string)

		err = rows.Scan(&bo.ID, &bo.UserID, &bo.Status, accrual, date)
		if err != nil {
			return nil, err
		}

		if accrual.Valid {
			bo.Accrual = uint64(accrual.Int64)
		}

		if bo.UploadedAt, err = time.Parse(time.RFC3339, *date); err != nil {
			return nil, err
		}
		orders = append(orders, &bo)
	}
	return orders, nil
}
