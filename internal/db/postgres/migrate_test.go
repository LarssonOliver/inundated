package postgres_test

import (
	"context"
	"os"
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
	// afterDown, if set, is run after migrating back down to fromVersion, to
	// assert the .down.sql reverses the change.
	afterDown func(t *testing.T, ctx context.Context, pool *pgxpool.Pool)
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

// 0006 adds columns and FKs to an existing deployment. If a partial apply is
// recovered by clearing golang-migrate's dirty flag, the file gets re-run, so
// every statement in it must be safe to apply twice.
func TestMigration0006_ReapplyIsIdempotent(t *testing.T) {
	ctx := context.Background()
	pool, dsn := testutils.StartPostgresContainer(ctx, t)

	require.NoError(t, postgres.ApplyMigrationsUpTo(ctx, dsn, 5))

	raw, err := os.ReadFile("migrations/0006_add_user_id.up.sql")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, string(raw))
	require.NoError(t, err)
	_, err = pool.Exec(ctx, string(raw))
	require.NoError(t, err, "re-running 0006 must not fail on an already-present constraint")

	assertForeignKeyExists(t, ctx, pool, "tags", "tags_user_id_fkey")
	assertForeignKeyExists(t, ctx, pool, "projects", "projects_user_id_fkey")
	assertForeignKeyExists(t, ctx, pool, "timespans", "timespans_user_id_fkey")
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
		{
			name:        "0006_add_user_id",
			fromVersion: 5,
			toVersion:   6,
			before: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertColumnNotExists(t, ctx, pool, "tags", "user_id")
				assertColumnNotExists(t, ctx, pool, "projects", "user_id")
				assertColumnNotExists(t, ctx, pool, "timespans", "user_id")
			},
			after: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertColumnExists(t, ctx, pool, "tags", "user_id", sptr("uuid"))
				assertIndexExists(t, ctx, pool, "idx_tags_user_id")
				assertForeignKeyExists(t, ctx, pool, "tags", "tags_user_id_fkey")

				assertColumnExists(t, ctx, pool, "projects", "user_id", sptr("uuid"))
				assertIndexExists(t, ctx, pool, "idx_projects_user_id")
				assertForeignKeyExists(t, ctx, pool, "projects", "projects_user_id_fkey")

				assertColumnExists(t, ctx, pool, "timespans", "user_id", sptr("uuid"))
				assertIndexExists(t, ctx, pool, "idx_timespans_user_id")
				assertForeignKeyExists(t, ctx, pool, "timespans", "timespans_user_id_fkey")
			},
			afterDown: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertColumnNotExists(t, ctx, pool, "tags", "user_id")
				assertColumnNotExists(t, ctx, pool, "projects", "user_id")
				assertColumnNotExists(t, ctx, pool, "timespans", "user_id")
			},
		},
		{
			name:        "0007_login_state_nonce",
			fromVersion: 6,
			toVersion:   7,
			before: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertColumnNotExists(t, ctx, pool, "login_states", "nonce")
			},
			after: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertColumnExists(t, ctx, pool, "login_states", "nonce", sptr("text"))
			},
			afterDown: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertColumnNotExists(t, ctx, pool, "login_states", "nonce")
			},
		},
		{
			name:        "0008_session_created_at",
			fromVersion: 7,
			toVersion:   8,
			before: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertColumnNotExists(t, ctx, pool, "sessions", "created_at")
			},
			after: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertColumnExists(t, ctx, pool, "sessions", "created_at", sptr("timestamp with time zone"))
			},
			afterDown: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertColumnNotExists(t, ctx, pool, "sessions", "created_at")
			},
		},
		{
			name:        "0009_session_token_hash",
			fromVersion: 8,
			toVersion:   9,
			before: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertColumnNotExists(t, ctx, pool, "sessions", "token_hash")
				assertIndexNotExists(t, ctx, pool, "idx_sessions_token_hash")
			},
			after: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertColumnExists(t, ctx, pool, "sessions", "token_hash", sptr("bytea"))
				assertUniqueIndexExists(t, ctx, pool, "idx_sessions_token_hash")
			},
			afterDown: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertColumnNotExists(t, ctx, pool, "sessions", "token_hash")
				assertIndexNotExists(t, ctx, pool, "idx_sessions_token_hash")
			},
		},
		{
			name:        "0010_archived_at",
			fromVersion: 9,
			toVersion:   10,
			before: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertColumnNotExists(t, ctx, pool, "tags", "archived_at")
				assertColumnNotExists(t, ctx, pool, "projects", "archived_at")
			},
			after: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertColumnExists(t, ctx, pool, "tags", "archived_at", sptr("timestamp with time zone"))
				assertColumnExists(t, ctx, pool, "projects", "archived_at", sptr("timestamp with time zone"))
			},
			afterDown: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertColumnNotExists(t, ctx, pool, "tags", "archived_at")
				assertColumnNotExists(t, ctx, pool, "projects", "archived_at")
			},
		},
		{
			name:        "0011_create_settings_table",
			fromVersion: 10,
			toVersion:   11,
			before: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertTableNotExists(t, ctx, pool, "settings")
			},
			after: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertTableExists(t, ctx, pool, "settings")
				assertColumnExists(t, ctx, pool, "settings", "id", sptr("uuid"))
				assertColumnExists(t, ctx, pool, "settings", "user_id", sptr("uuid"))
				assertColumnExists(t, ctx, pool, "settings", "week_start_day", sptr("text"))
				assertColumnExists(t, ctx, pool, "settings", "timezone", sptr("text"))
				assertColumnExists(t, ctx, pool, "settings", "duration_format", sptr("text"))
				assertColumnExists(t, ctx, pool, "settings", "time_format", sptr("text"))
				assertColumnExists(t, ctx, pool, "settings", "date_format", sptr("text"))
				assertForeignKeyExists(t, ctx, pool, "settings", "settings_user_id_fkey")
				assertUniqueIndexExists(t, ctx, pool, "idx_settings_user_id")
				assertUniqueIndexExists(t, ctx, pool, "idx_settings_unowned_singleton")
			},
			afterDown: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertTableNotExists(t, ctx, pool, "settings")
			},
		},
		{
			name:        "0012_timespans_start_time_index",
			fromVersion: 11,
			toVersion:   12,
			before: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertIndexNotExists(t, ctx, pool, "idx_timespans_user_id_start_time")
			},
			after: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertIndexExists(t, ctx, pool, "idx_timespans_user_id_start_time")
			},
			afterDown: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertIndexNotExists(t, ctx, pool, "idx_timespans_user_id_start_time")
			},
		},
		{
			name:        "0013_create_tasks",
			fromVersion: 12,
			toVersion:   13,
			before: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertTableNotExists(t, ctx, pool, "tasks")
				assertTableNotExists(t, ctx, pool, "task_tags")
			},
			after: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertTableExists(t, ctx, pool, "tasks")
				assertColumnExists(t, ctx, pool, "tasks", "id", sptr("uuid"))
				assertColumnExists(t, ctx, pool, "tasks", "tag_id", sptr("uuid"))
				assertColumnExists(t, ctx, pool, "tasks", "parent_id", sptr("uuid"))
				assertColumnExists(t, ctx, pool, "tasks", "due_date", sptr("date"))
				assertColumnExists(t, ctx, pool, "tasks", "estimate", sptr("interval"))
				assertColumnExists(t, ctx, pool, "tasks", "rank", sptr("text"))
				assertColumnExists(t, ctx, pool, "tasks", "close_reason", sptr("text"))
				assertColumnExists(t, ctx, pool, "tasks", "closed_at", sptr("timestamp with time zone"))
				assertForeignKeyExists(t, ctx, pool, "tasks", "tasks_user_id_fkey")
				assertForeignKeyExists(t, ctx, pool, "tasks", "tasks_tag_id_fkey")
				assertForeignKeyExists(t, ctx, pool, "tasks", "tasks_parent_id_fkey")
				assertIndexExists(t, ctx, pool, "idx_tasks_siblings")

				assertTableExists(t, ctx, pool, "task_tags")
				assertForeignKeyExists(t, ctx, pool, "task_tags", "task_tags_task_id_fkey")
				assertForeignKeyExists(t, ctx, pool, "task_tags", "task_tags_tag_id_fkey")
			},
			afterDown: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertTableNotExists(t, ctx, pool, "tasks")
				assertTableNotExists(t, ctx, pool, "task_tags")
			},
		},
		{
			name:        "0014_task_effective_tags",
			fromVersion: 13,
			toVersion:   14,
			before: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertTableNotExists(t, ctx, pool, "task_effective_tags")

				// A parent task tagged "label" with one subtask, to check the
				// backfill.
				_, err := pool.Exec(ctx, `
					INSERT INTO tags (id, name, color) VALUES
						('00000000-0000-0000-0000-00000000000a', 'label', '#000000'),
						('00000000-0000-0000-0000-00000000000b', 'parent', '#000000'),
						('00000000-0000-0000-0000-00000000000c', 'child', '#000000');
					INSERT INTO tasks (id, tag_id, parent_id, name, rank) VALUES
						('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-00000000000b', NULL, 'parent', 'V'),
						('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-00000000000c', '00000000-0000-0000-0000-000000000001', 'child', 'V');
					INSERT INTO task_tags (task_id, tag_id) VALUES
						('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-00000000000a');`)
				require.NoError(t, err)
			},
			after: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertTableExists(t, ctx, pool, "task_effective_tags")
				assertForeignKeyExists(t, ctx, pool, "task_effective_tags", "task_effective_tags_task_id_fkey")
				assertForeignKeyExists(t, ctx, pool, "task_effective_tags", "task_effective_tags_tag_id_fkey")
				assertIndexExists(t, ctx, pool, "idx_task_effective_tags_tag_id")

				var childTags []string
				require.NoError(t, pool.QueryRow(ctx, `
					SELECT array_agg(tag.name ORDER BY tag.name)
					FROM task_effective_tags te JOIN tags tag ON tag.id = te.tag_id
					WHERE te.task_id = '00000000-0000-0000-0000-000000000002'`).Scan(&childTags))
				require.Equal(t, []string{"child", "label", "parent"}, childTags)
			},
			afterDown: func(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
				assertTableNotExists(t, ctx, pool, "task_effective_tags")
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

			if tcase.afterDown != nil {
				require.NoError(t,
					postgres.ApplyMigrationsUpTo(ctx, dsn, tcase.fromVersion),
					"migrating back down to %d", tcase.fromVersion,
				)
				tcase.afterDown(t, ctx, pool)
			}
		})
	}
}
