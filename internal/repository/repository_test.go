package repository_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/repository/memory"
	"github.com/stretchr/testify/require"
)

func seedTags(
	t *testing.T,
	ctx context.Context,
	repo repository.TagRepository,
	n int,
) []uuid.UUID {
	t.Helper()

	ids := make([]uuid.UUID, n)
	for i := range n {
		tag := model.Tag{
			Name:  fmt.Sprintf("tag-%d", i),
			Color: "#123456",
		}
		created, err := repo.CreateTag(ctx, tag)
		require.NoError(t, err)
		ids[i] = created.Id
	}
	return ids
}

func TestTagRepositoryContract(t *testing.T) {
	ctx := context.Background()

	run := func(t *testing.T, newRepo func(t *testing.T) repository.TagRepository) {
		t.Run("CreateAndGet", func(t *testing.T) {
			repo := newRepo(t)

			tag := model.Tag{
				Name:  "work",
				Color: "#ff0000",
			}

			created, err := repo.CreateTag(ctx, tag)
			require.NoError(t, err)
			require.NotEqual(t, tag.Id, created.Id)

			got, err := repo.GetTag(ctx, created.Id)
			require.NoError(t, err)
			require.Equal(t, "work", got.Name)
		})

		t.Run("GetMissing", func(t *testing.T) {
			repo := newRepo(t)

			_, err := repo.GetTag(ctx, uuid.New())
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run("List", func(t *testing.T) {
			repo := newRepo(t)

			_, _ = repo.CreateTag(ctx, model.Tag{Name: "a", Color: "#ffffff"})
			_, _ = repo.CreateTag(ctx, model.Tag{Name: "b", Color: "#000000"})

			tags, err := repo.ListTags(ctx)
			require.NoError(t, err)
			require.Len(t, tags, 2)
		})

		t.Run("Update", func(t *testing.T) {
			repo := newRepo(t)

			tag := model.Tag{Name: "old", Color: "#00ff00"}
			created, _ := repo.CreateTag(ctx, tag)

			created.Name = "new"
			updated, err := repo.UpdateTag(ctx, created)
			require.NoError(t, err)
			require.Equal(t, "new", updated.Name)
			require.Equal(t, tag.Color, updated.Color)
			require.Equal(t, created.Id, updated.Id)

			got, _ := repo.GetTag(ctx, created.Id)
			require.Equal(t, "new", got.Name)
		})

		t.Run("UpdateMissing", func(t *testing.T) {
			repo := newRepo(t)

			_, err := repo.UpdateTag(ctx, model.Tag{
				Name:  "ghost",
				Color: "#000000",
			})

			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run("Delete", func(t *testing.T) {
			repo := newRepo(t)

			tag := model.Tag{Name: "tmp", Color: "#000"}
			created, _ := repo.CreateTag(ctx, tag)

			err := repo.DeleteTag(ctx, created.Id)
			require.NoError(t, err)

			_, err = repo.GetTag(ctx, created.Id)
			require.ErrorIs(t, err, model.ErrNotFound)
		})
	}

	run(t, func(t *testing.T) repository.TagRepository {
		return memory.NewMemoryStore()
	})

	// Later:
	// run(t, func(t *testing.T) repository.TagRepository {
	//     return newPostgresTagRepo(t)
	// })
}

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

	run(t, func(t *testing.T) repository.Repository {
		return memory.NewMemoryStore()
	})
}

func TestTimeSpanRepositoryContract(t *testing.T) {
	ctx := context.Background()

	run := func(
		t *testing.T,
		newRepo func(t *testing.T) repository.Repository,
	) {
		t.Run("CreateAndGet", func(t *testing.T) {
			repo := newRepo(t)

			tagIds := seedTags(t, ctx, repo, 2)

			start := time.Now().Add(-time.Hour)
			end := time.Now()

			ts := model.TimeSpan{
				Name:      "Work session",
				StartTime: start,
				EndTime:   end,
				TagIds:    tagIds,
			}

			created, err := repo.CreateTimeSpan(ctx, ts)
			ts.Id = created.Id
			require.NoError(t, err)

			got, err := repo.GetTimeSpan(ctx, ts.Id)
			require.NoError(t, err)
			require.Equal(t, "Work session", got.Name)
			require.True(t, got.StartTime.Equal(start))
			require.True(t, got.EndTime.Equal(end))
			require.ElementsMatch(t, tagIds, got.TagIds)
		})

		t.Run("CreateFailsIfTagMissing", func(t *testing.T) {
			repo := newRepo(t)

			ts := model.TimeSpan{
				Name:      "Invalid",
				StartTime: time.Now(),
				EndTime:   time.Now().Add(time.Hour),
				TagIds:    []uuid.UUID{uuid.New()},
			}

			_, err := repo.CreateTimeSpan(ctx, ts)
			require.ErrorIs(t, err, model.ErrInvalidReference)
		})

		t.Run("UpdateFailsIfTagMissing", func(t *testing.T) {
			repo := newRepo(t)

			ts := model.TimeSpan{
				Name:      "Valid",
				StartTime: time.Now(),
				EndTime:   time.Now().Add(time.Hour),
			}
			created, _ := repo.CreateTimeSpan(ctx, ts)
			ts.Id = created.Id

			ts.TagIds = []uuid.UUID{uuid.New()}

			_, err := repo.UpdateTimeSpan(ctx, ts)
			require.ErrorIs(t, err, model.ErrInvalidReference)
		})

		t.Run("Delete", func(t *testing.T) {
			repo := newRepo(t)

			ts := model.TimeSpan{
				Name:      "Temp",
				StartTime: time.Now(),
				EndTime:   time.Now().Add(time.Hour),
			}
			created, _ := repo.CreateTimeSpan(ctx, ts)
			ts.Id = created.Id

			err := repo.DeleteTimeSpan(ctx, ts.Id)
			require.NoError(t, err)

			_, err = repo.GetTimeSpan(ctx, ts.Id)
			require.ErrorIs(t, err, model.ErrNotFound)
		})
	}

	run(t, func(t *testing.T) repository.Repository {
		return memory.NewMemoryStore()
	})
}
