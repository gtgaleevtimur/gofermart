package repository

import (
	"context"
	"database/sql"
	"github.com/rs/zerolog/log"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

// Database - структура базы данных SQL.
type Database struct {
	DB     *sql.DB
	ctx    context.Context
	cancel context.CancelFunc
}

// NewDatabaseDSN - конструктор базы данных на основе SQL.
func NewDatabaseDSN(databaseDSN string) (*Database, error) {
	// Инициализация общего контекста.
	ctx, cancel := context.WithCancel(context.Background())
	s := &Database{
		ctx:    ctx,
		cancel: cancel,
	}
	// Соединение и проверка таблиц.
	err := s.Connect(databaseDSN)
	if err != nil {
		return nil, err
	}
	err = s.Bootstrap()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (d *Database) Bootstrap() (err error) {
	// Инициализация контекста по времени.
	ctx, cancel := context.WithTimeout(d.ctx, 10*time.Second)
	defer cancel()
	// Выполняем SQL запрос на создание таблиц сервиса, если их еще нет.
	_, err = d.DB.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS users (id serial PRIMARY KEY,
													login varchar NOT NULL,
													password bytea NOT NULL`)
	if err != nil {
		log.Info().Msg("Failed to create users table.")
		return err
	}
	log.Info().Msg("Table users created.")

	_, err = d.DB.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS sessions (user_id bigint NOT NULL,
													token varchar NOT NULL,
													expiry time NOT NULL`)
	if err != nil {
		log.Info().Msg("Failed to create sessions table.")
		return err
	}
	log.Info().Msg("Table sessions created.")

	_, err = d.DB.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS orders (id bigint PRIMARY KEY NOT NULL,
													user_id bigint NOT NULL,
													status char(256) NOT NULL,
    												accrual bigint,
    												uploaded_at timestamp NOT NULL`)
	if err != nil {
		log.Info().Msg("Failed to create orders table.")
		return err
	}
	log.Info().Msg("Table orders created.")

	_, err = d.DB.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS balances (user_id bigint PRIMARY KEY NOT NULL,
													current bigint NOT NULL,
													withdrawn bigint NOT NULL`)
	if err != nil {
		log.Info().Msg("Failed to create balances table.")
		return err
	}
	log.Info().Msg("Table balances created.")

	_, err = d.DB.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS withdrawals (order_id bigint PRIMARY KEY NOT NULL,
													user_id bigint NOT NULL,
													sum bigint NOT NULL,
													processed_at timestamp NOT NULL`)
	if err != nil {
		log.Info().Msg("Failed to create withdrawals table.")
		return err
	}
	log.Info().Msg("Table withdrawals created.")

	// Выставляем настройки базы данных.
	d.DB.SetMaxIdleConns(20)
	d.DB.SetMaxOpenConns(40)
	d.DB.SetConnMaxIdleTime(time.Second * 60)
	return nil
}

// Connect - метод выполняет соединение с базой данных.
func (d *Database) Connect(databaseDSN string) (err error) {
	d.DB, err = sql.Open("pgx", databaseDSN)
	if err != nil {
		return err
	}
	err = d.DB.Ping()
	if err != nil {
		return err
	}
	return nil
}

// Shutdown - метод закрытия соединения с БД.
func (d *Database) Shutdown() error {
	// Закрываем все текущие запросы.
	d.cancel()
	err := d.DB.Close()
	if err != nil {
		return err
	}
	log.Info().Msg("connection database closed")
	return nil
}
