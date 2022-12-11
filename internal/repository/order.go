package repository

import (
	"database/sql"
	"github.com/gtgaleevtimur/gofermart/internal/service"
	"time"
)

func (d *Database) GetOrder(orderID uint64) (*service.Order, error) {
	order := &service.Order{}
	accrual := new(sql.NullInt64)
	date := new(string)
	tx, err := d.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	get, err := tx.Prepare("SELECT * FROM orders WHERE id=$1")
	if err != nil {
		return nil, err
	}
	txGet := tx.StmtContext(d.ctx, get)
	row := txGet.QueryRowContext(d.ctx, orderID)
	err = row.Scan(&order.ID, &order.UserID, &order.Status, accrual, date)
	if err == sql.ErrNoRows {
		return nil, service.ErrOrderNotFound
	}
	if err != nil {
		return nil, service.ErrOrderGetTransaction
	}
	if accrual.Valid {
		order.Accrual = uint64(accrual.Int64)
	}
	if order.UploadedAt, err = time.Parse(time.RFC3339, *date); err != nil {
		return nil, err
	}
	return order, nil
}

func (d *Database) AddOrder(o *service.Order) error {
	tx, err := d.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	insert, err := tx.Prepare("INSERT INTO orders (id, user_id, status, uploaded_at) VALUES ($1, $2, $3, $4)")
	if err != nil {
		return err
	}
	get, err := tx.Prepare("SELECT * FROM orders WHERE id=$1")
	if err != nil {
		return err
	}
	txInsert := tx.StmtContext(d.ctx, insert)
	txGet := tx.StmtContext(d.ctx, get)
	var order service.Order
	date := new(string)
	accrual := new(sql.NullInt64)

	row := txGet.QueryRowContext(d.ctx, o.ID)
	err = row.Scan(&order.ID, &order.UserID, &order.Status, accrual, date)
	if err != nil {
		if err == sql.ErrNoRows {
			// добавим новую запись в случае отсутствия результата
			_, err = txInsert.ExecContext(d.ctx, o.ID, o.UserID, o.Status, o.UploadedAt)
			if err != nil {
				return err
			}

			// всё хорошо, выполним транзакцию
			err = tx.Commit()
			if err != nil {
				return service.ErrOrderAddTransaction
			}
			return nil
		}
		return err
	}
	err = tx.Commit()
	if err != nil {
		return service.ErrOrderAddTransaction
	}
	// заказ уже существует, обработаем ошибку
	if order.UserID == o.UserID {
		return service.ErrOrderAlreadyLoadedByUser
	}
	return service.ErrOrderAlreadyLoadedByAnotherUser
}

func (d *Database) GetUserOrders(id uint64) ([]*service.Order, error) {
	orders := make([]*service.Order, 0)
	tx, err := d.DB.Begin()
	if err != nil {
		return nil, service.ErrOrderGetTransactions
	}
	defer tx.Rollback()
	get, err := tx.Prepare("SELECT * FROM orders WHERE user_id=$1 order by uploaded_at")
	if err != nil {
		return nil, service.ErrOrderGetTransactions
	}
	txGet := tx.StmtContext(d.ctx, get)
	rows, err := txGet.QueryContext(d.ctx, id)
	if err != nil {
		return nil, service.ErrOrderGetTransactions
	}
	if rows.Err() != nil {
		return nil, service.ErrOrderGetTransactions
	}
	defer rows.Close()

	for rows.Next() {
		var order service.Order
		accrual := new(sql.NullInt64)
		date := new(string)

		err = rows.Scan(&order.ID, &order.UserID, &order.Status, accrual, date)
		if err != nil {
			return nil, err
		}

		if accrual.Valid {
			order.Accrual = uint64(accrual.Int64)
		}

		if order.UploadedAt, err = time.Parse(time.RFC3339, *date); err != nil {
			return nil, err
		}

		orders = append(orders, &order)
	}
	err = tx.Commit()
	if err != nil {
		return nil, service.ErrOrderGetTransactions
	}

	return orders, nil
}
