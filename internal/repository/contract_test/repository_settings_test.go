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

func TestSettingsRepositoryContract(t *testing.T) {
	ctx := context.Background()

	run := func(t *testing.T, repoName string, newRepo func(t *testing.T) repository.Repository) {
		t.Run(repoName+"CreateAndGet", func(t *testing.T) {
			repo := newRepo(t)

			created, err := repo.CreateSettings(ctx, testScope, model.DefaultSettings())
			require.NoError(t, err)
			require.NotEqual(t, uuid.Nil, created.Id)
			require.NotNil(t, created.UserId)
			require.Equal(t, *testScope.UserID(), *created.UserId)
			require.Equal(t, model.DefaultSettings().WeekStartDay, created.WeekStartDay)

			got, err := repo.GetSettings(ctx, testScope)
			require.NoError(t, err)
			require.Equal(t, created, got)
		})

		t.Run(repoName+"UnownedCreateHasNilUserId", func(t *testing.T) {
			repo := newRepo(t)

			created, err := repo.CreateSettings(ctx, model.UnownedScope(), model.DefaultSettings())
			require.NoError(t, err)
			require.Nil(t, created.UserId)

			got, err := repo.GetSettings(ctx, model.UnownedScope())
			require.NoError(t, err)
			require.Nil(t, got.UserId)
		})

		t.Run(repoName+"GetMissing", func(t *testing.T) {
			repo := newRepo(t)

			_, err := repo.GetSettings(ctx, testScope)
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"CreateTwiceForSameScopeFails", func(t *testing.T) {
			repo := newRepo(t)

			_, err := repo.CreateSettings(ctx, testScope, model.DefaultSettings())
			require.NoError(t, err)

			_, err = repo.CreateSettings(ctx, testScope, model.DefaultSettings())
			require.ErrorIs(t, err, model.ErrAlreadyExists)
		})

		t.Run(repoName+"Update", func(t *testing.T) {
			repo := newRepo(t)

			created, err := repo.CreateSettings(ctx, testScope, model.DefaultSettings())
			require.NoError(t, err)

			created.WeekStartDay = model.WeekStartSunday
			updated, err := repo.UpdateSettings(ctx, testScope, created)
			require.NoError(t, err)
			require.Equal(t, model.WeekStartSunday, updated.WeekStartDay)
			require.Equal(t, created.Id, updated.Id)

			got, err := repo.GetSettings(ctx, testScope)
			require.NoError(t, err)
			require.Equal(t, model.WeekStartSunday, got.WeekStartDay)
		})

		t.Run(repoName+"UpdateMissing", func(t *testing.T) {
			repo := newRepo(t)

			_, err := repo.UpdateSettings(ctx, testScope, model.DefaultSettings())
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"ScopeIsolation", func(t *testing.T) {
			repo := newRepo(t)
			scopeA := model.UserScope(uuid.New())
			scopeB := model.UserScope(uuid.New())
			seedScopeUser(t, ctx, repo, scopeA)
			seedScopeUser(t, ctx, repo, scopeB)

			settingsA, err := repo.CreateSettings(ctx, scopeA, model.DefaultSettings())
			require.NoError(t, err)
			settingsB, err := repo.CreateSettings(ctx, scopeB, model.DefaultSettings())
			require.NoError(t, err)

			// Get is scoped
			gotA, err := repo.GetSettings(ctx, scopeA)
			require.NoError(t, err)
			require.Equal(t, settingsA.Id, gotA.Id)

			// There is no id to target another scope's row through - Update
			// always applies to the caller's own scope, so this can only ever
			// touch scope B's row.
			settingsB.WeekStartDay = model.WeekStartSunday
			_, err = repo.UpdateSettings(ctx, scopeB, settingsB)
			require.NoError(t, err)

			gotA, err = repo.GetSettings(ctx, scopeA)
			require.NoError(t, err)
			require.Equal(t, model.WeekStartMonday, gotA.WeekStartDay, "scope A's settings must be unaffected by scope B's update")
		})

		t.Run(repoName+"UnownedScopeIsolation", func(t *testing.T) {
			repo := newRepo(t)
			user := model.UserScope(uuid.New())
			seedScopeUser(t, ctx, repo, user)

			owned, err := repo.CreateSettings(ctx, user, model.DefaultSettings())
			require.NoError(t, err)
			unowned, err := repo.CreateSettings(ctx, model.UnownedScope(), model.DefaultSettings())
			require.NoError(t, err)

			gotUnowned, err := repo.GetSettings(ctx, model.UnownedScope())
			require.NoError(t, err)
			require.Equal(t, unowned.Id, gotUnowned.Id)

			gotOwned, err := repo.GetSettings(ctx, user)
			require.NoError(t, err)
			require.Equal(t, owned.Id, gotOwned.Id)
			require.NotEqual(t, owned.Id, unowned.Id)
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
