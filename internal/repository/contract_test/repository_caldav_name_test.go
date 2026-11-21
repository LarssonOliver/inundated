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

func TestCaldavNameRepositoryContract(t *testing.T) {
	ctx := context.Background()
	start := time.Date(2026, 11, 21, 9, 0, 0, 0, time.UTC)
	newTimespan := func(name string) model.Timespan {
		return model.Timespan{StartTime: start, EndTime: start.Add(time.Hour), CaldavName: new(name)}
	}

	run := func(t *testing.T, repoName string, newRepo func(t *testing.T) repository.Repository) {
		t.Run(repoName+"TimespanCreateAndFindByName", func(t *testing.T) {
			repo := newRepo(t)
			created, err := repo.CreateTimespan(ctx, testScope, newTimespan("ABC.ics"))
			require.NoError(t, err)
			require.Equal(t, "ABC.ics", *created.CaldavName)

			found, err := repo.GetTimespanByCaldavName(ctx, testScope, "ABC.ics")
			require.NoError(t, err)
			require.Equal(t, created.Id, found.Id)
			require.Equal(t, "ABC.ics", *found.CaldavName)

			got, err := repo.GetTimespan(ctx, testScope, created.Id)
			require.NoError(t, err)
			require.Equal(t, "ABC.ics", *got.CaldavName)

			page, err := repo.ListTimespans(ctx, testScope, model.TimespanListParams{PaginationParams: model.DefaultPaginationParams()})
			require.NoError(t, err)
			require.Len(t, page.Data, 1)
			require.Equal(t, "ABC.ics", *page.Data[0].CaldavName)

			_, err = repo.GetTimespanByCaldavName(ctx, testScope, "missing.ics")
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"TimespanFindByNameIsScoped", func(t *testing.T) {
			repo := newRepo(t)
			other := model.UserScope(uuid.MustParse("22222222-2222-2222-2222-222222222222"))
			seedScopeUser(t, ctx, repo, other)
			_, err := repo.CreateTimespan(ctx, testScope, newTimespan("ABC.ics"))
			require.NoError(t, err)

			_, err = repo.GetTimespanByCaldavName(ctx, other, "ABC.ics")
			require.ErrorIs(t, err, model.ErrNotFound)

			// The same name is free in another scope.
			_, err = repo.CreateTimespan(ctx, other, newTimespan("ABC.ics"))
			require.NoError(t, err)
		})

		t.Run(repoName+"TimespanUpdateKeepsName", func(t *testing.T) {
			repo := newRepo(t)
			created, err := repo.CreateTimespan(ctx, testScope, newTimespan("ABC.ics"))
			require.NoError(t, err)

			_, err = repo.UpdateTimespan(ctx, testScope, model.Timespan{Id: created.Id, Name: "renamed", StartTime: start, EndTime: start.Add(2 * time.Hour)})
			require.NoError(t, err)

			got, err := repo.GetTimespan(ctx, testScope, created.Id)
			require.NoError(t, err)
			require.Equal(t, "renamed", got.Name)
			require.NotNil(t, got.CaldavName)
			require.Equal(t, "ABC.ics", *got.CaldavName)
		})

		t.Run(repoName+"TimespanDuplicateName", func(t *testing.T) {
			repo := newRepo(t)
			first, err := repo.CreateTimespan(ctx, testScope, newTimespan("ABC.ics"))
			require.NoError(t, err)

			_, err = repo.CreateTimespan(ctx, testScope, newTimespan("ABC.ics"))
			require.ErrorIs(t, err, model.ErrAlreadyExists)

			require.NoError(t, repo.DeleteTimespan(ctx, testScope, first.Id))
			_, err = repo.CreateTimespan(ctx, testScope, newTimespan("ABC.ics"))
			require.NoError(t, err)
		})

		t.Run(repoName+"TimespanWithoutName", func(t *testing.T) {
			repo := newRepo(t)
			for range 2 {
				created, err := repo.CreateTimespan(ctx, testScope, model.Timespan{StartTime: start, EndTime: start.Add(time.Hour)})
				require.NoError(t, err)
				require.Nil(t, created.CaldavName)
			}
		})

		t.Run(repoName+"TaskCreateAndFindByName", func(t *testing.T) {
			repo := newRepo(t)
			tagIds := seedTags(t, ctx, repo, testScope, 1)
			created, err := repo.CreateTask(ctx, testScope, model.Task{Name: "Buy milk", TagIds: tagIds, CaldavName: new("Some Name.ics")})
			require.NoError(t, err)
			require.Equal(t, "Some Name.ics", *created.CaldavName)

			found, err := repo.GetTaskByCaldavName(ctx, testScope, "Some Name.ics")
			require.NoError(t, err)
			require.Equal(t, created.Id, found.Id)
			require.Equal(t, "Some Name.ics", *found.CaldavName)
			require.ElementsMatch(t, tagIds, found.TagIds)

			got, err := repo.GetTask(ctx, testScope, created.Id)
			require.NoError(t, err)
			require.Equal(t, "Some Name.ics", *got.CaldavName)

			_, err = repo.GetTaskByCaldavName(ctx, testScope, "missing.ics")
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"TaskUpdateKeepsName", func(t *testing.T) {
			repo := newRepo(t)
			created, err := repo.CreateTask(ctx, testScope, model.Task{Name: "Buy milk", CaldavName: new("A.ics")})
			require.NoError(t, err)

			updated, err := repo.UpdateTask(ctx, testScope, created.Id, model.TaskPatch{Name: new("Buy oat milk")})
			require.NoError(t, err)
			require.Equal(t, "A.ics", *updated.CaldavName)

			got, err := repo.GetTask(ctx, testScope, created.Id)
			require.NoError(t, err)
			require.Equal(t, "A.ics", *got.CaldavName)
		})

		t.Run(repoName+"TaskDuplicateName", func(t *testing.T) {
			repo := newRepo(t)
			first, err := repo.CreateTask(ctx, testScope, model.Task{Name: "a", CaldavName: new("A.ics")})
			require.NoError(t, err)

			_, err = repo.CreateTask(ctx, testScope, model.Task{Name: "b", CaldavName: new("A.ics")})
			require.ErrorIs(t, err, model.ErrAlreadyExists)

			require.NoError(t, repo.DeleteTask(ctx, testScope, first.Id))
			_, err = repo.CreateTask(ctx, testScope, model.Task{Name: "c", CaldavName: new("A.ics")})
			require.NoError(t, err)
		})

		t.Run(repoName+"UnownedScopeDuplicateName", func(t *testing.T) {
			repo := newRepo(t)
			_, err := repo.CreateTimespan(ctx, model.UnownedScope(), newTimespan("A.ics"))
			require.NoError(t, err)
			_, err = repo.CreateTimespan(ctx, model.UnownedScope(), newTimespan("A.ics"))
			require.ErrorIs(t, err, model.ErrAlreadyExists)

			_, err = repo.CreateTask(ctx, model.UnownedScope(), model.Task{Name: "a", CaldavName: new("A.ics")})
			require.NoError(t, err)
			_, err = repo.CreateTask(ctx, model.UnownedScope(), model.Task{Name: "b", CaldavName: new("A.ics")})
			require.ErrorIs(t, err, model.ErrAlreadyExists)
		})
	}

	run(t, "memory", func(t *testing.T) repository.Repository {
		return memory.NewMemoryStore()
	})

	run(t, "postgres", func(t *testing.T) repository.Repository {
		t.Parallel()
		pool := testutils.StartPostgresContainerWithMigrationsApplied(ctx, t)
		repo := postgres.NewPostgresStoreFromPool(pool)
		seedScopeUser(t, ctx, repo, testScope)
		return repo
	})
}
