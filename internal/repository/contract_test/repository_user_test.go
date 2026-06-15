package contract_test

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

func TestUserRepositoryContract(t *testing.T) {
	ctx := context.Background()

	run := func(t *testing.T, repoName string, newRepo func(t *testing.T) repository.Repository) {
		t.Run(repoName+"CreateAndGetByID", func(t *testing.T) {
			repo := newRepo(t)

			user := &model.User{
				ID:    uuid.New(),
				Sub:   "auth0|user123",
				Email: "user@example.com",
				Name:  "Test User",
			}

			err := repo.Create(ctx, user)
			require.NoError(t, err)

			got, err := repo.GetByID(ctx, user.ID)
			require.NoError(t, err)
			require.Equal(t, user.Sub, got.Sub)
			require.Equal(t, user.Email, got.Email)
			require.Equal(t, user.Name, got.Name)
		})

		t.Run(repoName+"CreateAndGetBySub", func(t *testing.T) {
			repo := newRepo(t)

			user := &model.User{
				ID:    uuid.New(),
				Sub:   "google|user456",
				Email: "user@example.com",
				Name:  "Google User",
			}

			err := repo.Create(ctx, user)
			require.NoError(t, err)

			got, err := repo.GetBySub(ctx, user.Sub)
			require.NoError(t, err)
			require.Equal(t, user.ID, got.ID)
			require.Equal(t, user.Email, got.Email)
			require.Equal(t, user.Name, got.Name)
		})

		t.Run(repoName+"GetByIDMissing", func(t *testing.T) {
			repo := newRepo(t)

			_, err := repo.GetByID(ctx, uuid.New())
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"GetBySubMissing", func(t *testing.T) {
			repo := newRepo(t)

			_, err := repo.GetBySub(ctx, "nonexistent|sub")
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"CreateSetsTimestamps", func(t *testing.T) {
			repo := newRepo(t)

			user := &model.User{
				ID:    uuid.New(),
				Sub:   "auth0|user789",
				Email: "user@example.com",
				Name:  "User",
			}

			err := repo.Create(ctx, user)
			require.NoError(t, err)

			got, err := repo.GetByID(ctx, user.ID)
			require.NoError(t, err)
			require.NotZero(t, got.CreatedAt)
			require.NotZero(t, got.UpdatedAt)
		})

		t.Run(repoName+"Update", func(t *testing.T) {
			repo := newRepo(t)

			user := &model.User{
				ID:    uuid.New(),
				Sub:   "auth0|updatetest",
				Email: "old@example.com",
				Name:  "Old Name",
			}

			err := repo.Create(ctx, user)
			require.NoError(t, err)

			newEmail := "new@example.com"
			newName := "New Name"
			updated, err := repo.Update(ctx, user.ID, &model.UpdateUser{
				Email: &newEmail,
				Name:  &newName,
			})
			require.NoError(t, err)
			require.Equal(t, newEmail, updated.Email)
			require.Equal(t, newName, updated.Name)

			got, _ := repo.GetByID(ctx, user.ID)
			require.Equal(t, newEmail, got.Email)
			require.Equal(t, newName, got.Name)
		})

		t.Run(repoName+"UpdatePartial", func(t *testing.T) {
			repo := newRepo(t)

			user := &model.User{
				ID:    uuid.New(),
				Sub:   "auth0|partialtest",
				Email: "original@example.com",
				Name:  "Original Name",
			}

			err := repo.Create(ctx, user)
			require.NoError(t, err)

			newEmail := "changed@example.com"
			updated, err := repo.Update(ctx, user.ID, &model.UpdateUser{
				Email: &newEmail,
				Name:  nil,
			})
			require.NoError(t, err)
			require.Equal(t, newEmail, updated.Email)
			require.Equal(t, "Original Name", updated.Name) // unchanged

			got, _ := repo.GetByID(ctx, user.ID)
			require.Equal(t, newEmail, got.Email)
			require.Equal(t, "Original Name", got.Name)
		})

		t.Run(repoName+"UpdateMissing", func(t *testing.T) {
			repo := newRepo(t)

			email := "ghost@example.com"
			_, err := repo.Update(ctx, uuid.New(), &model.UpdateUser{
				Email: &email,
			})
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"SubIsUnique", func(t *testing.T) {
			repo := newRepo(t)

			user1 := &model.User{
				ID:    uuid.New(),
				Sub:   "auth0|unique",
				Email: "user1@example.com",
				Name:  "User 1",
			}
			err := repo.Create(ctx, user1)
			require.NoError(t, err)

			user2 := &model.User{
				ID:    uuid.New(),
				Sub:   "auth0|unique", // duplicate sub
				Email: "user2@example.com",
				Name:  "User 2",
			}
			err = repo.Create(ctx, user2)
			require.Error(t, err) // should fail due to unique constraint
		})

		t.Run(repoName+"UpdatePreservesCreatedAt", func(t *testing.T) {
			repo := newRepo(t)

			user := &model.User{
				ID:    uuid.New(),
				Sub:   "auth0|preserve",
				Email: "user@example.com",
				Name:  "User",
			}

			err := repo.Create(ctx, user)
			require.NoError(t, err)

			createdUser, _ := repo.GetByID(ctx, user.ID)
			originalCreatedAt := createdUser.CreatedAt

			newName := "Updated"
			_, err = repo.Update(ctx, user.ID, &model.UpdateUser{
				Name: &newName,
			})
			require.NoError(t, err)

			updated, _ := repo.GetByID(ctx, user.ID)
			require.Equal(t, originalCreatedAt, updated.CreatedAt)
			require.NotEqual(t, updated.UpdatedAt, originalCreatedAt) // UpdatedAt should change
		})

		t.Run(repoName+"EmptyNameAllowed", func(t *testing.T) {
			repo := newRepo(t)

			user := &model.User{
				ID:    uuid.New(),
				Sub:   "auth0|noname",
				Email: "user@example.com",
				Name:  "", // empty name
			}

			err := repo.Create(ctx, user)
			require.NoError(t, err)

			got, _ := repo.GetByID(ctx, user.ID)
			require.Equal(t, "", got.Name)
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
