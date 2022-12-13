package repository

import (
	"database/sql"
	"fmt"
	"github.com/gtgaleevtimur/gofermart/internal/service"
	"time"
)

// GetOrder - метод возвращающий заказ из БД по его номеру.
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
	if err = tx.Commit(); err != nil {
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

func (d *Database) PullOrders(limit uint32) (map[uint64]*service.Order, error) {
	orders := make(map[uint64]*service.Order)
	tx, err := d.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	get, err := tx.Prepare("SELECT * FROM orders WHERE status='NEW' or status='PROCESSING' order by uploaded_at LIMIT $1")
	if err != nil {
		return nil, err
	}
	txGet := tx.StmtContext(d.ctx, get)

	rows, err := txGet.QueryContext(d.ctx, limit)
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, err
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

		orders[order.ID] = &order
	}

	return orders, nil
}

func (d *Database) UpdateOrder(o *service.Order) error {
	tx, err := d.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	updateOrder, err := tx.Prepare("UPDATE orders SET status = $2, accrual = $3 WHERE id = $1")
	if err != nil {
		return err
	}
	txUpdateOrder := tx.StmtContext(d.ctx, updateOrder)

	updateBalance, err := tx.Prepare("UPDATE balances SET current = $2, withdrawn = $3 WHERE user_id = $1")
	if err != nil {
		return err
	}
	txUpdateBalance := tx.StmtContext(d.ctx, updateBalance)

	getBalance, err := tx.Prepare("SELECT * FROM balances WHERE user_id=$1")
	if err != nil {
		return err
	}
	txGetBalance := tx.StmtContext(d.ctx, getBalance)

	// обновим заказ
	if o.Status == "PROCESSED" {
		// записываем начисление только для заказов со статусом выполнено
		_, err = txUpdateOrder.ExecContext(d.ctx, o.ID, o.Status, o.Accrual)
		if err != nil {
			return fmt.Errorf("failed to update order - %w", err)
		}

		// обновим баланс пользователя: сначала получим текущее значение
		b := &service.Balance{}
		row := txGetBalance.QueryRowContext(d.ctx, o.UserID)
		err = row.Scan(&b.UserID, &b.Current, &b.Withdrawn)
		if err != nil {
			return fmt.Errorf("failed to get user balance - %w", err)
		}
		current := b.Current + o.Accrual // прибавим начисленные баллы
		// обновим баланс с новым значением
		_, err = txUpdateBalance.ExecContext(d.ctx, b.UserID, current, b.Withdrawn)
		if err != nil {
			return fmt.Errorf("failed to update user balance - %w", err)
		}
	} else {
		// для всех остальных статусов - начисления не записываем, баланс не обновляем
		_, err = txUpdateOrder.ExecContext(d.ctx, o.ID, o.Status, o.Accrual)
		if err != nil {
			return fmt.Errorf("failed to update order - %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("update order transaction failed - %w", err)
	}
	return nil
}
