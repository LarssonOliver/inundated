package contract_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository/postgres"
	"github.com/larssonoliver/inundated/test/testutils"
	"github.com/stretchr/testify/require"
)

// TestPostgres_ScopedListUsesIndex proves the index-friendly owner predicate
// (`user_id = $1`, not `user_id IS NOT DISTINCT FROM $1`) lets the planner use
// idx_tags_user_id from migration 0006 instead of seq-scanning tags.
func TestPostgres_ScopedListUsesIndex(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pool := testutils.StartPostgresContainerWithMigrationsApplied(ctx, t)
	repo := postgres.NewPostgresStoreFromPool(pool)

	a := model.UserScope(uuid.New())
	b := model.UserScope(uuid.New())
	seedScopeUser(t, ctx, repo, a)
	seedScopeUser(t, ctx, repo, b)
	for i := 0; i < 300; i++ {
		_, err := repo.CreateTag(ctx, a, model.Tag{Name: fmt.Sprintf("a%d", i), Color: "#111111"})
		require.NoError(t, err)
		_, err = repo.CreateTag(ctx, b, model.Tag{Name: fmt.Sprintf("b%d", i), Color: "#222222"})
		require.NoError(t, err)
	}
	_, err := pool.Exec(ctx, "ANALYZE tags")
	require.NoError(t, err)

	var plan string
	err = pool.QueryRow(ctx,
		`EXPLAIN (FORMAT TEXT) SELECT id, name, color, user_id FROM tags
		 WHERE deleted_at IS NULL AND user_id = $1 ORDER BY name LIMIT 25 OFFSET 0`,
		*a.UserID()).Scan(&plan)
	require.NoError(t, err)
	require.NotContains(t, plan, "Seq Scan on tags",
		"scoped list should use idx_tags_user_id, plan was:\n"+plan)
}
