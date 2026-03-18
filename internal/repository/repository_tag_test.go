package repository_test

import (
	"context"
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

	run := func(t *testing.T, newRepo func(t *testing.T) repository.Repository) {
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
				Id:    uuid.New(),
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

