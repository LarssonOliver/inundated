package contract_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository/postgres"
	"github.com/larssonoliver/inundated/test/testutils"
	"github.com/stretchr/testify/require"
)

// Attribution stays right when task_tags, tasks.parent_id or a tag's
// deleted state change through plain SQL, not the store's write paths:
// migration 0016's triggers keep task_effective_tags up to date.
func TestPostgres_TaskEffectiveTagsFollowRawWrites(t *testing.T) {
	ctx := context.Background()
	pool := testutils.StartPostgresContainerWithMigrationsApplied(ctx, t)
	repo := postgres.NewPostgresStoreFromPool(pool)
	seedScopeUser(t, ctx, repo, testScope)

	label, err := repo.CreateTag(ctx, testScope, model.Tag{Name: "client", Color: "#ff0000"})
	require.NoError(t, err)
	parent, err := repo.CreateTask(ctx, testScope, model.Task{Name: "parent", TagIds: []uuid.UUID{label.Id}})
	require.NoError(t, err)
	child, err := repo.CreateTask(ctx, testScope, model.Task{Name: "child", ParentId: &parent.Id})
	require.NoError(t, err)
	start := time.Date(2026, 9, 1, 9, 0, 0, 0, time.UTC)
	_, err = repo.CreateTimespan(ctx, testScope, model.Timespan{
		Name: "work", StartTime: start, EndTime: start.Add(time.Hour), TagIds: []uuid.UUID{child.TagId},
	})
	require.NoError(t, err)

	total := func() time.Duration {
		t.Helper()
		d, err := repo.GetTotalDurationByTags(ctx, testScope, []uuid.UUID{label.Id})
		require.NoError(t, err)
		return d
	}
	exec := func(sql string, args ...any) {
		t.Helper()
		_, err := pool.Exec(ctx, sql, args...)
		require.NoError(t, err)
	}
	require.Equal(t, time.Hour, total(), "child inherits the parent's tag")

	exec(`DELETE FROM task_tags WHERE task_id = $1`, parent.Id)
	require.Zero(t, total(), "removing the parent's tag reaches the child")

	exec(`INSERT INTO task_tags (task_id, tag_id) VALUES ($1, $2)`, parent.Id, label.Id)
	require.Equal(t, time.Hour, total())

	exec(`UPDATE tasks SET parent_id = NULL WHERE id = $1`, child.Id)
	require.Zero(t, total(), "a moved child stops inheriting")

	exec(`UPDATE tasks SET parent_id = $2 WHERE id = $1`, child.Id, parent.Id)
	require.Equal(t, time.Hour, total())

	require.NoError(t, repo.DeleteTag(ctx, testScope, label.Id))
	var rows int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM task_effective_tags WHERE tag_id = $1`, label.Id).Scan(&rows))
	require.Zero(t, rows, "deleting a tag drops its rows")

	exec(`UPDATE tags SET deleted_at = NULL WHERE id = $1`, label.Id)
	require.Equal(t, time.Hour, total(), "restoring a tag brings it back")

	rowsFor := func(tagId uuid.UUID) int {
		t.Helper()
		var n int
		require.NoError(t, pool.QueryRow(ctx,
			`SELECT count(*) FROM task_effective_tags WHERE tag_id = $1`, tagId).Scan(&n))
		return n
	}
	exec(`UPDATE tags SET deleted_at = now() WHERE id = $1`, parent.TagId)
	require.Zero(t, rowsFor(parent.TagId), "deleting a task tag drops its rows")
	exec(`UPDATE tags SET deleted_at = NULL WHERE id = $1`, parent.TagId)
	require.Equal(t, 2, rowsFor(parent.TagId), "restoring a task tag reaches the task and its subtask")
}
