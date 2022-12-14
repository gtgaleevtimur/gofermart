package repository

import (
	"context"
	"github.com/rs/zerolog/log"
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
