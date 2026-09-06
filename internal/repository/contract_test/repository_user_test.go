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

		t.Run(repoName+"CreateUserAdoptingOrphans_FirstUserClaimsExistingResources", func(t *testing.T) {
			repo := newRepo(t)

			seedOrphanResources(t, ctx, repo, 2, 3, 4)

			user := model.User{
				Id:    uuid.New(),
				Sub:   "auth0|first",
				Email: "first@example.com",
				Name:  "First User",
			}

			created, adoption, err := repo.CreateUserAdoptingOrphans(ctx, user)
			require.NoError(t, err)
			require.Equal(t, user.Sub, created.Sub)
			require.Equal(t, model.OrphanAdoption{Projects: 3, Tags: 2, Timespans: 4}, adoption)
		})

		t.Run(repoName+"CreateUserAdoptingOrphans_SecondUserAdoptsNothing", func(t *testing.T) {
			repo := newRepo(t)

			seedOrphanResources(t, ctx, repo, 1, 1, 1)

			_, first, err := repo.CreateUserAdoptingOrphans(ctx, model.User{
				Id: uuid.New(), Sub: "auth0|first", Email: "first@example.com", Name: "First",
			})
			require.NoError(t, err)
			require.Equal(t, 3, first.Total())

			// New resources created after the first user still land unowned
			// (ownership is not threaded through the create path yet), but the
			// second user must not adopt them.
			seedOrphanResources(t, ctx, repo, 1, 1, 1)

			_, second, err := repo.CreateUserAdoptingOrphans(ctx, model.User{
				Id: uuid.New(), Sub: "auth0|second", Email: "second@example.com", Name: "Second",
			})
			require.NoError(t, err)
			require.Equal(t, model.OrphanAdoption{}, second)
		})

		t.Run(repoName+"CreateUserAdoptingOrphans_NoResources", func(t *testing.T) {
			repo := newRepo(t)

			_, adoption, err := repo.CreateUserAdoptingOrphans(ctx, model.User{
				Id: uuid.New(), Sub: "auth0|solo", Email: "solo@example.com", Name: "Solo",
			})
			require.NoError(t, err)
			require.Equal(t, model.OrphanAdoption{}, adoption)
		})

		t.Run(repoName+"CreateUserAdoptingOrphans_DuplicateSub", func(t *testing.T) {
			repo := newRepo(t)

			_, _, err := repo.CreateUserAdoptingOrphans(ctx, model.User{
				Id: uuid.New(), Sub: "auth0|dup", Email: "a@example.com", Name: "A",
			})
			require.NoError(t, err)

			_, _, err = repo.CreateUserAdoptingOrphans(ctx, model.User{
				Id: uuid.New(), Sub: "auth0|dup", Email: "b@example.com", Name: "B",
			})
			require.ErrorIs(t, err, model.ErrAlreadyExists)
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
