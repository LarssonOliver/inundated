package postgres_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	// "time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/larssonoliver/inundated/internal/db/postgres"
	"github.com/stretchr/testify/require"
	testcontainers "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func startPostgresContainer(ctx context.Context, t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()

	container, err := testcontainers.Run(ctx, "postgres:16-alpine",
		testcontainers.WithDatabase("testdb"),
		testcontainers.WithUsername("user"),
		testcontainers.WithPassword("secret"),
		testcontainers.BasicWaitStrategies(),
	)
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := fmt.Sprintf("postgres://user:secret@%s:%s/testdb?sslmode=disable", host, port.Port())

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)

	return pool, func() {
		pool.Close()
		container.Terminate(ctx)
	}
}

func TestMigrations_ApplyCleanly(t *testing.T) {
	ctx := context.Background()
	pool, terminate := startPostgresContainer(ctx, t)
	defer terminate()

	migrationsDir := filepath.Join("..", "migrations")

	err := postgres.ApplyMigrations(ctx, pool, migrationsDir)
	require.NoError(t, err)

	// Optional: verify tables exist
	var exists bool
	err = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = 'projects'
		)
	`).Scan(&exists)
	require.NoError(t, err)
	require.True(t, exists)
}
