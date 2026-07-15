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

			user := model.User{
				Id:    uuid.New(),
				Sub:   "auth0|user123",
				Email: "user@example.com",
				Name:  "Test User",
			}

			got, err := repo.CreateUser(ctx, user)
			require.NoError(t, err)
			require.Equal(t, user.Sub, got.Sub)
			require.Equal(t, user.Email, got.Email)
			require.Equal(t, user.Name, got.Name)

			got, err = repo.GetUser(ctx, user.Id)
			require.NoError(t, err)
			require.Equal(t, user.Sub, got.Sub)
			require.Equal(t, user.Email, got.Email)
			require.Equal(t, user.Name, got.Name)
		})

		t.Run(repoName+"CreateAndGetBySub", func(t *testing.T) {
			repo := newRepo(t)

			user := model.User{
				Id:    uuid.New(),
				Sub:   "google|user456",
				Email: "user@example.com",
				Name:  "Google User",
			}

			_, err := repo.CreateUser(ctx, user)
			require.NoError(t, err)

			got, err := repo.GetUserBySub(ctx, user.Sub)
			require.NoError(t, err)
			require.Equal(t, user.Id, got.Id)
			require.Equal(t, user.Email, got.Email)
			require.Equal(t, user.Name, got.Name)
		})

		t.Run(repoName+"GetByIDMissing", func(t *testing.T) {
			repo := newRepo(t)

			_, err := repo.GetUser(ctx, uuid.New())
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"GetBySubMissing", func(t *testing.T) {
			repo := newRepo(t)

			_, err := repo.GetUserBySub(ctx, "nonexistent|sub")
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"Update", func(t *testing.T) {
			repo := newRepo(t)

			user := model.User{
				Id:    uuid.New(),
				Sub:   "auth0|updatetest",
				Email: "old@example.com",
				Name:  "Old Name",
			}

			_, err := repo.CreateUser(ctx, user)
			require.NoError(t, err)

			newEmail := "new@example.com"
			newName := "New Name"
			updated, err := repo.UpdateUser(ctx, model.User{
				Id:    user.Id,
				Email: newEmail,
				Name:  newName,
			})
			require.NoError(t, err)
			require.Equal(t, newEmail, updated.Email)
			require.Equal(t, newName, updated.Name)

			got, _ := repo.GetUser(ctx, user.Id)
			require.Equal(t, newEmail, got.Email)
			require.Equal(t, newName, got.Name)
		})

		t.Run(repoName+"UpdateMissing", func(t *testing.T) {
			repo := newRepo(t)

			email := "ghost@example.com"
			_, err := repo.UpdateUser(ctx, model.User{
				Id:    uuid.New(),
				Email: email,
			})
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"SubIsUnique", func(t *testing.T) {
			repo := newRepo(t)

			user1 := model.User{
				Id:    uuid.New(),
				Sub:   "auth0|unique",
				Email: "user1@example.com",
				Name:  "User 1",
			}
			_, err := repo.CreateUser(ctx, user1)
			require.NoError(t, err)

			user2 := model.User{
				Id:    uuid.New(),
				Sub:   "auth0|unique", // duplicate sub
				Email: "user2@example.com",
				Name:  "User 2",
			}
			_, err = repo.CreateUser(ctx, user2)
			require.ErrorIs(t, err, model.ErrAlreadyExists)
		})

		t.Run(repoName+"EmptyNameAllowed", func(t *testing.T) {
			repo := newRepo(t)

			user := model.User{
				Id:    uuid.New(),
				Sub:   "auth0|noname",
				Email: "user@example.com",
				Name:  "", // empty name
			}

			_, err := repo.CreateUser(ctx, user)
			require.NoError(t, err)

			got, _ := repo.GetUser(ctx, user.Id)
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
