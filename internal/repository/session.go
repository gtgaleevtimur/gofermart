package repository

import (
	"context"
	"github.com/rs/zerolog/log"
)

// initSessions - метод, создающий таблицу сессии авторизации пользователей, если ее нет. Подготавливает стейтменты для базы данных.
func (r *Repository) initSessions(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS sessions (
				user_id bigint NOT NULL,
				token varchar NOT NULL, 
				expiry time NOT NULL)`)
	if err != nil {
		return err
	}
	log.Debug().Msg("table sessions created")

	err = r.initSessionsStatements()
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) initSessionsStatements() error {
	stmt, err := r.db.PrepareContext(
		r.ctx,
		"INSERT INTO sessions (user_id, token, expiry) VALUES ($1, $2, $3)",
	)
	if err != nil {
		return err
	}
	r.stmts["sessionsInsert"] = stmt

	stmt, err = r.db.PrepareContext(
		r.ctx,
		"SELECT * FROM sessions WHERE token=$1",
	)
	if err != nil {
		return err
	}
	r.stmts["sessionsGet"] = stmt

	stmt, err = r.db.PrepareContext(
		r.ctx,
		"DELETE FROM sessions WHERE token=$1",
	)
	if err != nil {
		return err
	}
	r.stmts["sessionsDelete"] = stmt

	return nil
}
