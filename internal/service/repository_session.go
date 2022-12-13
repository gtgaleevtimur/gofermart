package repository

import (
	"database/sql"
	"github.com/gtgaleevtimur/gofermart/internal/service"
)

func (d *Database) DeleteSession(token string) error {
	tx, err := d.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	del, err := tx.Prepare("DELETE FROM sessions WHERE token=$1")
	if err != nil {
		return err
	}
	txDel := tx.StmtContext(d.ctx, del)
	res, err := txDel.ExecContext(d.ctx, token)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return service.ErrSessionNotFound
	}
	err = tx.Commit()
	if err != nil {
		return service.ErrSessionDeleteTransaction
	}
	return nil
}

func (d *Database) AddSession(session *service.Session) error {
	tx, err := d.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	add, err := tx.Prepare("INSERT INTO sessions (user_id, token, expiry) VALUES ($1, $2, $3)")
	if err != nil {
		return err
	}
	txAdd := tx.StmtContext(d.ctx, add)
	_, err = txAdd.ExecContext(d.ctx, session.UserID, session.Token, session.Expiry)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return service.ErrSessionAddTransaction
	}
	return nil
}

func (d *Database) GetSession(token string) (*service.Session, error) {
	session := &service.Session{}
	tx, err := d.DB.Begin()
	if err != nil {
		return nil, err
	}
	get, err := tx.Prepare("SELECT * FROM sessions WHERE token=$1")
	if err != nil {
		return nil, err
	}
	txGet := tx.StmtContext(d.ctx, get)
	row := txGet.QueryRowContext(d.ctx, token)
	err = row.Scan(&session.UserID, &session.Token, &session.Expiry)
	if err == sql.ErrNoRows {
		return nil, service.ErrSessionNotFound
	}
	if err != nil {
		return nil, service.ErrSessionGetTransaction
	}
	err = tx.Commit()
	if err != nil {
		return nil, service.ErrSessionGetTransaction
	}
	return session, nil
}
