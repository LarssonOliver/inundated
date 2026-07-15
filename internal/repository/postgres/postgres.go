package postgres

import (
	"context"
	"fmt"

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

// NewFromPool creates a Repository from an existing *pgxpool.Pool.
func NewPostgresStoreFromPool(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{db: pool}
}

// NewWithQuerier creates a Repository from any Querier implementation.
// Intended for use in tests with pgxmock.
func NewPostgresStoreWithQuerier(q Querier) *PostgresStore {
	return &PostgresStore{db: q}
}

// Querier abstracts the pgxpool.Pool methods used by the repository.
// pgxmock satisfies this interface, making unit tests straightforward.
type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// Ensure *pgxpool.Pool satisfies Querier at compile time.
var _ Querier = (*pgxpool.Pool)(nil)
