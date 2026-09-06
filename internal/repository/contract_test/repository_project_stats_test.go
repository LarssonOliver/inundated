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

func TestProjectStatsRepositoryContract(t *testing.T) {
	ctx := context.Background()

	run := func(
		t *testing.T,
		repoName string,
		newRepo func(t *testing.T) repository.Repository,
	) {
		t.Run(repoName+"AggregateTimeSpentByTagsAndBuckets_SplitsOverlapAndDeduplicatesByTimespan", func(t *testing.T) {
			repo := newRepo(t)

			tags := seedTags(t, ctx, repo, testScope, 3)
			base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

			// ts1 overlaps both buckets and contains two requested tags; it must be counted once.
			_, err := repo.CreateTimespan(ctx, testScope, model.Timespan{
				Name:      "ts1",
				StartTime: base.Add(15 * time.Minute),
				EndTime:   base.Add(75 * time.Minute),
				TagIds:    []uuid.UUID{tags[0], tags[1]},
			})
			require.NoError(t, err)

			// ts2 only contributes to the second bucket.
			_, err = repo.CreateTimespan(ctx, testScope, model.Timespan{
				Name:      "ts2",
				StartTime: base.Add(60 * time.Minute),
				EndTime:   base.Add(120 * time.Minute),
				TagIds:    []uuid.UUID{tags[1]},
			})
			require.NoError(t, err)

			// ts3 should not contribute because its tag is not requested.
			_, err = repo.CreateTimespan(ctx, testScope, model.Timespan{
				Name:      "ts3",
				StartTime: base,
				EndTime:   base.Add(2 * time.Hour),
				TagIds:    []uuid.UUID{tags[2]},
			})
			require.NoError(t, err)

			buckets := []model.BucketRange{
				{Start: base, End: base.Add(1 * time.Hour)},
				{Start: base.Add(1 * time.Hour), End: base.Add(2 * time.Hour)},
			}

			got, err := repo.AggregateTimeSpentByTagsAndBuckets(ctx, testScope, []uuid.UUID{tags[0], tags[1]}, buckets)
			require.NoError(t, err)
			require.Len(t, got, len(buckets))
			requireBucketEquals(t, buckets[0], got[0].Bucket)
			requireBucketEquals(t, buckets[1], got[1].Bucket)

			// Values are expected in seconds.
			require.InDelta(t, 45*60, got[0].Value, 0.0001)
			require.InDelta(t, 75*60, got[1].Value, 0.0001)
		})

		t.Run(repoName+"AggregateTimeSpentByTagsAndBuckets_NoMatchesReturnsZeroPerBucket", func(t *testing.T) {
			repo := newRepo(t)

			tags := seedTags(t, ctx, repo, testScope, 1)
			base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

			_, err := repo.CreateTimespan(ctx, testScope, model.Timespan{
				Name:      "ts",
				StartTime: base,
				EndTime:   base.Add(90 * time.Minute),
				TagIds:    []uuid.UUID{tags[0]},
			})
			require.NoError(t, err)

			buckets := []model.BucketRange{
				{Start: base, End: base.Add(1 * time.Hour)},
				{Start: base.Add(1 * time.Hour), End: base.Add(2 * time.Hour)},
			}

			got, err := repo.AggregateTimeSpentByTagsAndBuckets(ctx, testScope, []uuid.UUID{uuid.New()}, buckets)
			require.NoError(t, err)
			require.Len(t, got, len(buckets))
			requireBucketEquals(t, buckets[0], got[0].Bucket)
			requireBucketEquals(t, buckets[1], got[1].Bucket)
			require.InDelta(t, 0.0, got[0].Value, 0.0001)
			require.InDelta(t, 0.0, got[1].Value, 0.0001)
		})

		t.Run(repoName+"AggregateTimeSpentByTagsAndBuckets_EmptyTagFilterReturnsZeroPerBucket", func(t *testing.T) {
			repo := newRepo(t)

			tags := seedTags(t, ctx, repo, testScope, 1)
			base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

			_, err := repo.CreateTimespan(ctx, testScope, model.Timespan{
				Name:      "ts",
				StartTime: base,
				EndTime:   base.Add(90 * time.Minute),
				TagIds:    []uuid.UUID{tags[0]},
			})
			require.NoError(t, err)

			buckets := []model.BucketRange{
				{Start: base, End: base.Add(1 * time.Hour)},
				{Start: base.Add(1 * time.Hour), End: base.Add(2 * time.Hour)},
			}

			got, err := repo.AggregateTimeSpentByTagsAndBuckets(ctx, testScope, []uuid.UUID{}, buckets)
			require.NoError(t, err)
			require.Len(t, got, len(buckets))
			requireBucketEquals(t, buckets[0], got[0].Bucket)
			requireBucketEquals(t, buckets[1], got[1].Bucket)
			require.InDelta(t, 0.0, got[0].Value, 0.0001)
			require.InDelta(t, 0.0, got[1].Value, 0.0001)
		})

		t.Run(repoName+"AggregateTimeSpentByTagsAndBuckets_InvalidBucketReturnsInvalidArgument", func(t *testing.T) {
			repo := newRepo(t)

			base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			buckets := []model.BucketRange{
				{Start: base.Add(1 * time.Hour), End: base.Add(1 * time.Hour)},
			}

			_, err := repo.AggregateTimeSpentByTagsAndBuckets(ctx, testScope, []uuid.UUID{uuid.New()}, buckets)
			require.ErrorIs(t, err, model.ErrInvalidArgument)
		})

		t.Run(repoName+"AggregateTimeSpentByTagsAndBuckets_PreservesInputBucketOrder", func(t *testing.T) {
			repo := newRepo(t)

			base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			buckets := []model.BucketRange{
				{Start: base.Add(1 * time.Hour), End: base.Add(2 * time.Hour)},
				{Start: base, End: base.Add(1 * time.Hour)},
			}

			got, err := repo.AggregateTimeSpentByTagsAndBuckets(ctx, testScope, []uuid.UUID{}, buckets)
			require.NoError(t, err)
			require.Len(t, got, len(buckets))
			requireBucketEquals(t, buckets[0], got[0].Bucket)
			requireBucketEquals(t, buckets[1], got[1].Bucket)
		})
	}

	run(t, "memory", func(t *testing.T) repository.Repository {
		return memory.NewMemoryStore()
	})

	run(t, "postgres", func(t *testing.T) repository.Repository {
		t.Parallel()
		pool := testutils.StartPostgresContainerWithMigrationsApplied(ctx, t)
		return postgres.NewPostgresStoreFromPool(pool)
	})
}

func requireBucketEquals(t *testing.T, want model.BucketRange, got model.BucketRange) {
	t.Helper()

	require.True(t, got.Start.Equal(want.Start))
	require.True(t, got.End.Equal(want.End))
}
