package contract_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/repository/memory"
	"github.com/larssonoliver/inundated/internal/repository/postgres"
	"github.com/larssonoliver/inundated/test/testutils"
	"github.com/stretchr/testify/require"
)

// The zero UUID never names a row, so every by-id lookup of a user's
// resources reports it as a miss (a 404), never as a bad request.
func TestNilIdIsNotFoundContract(t *testing.T) {
	ctx := context.Background()

	run := func(t *testing.T, repoName string, newRepo func(t *testing.T) repository.Repository) {
		t.Run(repoName, func(t *testing.T) {
			repo := newRepo(t)
			id := uuid.Nil
			start := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
			calls := map[string]func() error{
				"GetTag": func() error { _, err := repo.GetTag(ctx, testScope, id); return err },
				"UpdateTag": func() error {
					_, err := repo.UpdateTag(ctx, testScope, model.Tag{Id: id, Name: "x", Color: "#000000"})
					return err
				},
				"DeleteTag":  func() error { return repo.DeleteTag(ctx, testScope, id) },
				"GetProject": func() error { _, err := repo.GetProject(ctx, testScope, id); return err },
				"UpdateProject": func() error {
					_, err := repo.UpdateProject(ctx, testScope, model.Project{Id: id, Name: "x", Color: "#000000"})
					return err
				},
				"DeleteProject": func() error { return repo.DeleteProject(ctx, testScope, id) },
				"GetTimespan":   func() error { _, err := repo.GetTimespan(ctx, testScope, id); return err },
				"UpdateTimespan": func() error {
					_, err := repo.UpdateTimespan(ctx, testScope, model.Timespan{Id: id, Name: "x", StartTime: start, EndTime: start.Add(time.Hour)})
					return err
				},
				"DeleteTimespan": func() error { return repo.DeleteTimespan(ctx, testScope, id) },
				"GetTask":        func() error { _, err := repo.GetTask(ctx, testScope, id); return err },
				"UpdateTask": func() error {
					name := "x"
					_, err := repo.UpdateTask(ctx, testScope, id, model.TaskPatch{Name: &name})
					return err
				},
				"MoveTask":   func() error { _, err := repo.MoveTask(ctx, testScope, id, nil, nil); return err },
				"DeleteTask": func() error { return repo.DeleteTask(ctx, testScope, id) },
			}
			for name, call := range calls {
				require.ErrorIs(t, call(), model.ErrNotFound, name)
			}
		})
	}

	run(t, "memory", func(t *testing.T) repository.Repository {
		return memory.NewMemoryStore()
	})

	run(t, "postgres", func(t *testing.T) repository.Repository {
		t.Parallel()
		pool := testutils.StartPostgresContainerWithMigrationsApplied(ctx, t)
		repo := postgres.NewPostgresStoreFromPool(pool)
		seedScopeUser(t, ctx, repo, testScope)
		return repo
	})
}
