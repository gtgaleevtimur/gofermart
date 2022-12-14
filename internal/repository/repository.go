package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gtgaleevtimur/gofermart/internal/entity"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

type Repository struct {
	db            *sql.DB
	ctx           context.Context
	cancel        context.CancelFunc
	stmts         map[string]*sql.Stmt
	userMemory    *entity.UsersMemory
	sessionMemory *entity.SessionMemory
}

// NewRepository - конструктор новой базы данных.
func NewRepository(addr string) (*Repository, error) {
	ctx, cancel := context.WithCancel(context.Background())

	r := &Repository{
		ctx:           ctx,
		cancel:        cancel,
		stmts:         make(map[string]*sql.Stmt),
		userMemory:    entity.NewUsers(),
		sessionMemory: entity.NewSessions(),
	}
	err := r.init(addr)
	if err != nil {
		return nil, fmt.Errorf("database initialization failed - %s", err.Error())
	}

	return r, nil
}

func (r *Repository) init(addr string) (err error) {
	r.db, err = sql.Open("pgx", addr)
	if err != nil {
		return err
	}
	err = r.db.Ping()
	if err != nil {
		return err
	}

	r.db.SetMaxOpenConns(40)
	r.db.SetMaxIdleConns(20)
	r.db.SetConnMaxIdleTime(time.Second * 60)

	ctx, cancel := context.WithTimeout(r.ctx, time.Second*60)
	defer cancel()

	err = r.initUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to create 'users' table - %s", err.Error())
	}

	err = r.initSessions(ctx)
	if err != nil {
		return fmt.Errorf("failed to create 'sessions' table - %s", err.Error())
	}

	err = r.initBalance(ctx)
	if err != nil {
		return fmt.Errorf("failed to create 'balance' table - %s", err.Error())
	}

	err = r.initWithdrawals(ctx)
	if err != nil {
		return fmt.Errorf("failed to create 'withdrawals' table - %s", err.Error())
	}

	err = r.initOrders(ctx)
	if err != nil {
		return fmt.Errorf("failed to create 'orders' table - %s", err.Error())
	}

	return nil
}
