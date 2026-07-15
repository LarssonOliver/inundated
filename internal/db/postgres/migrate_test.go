package postgres_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/larssonoliver/inundated/internal/db/postgres"
	"github.com/larssonoliver/inundated/test/testutils"
	"github.com/stretchr/testify/require"
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

func TestMigrations_ApplyCleanly(t *testing.T) {
	ctx := context.Background()
	pool, dsn := testutils.StartPostgresContainer(ctx, t)

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
				assertForeignKeyExists(t, ctx, pool, "project_tags", "project_tags_project_id_fkey")
				assertForeignKeyExists(t, ctx, pool, "project_tags", "project_tags_tag_id_fkey")

				assertTableExists(t, ctx, pool, "timespans")
				assertColumnExists(t, ctx, pool, "timespans", "id", sptr("uuid"))
				assertTableExists(t, ctx, pool, "timespan_tags")
				assertForeignKeyExists(t, ctx, pool, "timespan_tags", "timespan_tags_timespan_id_fkey")
				assertForeignKeyExists(t, ctx, pool, "timespan_tags", "timespan_tags_tag_id_fkey")
			},
		},
		{
			name:        "0002_deleted_at",
			fromVersion: 1,
			toVersion:   2,
			after: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertColumnExists(t, ctx, pool, "tags", "deleted_at", sptr("timestamp with time zone"))
				assertIndexExists(t, ctx, pool, "idx_tags_active")
				assertColumnExists(t, ctx, pool, "projects", "deleted_at", sptr("timestamp with time zone"))
				assertIndexExists(t, ctx, pool, "idx_projects_active")
				assertColumnExists(t, ctx, pool, "timespans", "deleted_at", sptr("timestamp with time zone"))
				assertIndexExists(t, ctx, pool, "idx_timespans_active")
			},
		},
		{
			name:        "0003_create_users_table",
			fromVersion: 2,
			toVersion:   3,
			before: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertTableNotExists(t, ctx, pool, "users")
			},
			after: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertTableExists(t, ctx, pool, "users")
				assertColumnExists(t, ctx, pool, "users", "id", sptr("uuid"))
				assertColumnExists(t, ctx, pool, "users", "sub", sptr("text"))
				assertColumnExists(t, ctx, pool, "users", "email", sptr("text"))
				assertColumnExists(t, ctx, pool, "users", "name", sptr("text"))
			},
		},
		{
			name:        "0004_create_sessions_table",
			fromVersion: 3,
			toVersion:   4,
			before: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertTableNotExists(t, ctx, pool, "sessions")
			},
			after: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertTableExists(t, ctx, pool, "sessions")
				assertColumnExists(t, ctx, pool, "sessions", "id", sptr("uuid"))
				assertColumnExists(t, ctx, pool, "sessions", "user_id", sptr("uuid"))
				assertColumnExists(t, ctx, pool, "sessions", "sub", sptr("text"))
				assertColumnExists(t, ctx, pool, "sessions", "expires_at", sptr("timestamp with time zone"))
				assertIndexExists(t, ctx, pool, "idx_sessions_user_id")
				assertIndexExists(t, ctx, pool, "idx_sessions_expires_at")
			},
		},
		{
			name:        "0005_create_login_states_table",
			fromVersion: 4,
			toVersion:   5,
			before: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertTableNotExists(t, ctx, pool, "login_states")
			},
			after: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertTableExists(t, ctx, pool, "login_states")
				assertColumnExists(t, ctx, pool, "login_states", "id", sptr("uuid"))
				assertColumnExists(t, ctx, pool, "login_states", "redirect_uri", sptr("text"))
				assertColumnExists(t, ctx, pool, "login_states", "code_verifier", sptr("text"))
				assertColumnExists(t, ctx, pool, "login_states", "expires_at", sptr("timestamp with time zone"))
				assertIndexExists(t, ctx, pool, "idx_login_states_expires_at")
			},
		},
	}

	for _, tc := range tests {
		tcase := tc // capture
		t.Run(tcase.name, func(t *testing.T) {
			t.Parallel()

			pool, dsn := testutils.StartPostgresContainer(ctx, t)

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
