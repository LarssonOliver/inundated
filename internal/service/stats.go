package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/utils"
)

const maxStatsBuckets = 10_000

type timeSpentSeriesResult struct {
	Interval    model.BucketRange
	Granularity string
	Series      []model.BucketValue
}

// computeTimeSpentSeries resolves the requested timezone/interval/granularity,
// builds the bucket boundaries, and aggregates time spent across tagIds into
// those buckets. Shared by GetProjectStats and GetTagStats, which only differ
// in how they resolve the tag IDs to aggregate over.
func (s *ServiceImpl) computeTimeSpentSeries(
	ctx context.Context,
	scope model.OwnerScope,
	tagIds []uuid.UUID,
	intervalRaw, granularityRaw, timezoneRaw *string,
	now time.Time,
) (timeSpentSeriesResult, error) {
	if now.IsZero() {
		now = time.Now().UTC()
	}

	timezone := optionalTrimmedString(timezoneRaw)
	location, err := utils.ParseTimezone(timezone)
	if err != nil {
		return timeSpentSeriesResult{}, model.ErrInvalidArgument
	}

	interval := optionalTrimmedString(intervalRaw)
	if interval == "" {
		interval = fmt.Sprintf("P30D/%s", now.UTC().Format(time.RFC3339))
	}

	resolvedInterval, err := utils.ParseISO8601Interval(interval, now, location)
	if err != nil {
		if errors.Is(err, utils.ErrISO8601Unprocessable) {
			return timeSpentSeriesResult{}, model.ErrUnprocessable
		}
		return timeSpentSeriesResult{}, model.ErrInvalidArgument
	}

	granularityRawValue := optionalTrimmedString(granularityRaw)
	if granularityRawValue == "" {
		granularityRawValue = "P1D"
	}

	granularity, err := utils.ParseISO8601Duration(granularityRawValue)
	if err != nil {
		return timeSpentSeriesResult{}, model.ErrInvalidArgument
	}

	buckets, err := utils.BuildTimeBuckets(utils.ResolvedInterval{
		Start: resolvedInterval.Start,
		End:   resolvedInterval.End,
	}, granularity, location, maxStatsBuckets)
	if err != nil {
		if errors.Is(err, utils.ErrISO8601Unprocessable) {
			return timeSpentSeriesResult{}, model.ErrUnprocessable
		}
		return timeSpentSeriesResult{}, model.ErrInvalidArgument
	}

	bucketRanges := make([]model.BucketRange, 0, len(buckets))
	for _, bucket := range buckets {
		bucketRanges = append(bucketRanges, model.BucketRange{
			Start: bucket.Start,
			End:   bucket.End,
		})
	}

	series, err := s.repository.AggregateTimeSpentByTagsAndBuckets(ctx, scope, tagIds, bucketRanges)
	if err != nil {
		if errors.Is(err, model.ErrInvalidArgument) {
			return timeSpentSeriesResult{}, model.ErrUnprocessable
		}

		return timeSpentSeriesResult{}, err
	}

	if len(series) != len(bucketRanges) {
		return timeSpentSeriesResult{}, fmt.Errorf("computeTimeSpentSeries: unexpected series length")
	}

	return timeSpentSeriesResult{
		Interval:    model.BucketRange{Start: resolvedInterval.Start, End: resolvedInterval.End},
		Granularity: granularityRawValue,
		Series:      series,
	}, nil
}

func optionalTrimmedString(value *string) string {
	if value == nil {
		return ""
	}

	return strings.TrimSpace(*value)
}
