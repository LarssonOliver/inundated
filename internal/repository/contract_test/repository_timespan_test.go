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

func TestTimespanRepositoryContract(t *testing.T) {
	ctx := context.Background()

	run := func(
		t *testing.T,
		repoName string,
		newRepo func(t *testing.T) repository.Repository,
	) {
		t.Run(repoName+"CreateAndGet", func(t *testing.T) {
			repo := newRepo(t)

			tagIds := seedTags(t, ctx, repo, testScope, 2)

			start := time.Now().Add(-time.Hour).Round(0)
			end := time.Now().Round(0)

			ts := model.Timespan{
				Name:      "Work session",
				StartTime: start,
				EndTime:   end,
				TagIds:    tagIds,
			}

			created, err := repo.CreateTimespan(ctx, testScope, ts)
			ts.Id = created.Id
			require.NoError(t, err)

			got, err := repo.GetTimespan(ctx, testScope, ts.Id)
			require.NoError(t, err)
			require.Equal(t, "Work session", got.Name)
			require.WithinDuration(t, start, got.StartTime, time.Millisecond)
			require.WithinDuration(t, end, got.EndTime, time.Millisecond)
			require.ElementsMatch(t, tagIds, got.TagIds)
		})

		t.Run("List", func(t *testing.T) {
			repo := newRepo(t)

			_, _ = repo.CreateTimespan(ctx, testScope, model.Timespan{Name: "a", StartTime: time.Now(), EndTime: time.Now().Add(time.Hour)})
			_, _ = repo.CreateTimespan(ctx, testScope, model.Timespan{Name: "b", StartTime: time.Now(), EndTime: time.Now().Add(time.Hour)})

			page, err := repo.ListTimespans(ctx, testScope, model.DefaultPaginationParams())
			require.NoError(t, err)
			require.Len(t, page.Data, 2)
			require.Equal(t, 2, page.TotalCount)
		})

		t.Run(repoName+"CreateWithEmptyName", func(t *testing.T) {
			repo := newRepo(t)
			tagIds := seedTags(t, ctx, repo, testScope, 1)
			start := time.Now().Add(-time.Hour)
			end := time.Now()
			ts := model.Timespan{
				Name:      "",
				StartTime: start,
				EndTime:   end,
				TagIds:    tagIds,
			}

			_, err := repo.CreateTimespan(ctx, testScope, ts)
			require.NoError(t, err)
		})

		t.Run(repoName+"CreateFailsIfTagMissing", func(t *testing.T) {
			repo := newRepo(t)

			ts := model.Timespan{
				Name:      "Invalid",
				StartTime: time.Now(),
				EndTime:   time.Now().Add(time.Hour),
				TagIds:    []uuid.UUID{uuid.New()},
			}

			_, err := repo.CreateTimespan(ctx, testScope, ts)
			require.ErrorIs(t, err, model.ErrInvalidReference)
		})

		t.Run(repoName+"UpdateFailsIfTimespanMissing", func(t *testing.T) {
			repo := newRepo(t)

			ts := model.Timespan{
				Name:      "Valid",
				StartTime: time.Now(),
				EndTime:   time.Now().Add(time.Hour),
			}
			created, _ := repo.CreateTimespan(ctx, testScope, ts)
			ts.Id = created.Id

			ts.TagIds = []uuid.UUID{uuid.New()}

			_, err := repo.UpdateTimespan(ctx, testScope, ts)
			require.ErrorIs(t, err, model.ErrInvalidReference)
		})

		t.Run(repoName+"UpdateSetEmptyName", func(t *testing.T) {
			repo := newRepo(t)
			ts := model.Timespan{
				Name:      "Non-empty",
				StartTime: time.Now(),
				EndTime:   time.Now().Add(time.Hour),
			}

			created, _ := repo.CreateTimespan(ctx, testScope, ts)
			ts.Id = created.Id
			ts.Name = ""
			uts, err := repo.UpdateTimespan(ctx, testScope, ts)
			require.NoError(t, err)
			require.Equal(t, "", uts.Name)
		})

		t.Run(repoName+"Delete", func(t *testing.T) {
			repo := newRepo(t)

			ts := model.Timespan{
				Name:      "Temp",
				StartTime: time.Now(),
				EndTime:   time.Now().Add(time.Hour),
			}
			created, _ := repo.CreateTimespan(ctx, testScope, ts)
			ts.Id = created.Id

			err := repo.DeleteTimespan(ctx, testScope, ts.Id)
			require.NoError(t, err)

			_, err = repo.GetTimespan(ctx, testScope, ts.Id)
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"GetTotalDurationByTags", func(t *testing.T) {
			repo := newRepo(t)

			tags := seedTags(t, ctx, repo, testScope, 4)
			baseTime := time.Now().Add(-3 * time.Hour)

			ts1 := model.Timespan{StartTime: baseTime, EndTime: baseTime.Add(time.Hour), TagIds: []uuid.UUID{tags[0], tags[1]}}
			ts2 := model.Timespan{StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(4 * time.Hour), TagIds: []uuid.UUID{tags[1], tags[2]}}
			_, err := repo.CreateTimespan(ctx, testScope, ts1)
			require.NoError(t, err)
			_, err = repo.CreateTimespan(ctx, testScope, ts2)
			require.NoError(t, err)

			d1, err := repo.GetTotalDurationByTags(ctx, testScope, []uuid.UUID{tags[0]})
			require.NoError(t, err)
			require.Equal(t, time.Hour, d1)

			d2, err := repo.GetTotalDurationByTags(ctx, testScope, []uuid.UUID{tags[1]})
			require.NoError(t, err)
			require.Equal(t, 3*time.Hour, d2)

			d3, err := repo.GetTotalDurationByTags(ctx, testScope, tags)
			require.NoError(t, err)
			require.Equal(t, 3*time.Hour, d3)

			_, err = repo.GetTotalDurationByTags(ctx, testScope, []uuid.UUID{uuid.New()})
			require.NoError(t, err)

			d4, err := repo.GetTotalDurationByTags(ctx, testScope, []uuid.UUID{})
			require.NoError(t, err)
			require.Equal(t, 0*time.Hour, d4)

			d5, err := repo.GetTotalDurationByTags(ctx, testScope, []uuid.UUID{tags[3]})
			require.NoError(t, err)
			require.Equal(t, 0*time.Hour, d5)
		})

		t.Run(repoName+"GetTotalDurationByTags_IsScoped", func(t *testing.T) {
			repo := newRepo(t)
			scopeA := model.UserScope(uuid.New())
			scopeB := model.UserScope(uuid.New())
			seedScopeUser(t, ctx, repo, scopeA)
			seedScopeUser(t, ctx, repo, scopeB)

			tagA := seedTags(t, ctx, repo, scopeA, 1)[0]
			tagB := seedTags(t, ctx, repo, scopeB, 1)[0]

			start := time.Now().UTC().Truncate(time.Second)
			_, err := repo.CreateTimespan(ctx, scopeA, model.Timespan{
				Name: "a", StartTime: start, EndTime: start.Add(time.Hour), TagIds: []uuid.UUID{tagA},
			})
			require.NoError(t, err)
			_, err = repo.CreateTimespan(ctx, scopeB, model.Timespan{
				Name: "b", StartTime: start, EndTime: start.Add(2 * time.Hour), TagIds: []uuid.UUID{tagB},
			})
			require.NoError(t, err)

			durA, err := repo.GetTotalDurationByTags(ctx, scopeA, []uuid.UUID{tagA})
			require.NoError(t, err)
			require.Equal(t, time.Hour, durA)

			// even asking for B's tag under scope A yields nothing
			durCross, err := repo.GetTotalDurationByTags(ctx, scopeA, []uuid.UUID{tagB})
			require.NoError(t, err)
			require.Equal(t, time.Duration(0), durCross)
		})

		t.Run(repoName+"ListPagination_OffsetAndLimit", func(t *testing.T) {
			repo := newRepo(t)

			base := time.Now()
			for i := range 5 {
				_, _ = repo.CreateTimespan(ctx, testScope, model.Timespan{
					Name:      fmt.Sprintf("timespan-%d", i),
					StartTime: base.Add(time.Duration(i) * time.Hour),
					EndTime:   base.Add(time.Duration(i+1) * time.Hour),
				})
			}

			page, err := repo.ListTimespans(ctx, testScope, model.PaginationParams{Limit: 2, Offset: 0})
			require.NoError(t, err)
			require.Len(t, page.Data, 2)
			require.Equal(t, 5, page.TotalCount)

			page2, err := repo.ListTimespans(ctx, testScope, model.PaginationParams{Limit: 2, Offset: 2})
			require.NoError(t, err)
			require.Len(t, page2.Data, 2)
			require.Equal(t, 5, page2.TotalCount)

			ids1 := make(map[uuid.UUID]bool)
			for _, ts := range page.Data {
				ids1[ts.Id] = true
			}
			for _, ts := range page2.Data {
				require.False(t, ids1[ts.Id], "duplicate item across pages")
			}
		})

		t.Run(repoName+"ListPagination_OffsetBeyondEnd", func(t *testing.T) {
			repo := newRepo(t)

			base := time.Now()
			_, _ = repo.CreateTimespan(ctx, testScope, model.Timespan{
				Name:      "only",
				StartTime: base,
				EndTime:   base.Add(time.Hour),
			})

			page, err := repo.ListTimespans(ctx, testScope, model.PaginationParams{Limit: 10, Offset: 100})
			require.NoError(t, err)
			require.Empty(t, page.Data)
			require.Equal(t, 1, page.TotalCount)
		})

		t.Run(repoName+"ListPagination_LastPagePartial", func(t *testing.T) {
			repo := newRepo(t)

			base := time.Now()
			for i := range 5 {
				_, _ = repo.CreateTimespan(ctx, testScope, model.Timespan{
					Name:      fmt.Sprintf("timespan-%d", i),
					StartTime: base.Add(time.Duration(i) * time.Hour),
					EndTime:   base.Add(time.Duration(i+1) * time.Hour),
				})
			}

			page, err := repo.ListTimespans(ctx, testScope, model.PaginationParams{Limit: 3, Offset: 3})
			require.NoError(t, err)
			require.Len(t, page.Data, 2)
			require.Equal(t, 5, page.TotalCount)
		})

		t.Run(repoName+"ListPagination_EmptyStore", func(t *testing.T) {
			repo := newRepo(t)

			page, err := repo.ListTimespans(ctx, testScope, model.PaginationParams{Limit: 10, Offset: 0})
			require.NoError(t, err)
			require.Empty(t, page.Data)
			require.Equal(t, 0, page.TotalCount)
		})

		t.Run(repoName+"ListPagination_TotalCountUnaffectedByLimit", func(t *testing.T) {
			repo := newRepo(t)

			base := time.Now()
			for i := range 5 {
				_, _ = repo.CreateTimespan(ctx, testScope, model.Timespan{
					Name:      fmt.Sprintf("timespan-%d", i),
					StartTime: base.Add(time.Duration(i) * time.Hour),
					EndTime:   base.Add(time.Duration(i+1) * time.Hour),
				})
			}

			page, err := repo.ListTimespans(ctx, testScope, model.PaginationParams{Limit: 1, Offset: 0})
			require.NoError(t, err)
			require.Len(t, page.Data, 1)
			require.Equal(t, 5, page.TotalCount)
		})

		t.Run(repoName+"ScopeIsolation", func(t *testing.T) {
			repo := newRepo(t)
			scopeA := model.UserScope(uuid.New())
			scopeB := model.UserScope(uuid.New())
			seedScopeUser(t, ctx, repo, scopeA)
			seedScopeUser(t, ctx, repo, scopeB)

			start := time.Now().UTC()

			tsA, err := repo.CreateTimespan(ctx, scopeA, model.Timespan{
				Name: "a", StartTime: start, EndTime: start.Add(time.Hour),
			})
			require.NoError(t, err)
			_, err = repo.CreateTimespan(ctx, scopeB, model.Timespan{
				Name: "b", StartTime: start, EndTime: start.Add(time.Hour),
			})
			require.NoError(t, err)

			pageA, err := repo.ListTimespans(ctx, scopeA, model.DefaultPaginationParams())
			require.NoError(t, err)
			require.Len(t, pageA.Data, 1)
			require.Equal(t, 1, pageA.TotalCount)

			_, err = repo.GetTimespan(ctx, scopeB, tsA.Id)
			require.ErrorIs(t, err, model.ErrNotFound)

			tsA.Name = "hijack"
			_, err = repo.UpdateTimespan(ctx, scopeB, tsA)
			require.ErrorIs(t, err, model.ErrNotFound)

			err = repo.DeleteTimespan(ctx, scopeB, tsA.Id)
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"CannotAttachAnotherUsersTag", func(t *testing.T) {
			repo := newRepo(t)
			scopeA := model.UserScope(uuid.New())
			scopeB := model.UserScope(uuid.New())
			seedScopeUser(t, ctx, repo, scopeA)
			seedScopeUser(t, ctx, repo, scopeB)

			tagB := seedTags(t, ctx, repo, scopeB, 1)[0]

			start := time.Now().UTC()
			_, err := repo.CreateTimespan(ctx, scopeA, model.Timespan{
				Name: "a", StartTime: start, EndTime: start.Add(time.Hour),
				TagIds: []uuid.UUID{tagB},
			})
			require.ErrorIs(t, err, model.ErrInvalidReference)

			// The update path enforces the same guard: create a valid
			// resource under scopeA, then try to attach scopeB's tag.
			tsA, err := repo.CreateTimespan(ctx, scopeA, model.Timespan{
				Name: "a", StartTime: start, EndTime: start.Add(time.Hour),
			})
			require.NoError(t, err)

			tsA.TagIds = []uuid.UUID{tagB}
			_, err = repo.UpdateTimespan(ctx, scopeA, tsA)
			require.ErrorIs(t, err, model.ErrInvalidReference)
		})

		t.Run(repoName+"UnownedScopeIsolation", func(t *testing.T) {
			repo := newRepo(t)
			user := model.UserScope(uuid.New())
			seedScopeUser(t, ctx, repo, user)

			start := time.Now().UTC()

			owned, err := repo.CreateTimespan(ctx, user, model.Timespan{
				Name: "owned", StartTime: start, EndTime: start.Add(time.Hour),
			})
			require.NoError(t, err)
			_, err = repo.CreateTimespan(ctx, model.UnownedScope(), model.Timespan{
				Name: "unowned", StartTime: start, EndTime: start.Add(time.Hour),
			})
			require.NoError(t, err)

			page, err := repo.ListTimespans(ctx, model.UnownedScope(), model.DefaultPaginationParams())
			require.NoError(t, err)
			require.Len(t, page.Data, 1)

			_, err = repo.GetTimespan(ctx, model.UnownedScope(), owned.Id)
			require.ErrorIs(t, err, model.ErrNotFound)
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
