package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/gtgaleevtimur/gofermart/internal/entity"
)

// initOrders - метод, создающий таблицу заказов пользователя, если ее нет. Подготавливает стейтменты для базы данных.
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

// initOrdersStatements - метод, подготавливающий стейтменты БД для работы с таблицей заказов.
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

// GetOrderDB - метод, возвращающий информацию о заказе из БД по его ID.
func (r *Repository) GetOrderDB(orderID uint64) (entity.Order, error) {
	o := entity.Order{}
	accrual := new(sql.NullInt64)
	date := new(string)
	row := r.stmts["orderGetByID"].QueryRowContext(r.ctx, orderID)
	err := row.Scan(&o.ID, &o.UserID, &o.Status, accrual, date)
	if err == sql.ErrNoRows {
		return o, fmt.Errorf("order not found - %s", err.Error())
	}
	if err != nil {
		return o, fmt.Errorf("failed to get order - %s", err.Error())
	}
	if accrual.Valid {
		o.Accrual = uint64(accrual.Int64)
	}
	if o.UploadedAt, err = time.Parse(time.RFC3339, *date); err != nil {
		return o, err
	}
	return o, nil
}

// AddOrderDB - метод, добавляющий заказ пользователя в БД.
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
			_, err = txInsert.ExecContext(r.ctx, o.ID, o.UserID, o.Status, o.UploadedAt)
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
	if bo.UserID == o.UserID {
		return ErrOrderAlreadyLoadedByUser
	}
	return ErrOrderAlreadyLoadedByAnotherUser
}

// GetOrdersDB - метод, возвращающий заказы пользователя по его ID.
func (r *Repository) GetOrdersDB(id uint64) ([]entity.Order, error) {
	orders := make([]entity.Order, 0)
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
		orders = append(orders, bo)
	}
	return orders, nil
}

// GetPullOrders - метод, возвращающий заказы для обновления балансов пользователей в системе.
func (r *Repository) GetPullOrders(limit uint32) (map[uint64]entity.Order, error) {
	orders := make(map[uint64]entity.Order)
	rows, err := r.stmts["ordersGetForPool"].QueryContext(r.ctx, limit)
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
		orders[bo.ID] = bo
	}
	return orders, nil
}

// UpdateOrder - метод, обновляющий состояние заказа в БД.
func (r *Repository) UpdateOrder(o entity.Order) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	txUpdateOrder := tx.StmtContext(r.ctx, r.stmts["ordersUpdate"])
	txUpdateBalance := tx.StmtContext(r.ctx, r.stmts["balanceUpdate"])
	txGetBalance := tx.StmtContext(r.ctx, r.stmts["balanceGet"])
	if o.Status == "PROCESSED" {
		_, err = txUpdateOrder.ExecContext(r.ctx, o.ID, o.Status, o.Accrual)
		if err != nil {
			return fmt.Errorf("failed to update order - %s", err.Error())
		}
		b := &entity.Balance{}
		row := txGetBalance.QueryRowContext(r.ctx, o.UserID)
		err = row.Scan(&b.UserID, &b.Current, &b.Withdrawn)
		if err != nil {
			return fmt.Errorf("failed to get user balance - %s", err.Error())
		}
		current := b.Current + o.Accrual
		_, err = txUpdateBalance.ExecContext(r.ctx, b.UserID, current, b.Withdrawn)
		if err != nil {
			return fmt.Errorf("failed to update user balance - %s", err.Error())
		}
	} else {
		_, err = txUpdateOrder.ExecContext(r.ctx, o.ID, o.Status, o.Accrual)
		if err != nil {
			return fmt.Errorf("failed to update order - %s", err.Error())
		}
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("update order transaction failed - %s", err.Error())
	}
	return nil
}
