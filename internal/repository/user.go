package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/gtgaleevtimur/gofermart/internal/entity"
)

// initUsers - метод, создающий таблицу пользователей, если ее нет. Подготавливает стейтменты для базы данных.
func (r *Repository) initUsers(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS users (
				id serial PRIMARY KEY,
				login varchar NOT NULL, 
				password bytea NOT NULL)`)
	if err != nil {
		return err
	}
	log.Debug().Msg("table users created")
	err = r.initUsersStatements()
	if err != nil {
		return err
	}
	return nil
}

// initUsersStatements - метод, добавляющий стейтменты БД для работы с таблицей пользователей.
func (r *Repository) initUsersStatements() error {
	stmt, err := r.db.PrepareContext(
		r.ctx,
		"INSERT INTO users (login, password) VALUES ($1, $2)",
	)
	if err != nil {
		return err
	}
	r.stmts["usersInsert"] = stmt
	stmt, err = r.db.PrepareContext(
		r.ctx,
		"SELECT * FROM users WHERE login=$1",
	)
	if err != nil {
		return err
	}
	r.stmts["usersGetByLogin"] = stmt
	stmt, err = r.db.PrepareContext(
		r.ctx,
		"SELECT * FROM users WHERE id=$1",
	)
	if err != nil {
		return err
	}
	r.stmts["usersGetByID"] = stmt
	stmt, err = r.db.PrepareContext(
		r.ctx,
		"DELETE FROM users WHERE login=$1",
	)
	if err != nil {
		return err
	}
	r.stmts["usersDelete"] = stmt
	return nil
}

// AddUserDB - метод, добавляющий пользователя в БД.
func (r *Repository) AddUserDB(u *entity.User) (uint64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	txInsert := tx.StmtContext(r.ctx, r.stmts["usersInsert"])
	txGet := tx.StmtContext(r.ctx, r.stmts["usersGetByLogin"])
	txInsertBalance := tx.StmtContext(r.ctx, r.stmts["balanceInsert"])
	row := txGet.QueryRowContext(r.ctx, u.Login)
	blankUser := entity.User{}
	err = row.Scan(&blankUser.ID, &blankUser.Login, &blankUser.Password)
	if err == sql.ErrNoRows {
		_, err = txInsert.ExecContext(r.ctx, u.Login, u.Password)
		if err != nil {
			return 0, err
		}
		row = txGet.QueryRowContext(r.ctx, u.Login)
		err = row.Scan(&u.ID, &u.Login, &u.Password)
		if err != nil {
			return 0, err
		}
		_, err = txInsertBalance.ExecContext(r.ctx, u.ID)
		if err != nil {
			return 0, err
		}
	} else if err != nil {
		return 0, err
	} else {
		return 0, ErrLoginAlreadyTaken
	}
	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("add user transaction failed - %s", err.Error())
	}
	return u.ID, nil
}

// GetUserDB - метод, возвращающий информацию о пользователе из таблицы пользователей.
func (r *Repository) GetUserDB(byKey interface{}) (entity.User, error) {
	var u entity.User
	tx, err := r.db.Begin()
	if err != nil {
		return u, err
	}
	defer tx.Rollback()
	txGetByLogin := tx.StmtContext(r.ctx, r.stmts["usersGetByLogin"])
	txGetByID := tx.StmtContext(r.ctx, r.stmts["usersGetByID"])
	var row *sql.Row
	switch key := byKey.(type) {
	case string:
		row = txGetByLogin.QueryRowContext(r.ctx, key)
	case uint64:
		row = txGetByID.QueryRowContext(r.ctx, key)
	default:
		return u, fmt.Errorf("given type not implemented")
	}
	err = row.Scan(&u.ID, &u.Login, &u.Password)
	if err == sql.ErrNoRows {
		return u, ErrUserNotFound
	}
	if err != nil {
		return u, err
	}
	err = tx.Commit()
	if err != nil {
		return u, fmt.Errorf("get user transaction failed - %s", err.Error())
	}

	return u, nil
}
