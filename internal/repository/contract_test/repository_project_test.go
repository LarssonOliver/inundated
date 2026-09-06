package contract_test

import (
	"context"
	"fmt"
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
		repoName string,
		newRepo func(t *testing.T) repository.Repository,
	) {
		t.Run(repoName+"CreateAndGet", func(t *testing.T) {
			repo := newRepo(t)

			tagIds := seedTags(t, ctx, repo, testScope, 2)

			budget := time.Hour

			project := model.Project{
				Name:       "Project A",
				Color:      "#ff0000",
				TimeBudget: &budget,
				TagIds:     tagIds,
			}

			created, err := repo.CreateProject(ctx, testScope, project)
			require.NoError(t, err)
			require.NotEqual(t, project.Id, created.Id)

			got, err := repo.GetProject(ctx, testScope, created.Id)
			require.NoError(t, err)
			require.Equal(t, "Project A", got.Name)
			require.Equal(t, "#ff0000", got.Color)
			require.ElementsMatch(t, tagIds, got.TagIds)
			require.NotNil(t, got.TimeBudget)
			require.Equal(t, budget, *got.TimeBudget)
		})

		t.Run(repoName+"CreateFailsIfTagMissing", func(t *testing.T) {
			repo := newRepo(t)

			project := model.Project{
				Name:   "Invalid",
				Color:  "#00ff00",
				TagIds: []uuid.UUID{uuid.New()},
			}

			_, err := repo.CreateProject(ctx, testScope, project)
			require.ErrorIs(t, err, model.ErrInvalidReference)
		})

		t.Run(repoName+"List", func(t *testing.T) {
			repo := newRepo(t)

			_, _ = repo.CreateProject(ctx, testScope, model.Project{Name: "a", Color: "#ffffff"})
			_, _ = repo.CreateProject(ctx, testScope, model.Project{Name: "b", Color: "#000000"})

			page, err := repo.ListProjects(ctx, testScope, model.DefaultPaginationParams())
			require.NoError(t, err)
			require.Len(t, page.Data, 2)
			require.Equal(t, 2, page.TotalCount)
		})

		t.Run(repoName+"Update", func(t *testing.T) {
			repo := newRepo(t)

			initialTags := seedTags(t, ctx, repo, testScope, 1)
			newTags := seedTags(t, ctx, repo, testScope, 2)

			project := model.Project{
				Name:   "Initial",
				Color:  "#00ff00",
				TagIds: initialTags,
			}
			created, _ := repo.CreateProject(ctx, testScope, project)

			project.Id = created.Id
			project.Name = "Updated"
			project.TagIds = newTags

			updated, err := repo.UpdateProject(ctx, testScope, project)
			require.NoError(t, err)
			require.Equal(t, "Updated", updated.Name)
			require.ElementsMatch(t, newTags, updated.TagIds)
		})

		t.Run(repoName+"UpdateFailsIfTagMissing", func(t *testing.T) {
			repo := newRepo(t)

			project := model.Project{
				Id:    uuid.New(),
				Color: "#0000ff",
				Name:  "P",
			}
			created, _ := repo.CreateProject(ctx, testScope, project)

			project.Id = created.Id
			project.TagIds = []uuid.UUID{uuid.New()}

			_, err := repo.UpdateProject(ctx, testScope, project)
			require.ErrorIs(t, err, model.ErrInvalidReference)
		})

		t.Run(repoName+"Delete", func(t *testing.T) {
			repo := newRepo(t)

			project := model.Project{
				Color: "#0000ff",
				Name:  "Temp",
			}
			created, _ := repo.CreateProject(ctx, testScope, project)
			project.Id = created.Id

			err := repo.DeleteProject(ctx, testScope, project.Id)
			require.NoError(t, err)

			_, err = repo.GetProject(ctx, testScope, project.Id)
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"ListPagination_OffsetAndLimit", func(t *testing.T) {
			repo := newRepo(t)

			for i := range 5 {
				_, _ = repo.CreateProject(ctx, testScope, model.Project{
					Name:  fmt.Sprintf("project-%d", i),
					Color: "#000000",
				})
			}

			page, err := repo.ListProjects(ctx, testScope, model.PaginationParams{Limit: 2, Offset: 0})
			require.NoError(t, err)
			require.Len(t, page.Data, 2)
			require.Equal(t, 5, page.TotalCount)

			page2, err := repo.ListProjects(ctx, testScope, model.PaginationParams{Limit: 2, Offset: 2})
			require.NoError(t, err)
			require.Len(t, page2.Data, 2)
			require.Equal(t, 5, page2.TotalCount)

			ids1 := make(map[uuid.UUID]bool)
			for _, p := range page.Data {
				ids1[p.Id] = true
			}
			for _, p := range page2.Data {
				require.False(t, ids1[p.Id], "duplicate item across pages")
			}
		})

		t.Run(repoName+"ListPagination_OffsetBeyondEnd", func(t *testing.T) {
			repo := newRepo(t)

			_, _ = repo.CreateProject(ctx, testScope, model.Project{Name: "only", Color: "#000000"})

			page, err := repo.ListProjects(ctx, testScope, model.PaginationParams{Limit: 10, Offset: 100})
			require.NoError(t, err)
			require.Empty(t, page.Data)
			require.Equal(t, 1, page.TotalCount)
		})

		t.Run(repoName+"ListPagination_LastPagePartial", func(t *testing.T) {
			repo := newRepo(t)

			for i := range 5 {
				_, _ = repo.CreateProject(ctx, testScope, model.Project{
					Name:  fmt.Sprintf("project-%d", i),
					Color: "#000000",
				})
			}

			page, err := repo.ListProjects(ctx, testScope, model.PaginationParams{Limit: 3, Offset: 3})
			require.NoError(t, err)
			require.Len(t, page.Data, 2)
			require.Equal(t, 5, page.TotalCount)
		})

		t.Run(repoName+"ListPagination_EmptyStore", func(t *testing.T) {
			repo := newRepo(t)

			page, err := repo.ListProjects(ctx, testScope, model.PaginationParams{Limit: 10, Offset: 0})
			require.NoError(t, err)
			require.Empty(t, page.Data)
			require.Equal(t, 0, page.TotalCount)
		})

		t.Run(repoName+"ListPagination_TotalCountUnaffectedByLimit", func(t *testing.T) {
			repo := newRepo(t)

			for i := range 5 {
				_, _ = repo.CreateProject(ctx, testScope, model.Project{
					Name:  fmt.Sprintf("project-%d", i),
					Color: "#000000",
				})
			}

			page, err := repo.ListProjects(ctx, testScope, model.PaginationParams{Limit: 1, Offset: 0})
			require.NoError(t, err)
			require.Len(t, page.Data, 1)
			require.Equal(t, 5, page.TotalCount)
		})
	}

	// Memory

	run(t, "memory", func(t *testing.T) repository.Repository {
		return memory.NewMemoryStore()
	})

	// Postgres
	run(t, "postgres", func(t *testing.T) repository.Repository {
		t.Parallel()
		pool := testutils.StartPostgresContainerWithMigrationsApplied(ctx, t)
		repo := postgres.NewPostgresStoreFromPool(pool)
		seedScopeUser(t, ctx, repo, testScope)
		return repo
	})
}
