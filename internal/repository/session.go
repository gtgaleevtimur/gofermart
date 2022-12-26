package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/gtgaleevtimur/gofermart/internal/entity"
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

// initSessionsStatements - метод, подготавливающий стейтменты БД для работы с сессиями пользователей.
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

// DeleteSessionDB - метод, удаляющий сессию из БД по его токену.
func (r *Repository) DeleteSessionDB(token string) error {
	res, err := r.stmts["sessionsDelete"].ExecContext(r.ctx, token)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

// AddSessionDB - метод, добавляющий сессию пользователя в БД.
func (r *Repository) AddSessionDB(session *entity.Session) error {
	_, err := r.stmts["sessionsInsert"].ExecContext(r.ctx, session.UserID, session.Token, session.Expiry)
	if err != nil {
		return err
	}
	return nil
}

// GetSessionDB - метод, возвращающий сессию пользователя по токену.
func (r *Repository) GetSessionDB(token string) (entity.Session, error) {
	session := &entity.Session{}
	row := r.stmts["sessionsGet"].QueryRowContext(r.ctx, token)
	err := row.Scan(&session.UserID, &session.Token, &session.Expiry)
	if err == sql.ErrNoRows {
		return *session, ErrSessionNotFound
	}
	if err != nil {
		return *session, fmt.Errorf("failed to get session - %s", err.Error())
	}
	return *session, nil
}
