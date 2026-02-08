package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/larssonoliver/inundated/internal/db/postgres"
	"github.com/stretchr/testify/require"
	testcontainers "github.com/testcontainers/testcontainers-go/modules/postgres"
)

type migrationTestCase struct {
	name        string
	fromVersion uint
	toVersion   uint

	before func(t *testing.T, ctx context.Context, pool *pgxpool.Pool)
	after  func(t *testing.T, ctx context.Context, pool *pgxpool.Pool)
}

func sptr(s string) *string {
	return &s
}

func startPostgresContainer(ctx context.Context, t *testing.T) (*pgxpool.Pool, string, func()) {
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

	return pool, dsn, func() {
		pool.Close()
		container.Terminate(ctx)
	}
}

func TestMigrations_ApplyCleanly(t *testing.T) {
	ctx := context.Background()
	pool, dsn, terminate := startPostgresContainer(ctx, t)
	defer terminate()

	err := postgres.ApplyMigrations(ctx, dsn)
	require.NoError(t, err)

	assertTableExists(t, ctx, pool, "projects")
	assertTableExists(t, ctx, pool, "tags")
	assertTableExists(t, ctx, pool, "timespans")
}

func TestIndividualMigrations(t *testing.T) {
	ctx := context.Background()

	tests := []migrationTestCase{
		{
			name:        "0001_init",
			fromVersion: 0,
			toVersion:   1,
			before: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertTableNotExists(t, ctx, pool, "tags")
				assertTableNotExists(t, ctx, pool, "projects")
				assertTableNotExists(t, ctx, pool, "timespans")
			},
			after: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertTableExists(t, ctx, pool, "tags")
				assertColumnExists(t, ctx, pool, "tags", "id", sptr("uuid"))

				assertTableExists(t, ctx, pool, "projects")
				assertColumnExists(t, ctx, pool, "projects", "id", sptr("uuid"))
				assertTableExists(t, ctx, pool, "project_tags")

				assertTableExists(t, ctx, pool, "timespans")
				assertColumnExists(t, ctx, pool, "timespans", "id", sptr("uuid"))
				assertTableExists(t, ctx, pool, "timespan_tags")
			},
		},
	}

	for _, tc := range tests {
		tcase := tc // capture
		t.Run(tcase.name, func(t *testing.T) {
			t.Parallel()

			pool, dsn, terminate := startPostgresContainer(ctx, t)
			defer terminate()

			require.NoError(t,
				postgres.ApplyMigrationsUpTo(ctx, dsn, tcase.fromVersion),
			)

			if tcase.before != nil {
				tcase.before(t, ctx, pool)
			}

			require.NoError(t,
				postgres.ApplyMigrationsUpTo(ctx, dsn, tcase.toVersion),
			)

			if tcase.after != nil {
				tcase.after(t, ctx, pool)
			}
		})
	}
}
