package repository_test

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

func TestTimespanRepositoryContract(t *testing.T) {
	ctx := context.Background()

	run := func(
		t *testing.T,
		repoName string,
		newRepo func(t *testing.T) repository.Repository,
	) {
		t.Run(repoName+"CreateAndGet", func(t *testing.T) {
			repo := newRepo(t)

			tagIds := seedTags(t, ctx, repo, 2)

			start := time.Now().Add(-time.Hour).Round(0)
			end := time.Now().Round(0)

			ts := model.Timespan{
				Name:      "Work session",
				StartTime: start,
				EndTime:   end,
				TagIds:    tagIds,
			}

			created, err := repo.CreateTimespan(ctx, ts)
			ts.Id = created.Id
			require.NoError(t, err)

			got, err := repo.GetTimespan(ctx, ts.Id)
			require.NoError(t, err)
			require.Equal(t, "Work session", got.Name)
			require.WithinDuration(t, start, got.StartTime, time.Millisecond)
			require.WithinDuration(t, end, got.EndTime, time.Millisecond)
			require.ElementsMatch(t, tagIds, got.TagIds)
		})

		t.Run(repoName+"CreateWithEmptyName", func(t *testing.T) {
			repo := newRepo(t)
			tagIds := seedTags(t, ctx, repo, 1)
			start := time.Now().Add(-time.Hour)
			end := time.Now()
			ts := model.Timespan{
				Name:      "",
				StartTime: start,
				EndTime:   end,
				TagIds:    tagIds,
			}

			_, err := repo.CreateTimespan(ctx, ts)
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

			_, err := repo.CreateTimespan(ctx, ts)
			require.ErrorIs(t, err, model.ErrInvalidReference)
		})

		t.Run(repoName+"UpdateFailsIfTimespanMissing", func(t *testing.T) {
			repo := newRepo(t)

			ts := model.Timespan{
				Name:      "Valid",
				StartTime: time.Now(),
				EndTime:   time.Now().Add(time.Hour),
			}
			created, _ := repo.CreateTimespan(ctx, ts)
			ts.Id = created.Id

			ts.TagIds = []uuid.UUID{uuid.New()}

			_, err := repo.UpdateTimespan(ctx, ts)
			require.ErrorIs(t, err, model.ErrInvalidReference)
		})

		t.Run(repoName+"UpdateSetEmptyName", func(t *testing.T) {
			repo := newRepo(t)
			ts := model.Timespan{
				Name:      "Non-empty",
				StartTime: time.Now(),
				EndTime:   time.Now().Add(time.Hour),
			}

			created, _ := repo.CreateTimespan(ctx, ts)
			ts.Id = created.Id
			ts.Name = ""
			uts, err := repo.UpdateTimespan(ctx, ts)
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
			created, _ := repo.CreateTimespan(ctx, ts)
			ts.Id = created.Id

			err := repo.DeleteTimespan(ctx, ts.Id)
			require.NoError(t, err)

			_, err = repo.GetTimespan(ctx, ts.Id)
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"GetTotalDurationByTags", func(t *testing.T) {
			repo := newRepo(t)

			tags := seedTags(t, ctx, repo, 4)
			baseTime := time.Now().Add(-3 * time.Hour)

			ts1 := model.Timespan{StartTime: baseTime, EndTime: baseTime.Add(time.Hour), TagIds: []uuid.UUID{tags[0], tags[1]}}
			ts2 := model.Timespan{StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(4 * time.Hour), TagIds: []uuid.UUID{tags[1], tags[2]}}
			_, err := repo.CreateTimespan(ctx, ts1)
			require.NoError(t, err)
			_, err = repo.CreateTimespan(ctx, ts2)
			require.NoError(t, err)

			d1, err := repo.GetTotalDurationByTags(ctx, []uuid.UUID{tags[0]})
			require.NoError(t, err)
			require.Equal(t, time.Hour, d1)

			d2, err := repo.GetTotalDurationByTags(ctx, []uuid.UUID{tags[1]})
			require.NoError(t, err)
			require.Equal(t, 3*time.Hour, d2)

			d3, err := repo.GetTotalDurationByTags(ctx, tags)
			require.NoError(t, err)
			require.Equal(t, 3*time.Hour, d3)

			_, err = repo.GetTotalDurationByTags(ctx, []uuid.UUID{uuid.New()})
			require.NoError(t, err)

			d4, err := repo.GetTotalDurationByTags(ctx, []uuid.UUID{})
			require.NoError(t, err)
			require.Equal(t, 0*time.Hour, d4)

			d5, err := repo.GetTotalDurationByTags(ctx, []uuid.UUID{tags[3]})
			require.NoError(t, err)
			require.Equal(t, 0*time.Hour, d5)
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

