package postgres_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func assertTableExists(t *testing.T, ctx context.Context, pool *pgxpool.Pool, table string) {
	t.Helper()

	var exists bool
	err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public'
			  AND table_name = $1
		)
	`, table).Scan(&exists)

	require.NoError(t, err)
	require.True(t, exists, "expected table %q to exist", table)
}

func assertTableNotExists(t *testing.T, ctx context.Context, pool *pgxpool.Pool, table string) {
	t.Helper()

	var exists bool
	err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public'
			  AND table_name = $1
		)
	`, table).Scan(&exists)

	require.NoError(t, err)
	require.False(t, exists, "expected table %q to NOT exist", table)
}

func assertColumnExists(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	table, column string,
	expectedType *string, // nil = don't care
) {
	t.Helper()

	var dataType string
	err := pool.QueryRow(ctx, `
		SELECT data_type
		FROM information_schema.columns
		WHERE table_schema = 'public'
		  AND table_name = $1
		  AND column_name = $2
	`, table, column).Scan(&dataType)

	require.NoError(t, err, "expected column %s.%s to exist", table, column)

	if expectedType != nil {
		require.Equal(t, *expectedType, dataType,
			"unexpected type for %s.%s", table, column)
	}
}

func assertIndexExists(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	indexName string,
) {
	t.Helper()

	var exists bool
	err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM pg_indexes
			WHERE schemaname = 'public'
			  AND indexname = $1
		)
	`, indexName).Scan(&exists)

	require.NoError(t, err)
	require.True(t, exists, "expected index %q to exist", indexName)
}

func assertForeignKeyExists(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	table, constraint string,
) {
	t.Helper()

	var exists bool
	err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.table_constraints
			WHERE table_schema = 'public'
			  AND table_name = $1
			  AND constraint_name = $2
			  AND constraint_type = 'FOREIGN KEY'
		)
	`, table, constraint).Scan(&exists)

	require.NoError(t, err)
	require.True(t, exists,
		"expected foreign key %q on table %q to exist", constraint, table)
}
