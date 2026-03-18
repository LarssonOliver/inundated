package repository_test

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

func TestProjectRepositoryContract(t *testing.T) {
	ctx := context.Background()

	run := func(
		t *testing.T,
		newRepo func(t *testing.T) repository.Repository,
	) {
		t.Run("CreateAndGet", func(t *testing.T) {
			repo := newRepo(t)

			tagIds := seedTags(t, ctx, repo, 2)

			budget := time.Hour

			project := model.Project{
				Name:       "Project A",
				Color:      "#ff0000",
				TimeBudget: &budget,
				TagIds:     tagIds,
			}

			created, err := repo.CreateProject(ctx, project)
			require.NoError(t, err)
			require.NotEqual(t, project.Id, created.Id)

			got, err := repo.GetProject(ctx, created.Id)
			require.NoError(t, err)
			require.Equal(t, "Project A", got.Name)
			require.Equal(t, "#ff0000", got.Color)
			require.ElementsMatch(t, tagIds, got.TagIds)
			require.NotNil(t, got.TimeBudget)
			require.Equal(t, budget, *got.TimeBudget)
		})

		t.Run("CreateFailsIfTagMissing", func(t *testing.T) {
			repo := newRepo(t)

			project := model.Project{
				Name:   "Invalid",
				Color:  "#00ff00",
				TagIds: []uuid.UUID{uuid.New()},
			}

			_, err := repo.CreateProject(ctx, project)
			require.ErrorIs(t, err, model.ErrInvalidReference)
		})

		t.Run("Update", func(t *testing.T) {
			repo := newRepo(t)

			initialTags := seedTags(t, ctx, repo, 1)
			newTags := seedTags(t, ctx, repo, 2)

			project := model.Project{
				Name:   "Initial",
				Color:  "#00ff00",
				TagIds: initialTags,
			}
			created, _ := repo.CreateProject(ctx, project)

			project.Id = created.Id
			project.Name = "Updated"
			project.TagIds = newTags

			updated, err := repo.UpdateProject(ctx, project)
			require.NoError(t, err)
			require.Equal(t, "Updated", updated.Name)
			require.ElementsMatch(t, newTags, updated.TagIds)
		})

		t.Run("UpdateFailsIfTagMissing", func(t *testing.T) {
			repo := newRepo(t)

			project := model.Project{
				Id:    uuid.New(),
				Color: "#0000ff",
				Name:  "P",
			}
			created, _ := repo.CreateProject(ctx, project)

			project.Id = created.Id
			project.TagIds = []uuid.UUID{uuid.New()}

			_, err := repo.UpdateProject(ctx, project)
			require.ErrorIs(t, err, model.ErrInvalidReference)
		})

		t.Run("Delete", func(t *testing.T) {
			repo := newRepo(t)

			project := model.Project{
				Color: "#0000ff",
				Name:  "Temp",
			}
			created, _ := repo.CreateProject(ctx, project)
			project.Id = created.Id

			err := repo.DeleteProject(ctx, project.Id)
			require.NoError(t, err)

			_, err = repo.GetProject(ctx, project.Id)
			require.ErrorIs(t, err, model.ErrNotFound)
		})
	}

	// Memory

	run(t, func(t *testing.T) repository.Repository {
		return memory.NewMemoryStore()
	})

	// Postgres
	run(t, func(t *testing.T) repository.Repository {
		t.Parallel()
		pool := testutils.StartPostgresContainerWithMigrationsApplied(ctx, t)
		return postgres.NewPostgresStoreFromPool(pool)
	})
}

