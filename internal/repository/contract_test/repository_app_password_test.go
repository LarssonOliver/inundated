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

func TestAppPasswordRepositoryContract(t *testing.T) {
	ctx := context.Background()

	anAppPassword := func(name string, createdAt time.Time) model.AppPassword {
		return model.AppPassword{Id: uuid.New(), Name: name, CreatedAt: createdAt}
	}

	run := func(t *testing.T, repoName string, newRepo func(t *testing.T) repository.Repository) {
		t.Run(repoName+"CreateAndGetByToken", func(t *testing.T) {
			repo := newRepo(t)
			password := anAppPassword("Phone", time.Now().Add(-time.Minute).UTC())

			created, err := repo.CreateAppPassword(ctx, testScope, password, "tok-1")
			require.NoError(t, err)
			require.Equal(t, password.Id, created.Id)

			got, err := repo.GetAppPasswordByToken(ctx, "tok-1")
			require.NoError(t, err)
			require.Equal(t, password.Id, got.Id)
			require.Equal(t, password.Name, got.Name)
			require.Equal(t, testScope.UserID(), got.UserId)
			require.WithinDuration(t, password.CreatedAt, got.CreatedAt, time.Second)
			require.Nil(t, got.LastUsedAt)
		})

		t.Run(repoName+"IdIsNotTheToken", func(t *testing.T) {
			repo := newRepo(t)
			created, err := repo.CreateAppPassword(ctx, testScope, anAppPassword("Phone", time.Now()), "tok-1")
			require.NoError(t, err)

			_, err = repo.GetAppPasswordByToken(ctx, created.Id.String())
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"ListNewestFirst", func(t *testing.T) {
			repo := newRepo(t)
			now := time.Now().UTC()
			oldest, err := repo.CreateAppPassword(ctx, testScope, anAppPassword("oldest", now.Add(-2*time.Hour)), "tok-1")
			require.NoError(t, err)
			middle, err := repo.CreateAppPassword(ctx, testScope, anAppPassword("middle", now.Add(-time.Hour)), "tok-2")
			require.NoError(t, err)
			newest, err := repo.CreateAppPassword(ctx, testScope, anAppPassword("newest", now), "tok-3")
			require.NoError(t, err)

			listed, err := repo.ListAppPasswords(ctx, testScope)
			require.NoError(t, err)
			require.Len(t, listed, 3)
			require.Equal(t, []uuid.UUID{newest.Id, middle.Id, oldest.Id}, []uuid.UUID{listed[0].Id, listed[1].Id, listed[2].Id})
		})

		t.Run(repoName+"Touch", func(t *testing.T) {
			repo := newRepo(t)
			created, err := repo.CreateAppPassword(ctx, testScope, anAppPassword("Phone", time.Now()), "tok-1")
			require.NoError(t, err)

			usedAt := time.Now().Add(time.Minute).UTC()
			require.NoError(t, repo.TouchAppPassword(ctx, created.Id, usedAt))

			got, err := repo.GetAppPasswordByToken(ctx, "tok-1")
			require.NoError(t, err)
			require.NotNil(t, got.LastUsedAt)
			require.WithinDuration(t, usedAt, *got.LastUsedAt, time.Second)
		})

		t.Run(repoName+"TouchMissing", func(t *testing.T) {
			repo := newRepo(t)

			err := repo.TouchAppPassword(ctx, uuid.New(), time.Now())
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"DeleteThenTokenFails", func(t *testing.T) {
			repo := newRepo(t)
			created, err := repo.CreateAppPassword(ctx, testScope, anAppPassword("Phone", time.Now()), "tok-1")
			require.NoError(t, err)

			require.NoError(t, repo.DeleteAppPassword(ctx, testScope, created.Id))

			_, err = repo.GetAppPasswordByToken(ctx, "tok-1")
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"ScopeIsolation", func(t *testing.T) {
			repo := newRepo(t)
			other := model.UserScope(uuid.New())
			seedScopeUser(t, ctx, repo, other)
			created, err := repo.CreateAppPassword(ctx, testScope, anAppPassword("Phone", time.Now()), "tok-1")
			require.NoError(t, err)

			listed, err := repo.ListAppPasswords(ctx, other)
			require.NoError(t, err)
			require.Empty(t, listed)

			err = repo.DeleteAppPassword(ctx, other, created.Id)
			require.ErrorIs(t, err, model.ErrNotFound)

			_, err = repo.GetAppPasswordByToken(ctx, "tok-1")
			require.NoError(t, err)
		})

		t.Run(repoName+"UnownedScopeIsolation", func(t *testing.T) {
			repo := newRepo(t)
			owned, err := repo.CreateAppPassword(ctx, testScope, anAppPassword("owned", time.Now()), "tok-1")
			require.NoError(t, err)
			unowned, err := repo.CreateAppPassword(ctx, model.UnownedScope(), anAppPassword("unowned", time.Now()), "tok-2")
			require.NoError(t, err)
			require.Nil(t, unowned.UserId)

			listed, err := repo.ListAppPasswords(ctx, model.UnownedScope())
			require.NoError(t, err)
			require.Len(t, listed, 1)
			require.Equal(t, unowned.Id, listed[0].Id)

			listed, err = repo.ListAppPasswords(ctx, testScope)
			require.NoError(t, err)
			require.Len(t, listed, 1)
			require.Equal(t, owned.Id, listed[0].Id)

			require.ErrorIs(t, repo.DeleteAppPassword(ctx, model.UnownedScope(), owned.Id), model.ErrNotFound)
			require.ErrorIs(t, repo.DeleteAppPassword(ctx, testScope, unowned.Id), model.ErrNotFound)
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
