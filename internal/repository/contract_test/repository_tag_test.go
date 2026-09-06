package contract_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/repository/memory"
	"github.com/larssonoliver/inundated/internal/repository/postgres"
	"github.com/larssonoliver/inundated/test/testutils"
	"github.com/stretchr/testify/require"
)

func TestTagRepositoryContract(t *testing.T) {
	ctx := context.Background()

	run := func(t *testing.T, repoName string, newRepo func(t *testing.T) repository.Repository) {
		t.Run(repoName+"CreateAndGet", func(t *testing.T) {
			repo := newRepo(t)

			tag := model.Tag{
				Name:  "work",
				Color: "#ff0000",
			}

			created, err := repo.CreateTag(ctx, testScope, tag)
			require.NoError(t, err)
			require.NotEqual(t, tag.Id, created.Id)

			got, err := repo.GetTag(ctx, testScope, created.Id)
			require.NoError(t, err)
			require.Equal(t, "work", got.Name)
		})

		t.Run(repoName+"GetMissing", func(t *testing.T) {
			repo := newRepo(t)

			_, err := repo.GetTag(ctx, testScope, uuid.New())
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"List", func(t *testing.T) {
			repo := newRepo(t)

			_, _ = repo.CreateTag(ctx, testScope, model.Tag{Name: "a", Color: "#ffffff"})
			_, _ = repo.CreateTag(ctx, testScope, model.Tag{Name: "b", Color: "#000000"})

			page, err := repo.ListTags(ctx, testScope, model.DefaultPaginationParams())
			require.NoError(t, err)
			require.Len(t, page.Data, 2)
			require.Equal(t, 2, page.TotalCount)
		})

		t.Run(repoName+"Update", func(t *testing.T) {
			repo := newRepo(t)

			tag := model.Tag{Name: "old", Color: "#00ff00"}
			created, _ := repo.CreateTag(ctx, testScope, tag)

			created.Name = "new"
			updated, err := repo.UpdateTag(ctx, testScope, created)
			require.NoError(t, err)
			require.Equal(t, "new", updated.Name)
			require.Equal(t, tag.Color, updated.Color)
			require.Equal(t, created.Id, updated.Id)

			got, _ := repo.GetTag(ctx, testScope, created.Id)
			require.Equal(t, "new", got.Name)
		})

		t.Run(repoName+"UpdateMissing", func(t *testing.T) {
			repo := newRepo(t)

			_, err := repo.UpdateTag(ctx, testScope, model.Tag{
				Id:    uuid.New(),
				Name:  "ghost",
				Color: "#000000",
			})

			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"Delete", func(t *testing.T) {
			repo := newRepo(t)

			tag := model.Tag{Name: "tmp", Color: "#000"}
			created, _ := repo.CreateTag(ctx, testScope, tag)

			err := repo.DeleteTag(ctx, testScope, created.Id)
			require.NoError(t, err)

			_, err = repo.GetTag(ctx, testScope, created.Id)
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"ListPagination_OffsetAndLimit", func(t *testing.T) {
			repo := newRepo(t)

			for i := range 5 {
				_, _ = repo.CreateTag(ctx, testScope, model.Tag{
					Name:  fmt.Sprintf("tag-%d", i),
					Color: "#000000",
				})
			}

			page, err := repo.ListTags(ctx, testScope, model.PaginationParams{Limit: 2, Offset: 0})
			require.NoError(t, err)
			assertPage(t, page, 2, 5)

			page2, err := repo.ListTags(ctx, testScope, model.PaginationParams{Limit: 2, Offset: 2})
			require.NoError(t, err)
			assertPage(t, page2, 2, 5)

			//.Data on page 1 and page 2 must be disjoint.
			// (requires Tags to have comparable identity — use Id)
			ids1 := make(map[uuid.UUID]bool)
			for _, tag := range page.Data {
				ids1[tag.Id] = true
			}
			for _, tag := range page2.Data {
				require.False(t, ids1[tag.Id], "duplicate item across pages")
			}
		})

		t.Run(repoName+"ListPagination_OffsetBeyondEnd", func(t *testing.T) {
			repo := newRepo(t)

			_, _ = repo.CreateTag(ctx, testScope, model.Tag{Name: "only", Color: "#000000"})

			page, err := repo.ListTags(ctx, testScope, model.PaginationParams{Limit: 10, Offset: 100})
			require.NoError(t, err)
			assertPage(t, page, 0, 1) // empty items, but TotalCount still reflects reality
		})

		t.Run(repoName+"ListPagination_LastPagePartial", func(t *testing.T) {
			repo := newRepo(t)

			for i := range 5 {
				_, _ = repo.CreateTag(ctx, testScope, model.Tag{
					Name:  fmt.Sprintf("tag-%d", i),
					Color: "#000000",
				})
			}

			page, err := repo.ListTags(ctx, testScope, model.PaginationParams{Limit: 3, Offset: 3})
			require.NoError(t, err)
			assertPage(t, page, 2, 5) // only 2 items remain
		})

		t.Run(repoName+"ListPagination_EmptyStore", func(t *testing.T) {
			repo := newRepo(t)

			page, err := repo.ListTags(ctx, testScope, model.PaginationParams{Limit: 10, Offset: 0})
			require.NoError(t, err)
			assertPage(t, page, 0, 0)
		})

		t.Run(repoName+"ListPagination_TotalCountUnaffectedByLimit", func(t *testing.T) {
			repo := newRepo(t)

			for i := range 5 {
				_, _ = repo.CreateTag(ctx, testScope, model.Tag{
					Name:  fmt.Sprintf("tag-%d", i),
					Color: "#000000",
				})
			}

			page, err := repo.ListTags(ctx, testScope, model.PaginationParams{Limit: 1, Offset: 0})
			require.NoError(t, err)
			require.Equal(t, 5, page.TotalCount) // TotalCount is always the full count
			require.Len(t, page.Data, 1)
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
		return postgres.NewPostgresStoreFromPool(pool)
	})
}
