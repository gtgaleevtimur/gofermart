package repository

import (
	"database/sql"
	"github.com/gtgaleevtimur/gofermart/internal/service"
)

func (d *Database) AddUser(user *service.User) (uint64, error) {
	tx, err := d.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	stGet, err := tx.Prepare(`SELECT * FROM users WHERE login=$1`)
	if err != nil {
		return 0, err
	}
	stAdd, err := tx.Prepare(`INSERT INTO users (login, password) VALUES ($1, $2)`)
	if err != nil {
		return 0, err
	}
	stInsBal, err := tx.Prepare(`iNSERT INTO balances (user_id, current, withdrawn) VALUES ($1, 0, 0)`)
	if err != nil {
		return 0, err
	}
	txAdd := tx.StmtContext(d.ctx, stAdd)
	txGet := tx.StmtContext(d.ctx, stGet)
	txInsBal := tx.StmtContext(d.ctx, stInsBal)
	row := txGet.QueryRowContext(d.ctx, user.Login)
	u := service.User{}
	err = row.Scan(&u.ID, &u.Login, &u.Password)
	if err == sql.ErrNoRows {
		_, err = txAdd.ExecContext(d.ctx, user.Login, user.Password)
		if err != nil {
			return 0, err
		}
		row = txGet.QueryRowContext(d.ctx, user.Login)
		err = row.Scan(&u.ID, &u.Login, &u.Password)
		if err != nil {
			return 0, err
		}
		_, err = txInsBal.ExecContext(d.ctx, user.ID)
		if err != nil {
			return 0, err
		}
	} else if err != nil {
		return 0, err
	} else {
		return 0, service.ErrLoginAlreadyTaken
	}
	err = tx.Commit()
	if err != nil {
		return 0, service.ErrUserAddedTransaction
	}
	return u.ID, nil
}

func (d *Database) GetUser(key interface{}) (*service.User, error) {
	tx, err := d.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	stById, err := tx.Prepare("SELECT * FROM users WHERE id=$1")
	if err != nil {
		return nil, err
	}
	stByLog, err := tx.Prepare("SELECT * FROM users WHERE login=$1")
	if err != nil {
		return nil, err
	}
	txById := tx.StmtContext(d.ctx, stById)
	txByLog := tx.StmtContext(d.ctx, stByLog)
	u := service.User{}
	var row *sql.Row
	switch k := key.(type) {
	case string:
		row = txByLog.QueryRowContext(d.ctx, k)
	case uint64:
		row = txById.QueryRowContext(d.ctx, k)
	default:
		return nil, service.ErrTypeNotAllowed
	}
	err = row.Scan(&u.ID, &u.Login, &u.Password)
	if err == sql.ErrNoRows {
		return nil, service.ErrUserNotFound
	} else if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, service.ErrUserGetTransaction
	}
	return &u, nil
}
