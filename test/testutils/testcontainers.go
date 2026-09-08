package testutils

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	testcontainerspg "github.com/testcontainers/testcontainers-go/modules/postgres"

	dbpostgres "github.com/larssonoliver/inundated/internal/db/postgres"
)

// StartPostgresContainer starts a PostgreSQL container for testing and returns a connected pgxpool.Pool, the DSN, and a cleanup function.
func StartPostgresContainer(ctx context.Context, t *testing.T) (*pgxpool.Pool, string) {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping container tests in short mode")
	}

	container, err := testcontainerspg.Run(ctx, "postgres:16-alpine",
		testcontainerspg.WithDatabase("testdb"),
		testcontainerspg.WithUsername("user"),
		testcontainerspg.WithPassword("secret"),
		testcontainerspg.BasicWaitStrategies(),
	)
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := fmt.Sprintf("postgres://user:secret@%s:%s/testdb?sslmode=disable", host, port.Port())

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)

	t.Cleanup(func() {
		pool.Close()
		_ = container.Terminate(ctx)
	})

	return pool, dsn
}

// StartPostgresContainerWithMigrationsApplied starts a PostgreSQL container, applies migrations, and returns a connected pgxpool.Pool and a cleanup function.
func StartPostgresContainerWithMigrationsApplied(ctx context.Context, t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool, dsn := StartPostgresContainer(ctx, t)
	err := dbpostgres.ApplyMigrations(ctx, dsn)
	require.NoError(t, err)
	return pool
}
