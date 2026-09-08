package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/larssonoliver/inundated/internal/repository"
)

type PostgresStore struct {
	db Querier
}

var _ repository.Repository = (*PostgresStore)(nil)
var _ repository.SessionRepository = (*PostgresStore)(nil)
var _ repository.LoginStateRepository = (*PostgresStore)(nil)

// New creates a Repository by opening a new pgxpool.
func NewPostgresStore(ctx context.Context, connString string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("creating pgx pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging postgres: %w", err)
	}
	return &PostgresStore{db: pool}, nil
}

func NewPostgresStoreFromPool(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{db: pool}
}

func NewPostgresStoreWithQuerier(q Querier) *PostgresStore {
	return &PostgresStore{db: q}
}

type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

var _ Querier = (*pgxpool.Pool)(nil)

// committing on success and rolling back on any error or panic.
func (r *PostgresStore) withTx(ctx context.Context, fn func(q Querier) error) (err error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			rollback(ctx, tx)
			panic(p)
		}
		if err != nil {
			rollback(ctx, tx)
		}
	}()
	if err = fn(tx); err != nil {
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func rollback(ctx context.Context, tx pgx.Tx) {
	rbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	_ = tx.Rollback(rbCtx)
}
