package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/utils"
)

const maxStatsBuckets = 10_000

func (s *ServiceImpl) GetProjectStats(ctx context.Context, input GetProjectStatsInput) (model.ProjectStats, error) {
	if input.Metric != model.ProjectStatsMetricTimeSpent {
		return model.ProjectStats{}, model.ErrInvalidArgument
	}

	project, err := s.repository.GetProject(ctx, input.ProjectID)
	if err != nil {
		return model.ProjectStats{}, err
	}

	now := input.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}

	timezoneRaw := optionalTrimmedString(input.TimezoneRaw)
	location, err := utils.ParseTimezone(timezoneRaw)
	if err != nil {
		return model.ProjectStats{}, model.ErrInvalidArgument
	}

	intervalRaw := optionalTrimmedString(input.IntervalRaw)
	if intervalRaw == "" {
		intervalRaw = fmt.Sprintf("P30D/%s", now.UTC().Format(time.RFC3339))
	}

	interval, err := utils.ParseISO8601Interval(intervalRaw, now, location)
	if err != nil {
		if errors.Is(err, utils.ErrISO8601Unprocessable) {
			return model.ProjectStats{}, model.ErrUnprocessable
		}
		return model.ProjectStats{}, model.ErrInvalidArgument
	}

	granularityRaw := optionalTrimmedString(input.GranularityRaw)
	if granularityRaw == "" {
		granularityRaw = "P1D"
	}

	granularity, err := utils.ParseISO8601Duration(granularityRaw)
	if err != nil {
		return model.ProjectStats{}, model.ErrInvalidArgument
	}

	buckets, err := utils.BuildTimeBuckets(utils.ResolvedInterval{
		Start: interval.Start,
		End:   interval.End,
	}, granularity, location, maxStatsBuckets)
	if err != nil {
		if errors.Is(err, utils.ErrISO8601Unprocessable) {
			return model.ProjectStats{}, model.ErrUnprocessable
		}
		return model.ProjectStats{}, model.ErrInvalidArgument
	}

	bucketRanges := make([]model.BucketRange, 0, len(buckets))
	for _, bucket := range buckets {
		bucketRanges = append(bucketRanges, model.BucketRange{
			Start: bucket.Start,
			End:   bucket.End,
		})
	}

	series, err := s.repository.AggregateTimeSpentByTagsAndBuckets(ctx, project.TagIds, bucketRanges)
	if err != nil {
		if errors.Is(err, model.ErrInvalidArgument) {
			return model.ProjectStats{}, model.ErrUnprocessable
		}

		return model.ProjectStats{}, err
	}

	if len(series) != len(bucketRanges) {
		return model.ProjectStats{}, fmt.Errorf("GetProjectStats: unexpected series length")
	}

	return model.ProjectStats{
		ProjectID:   project.Id,
		Metric:      input.Metric,
		Interval:    model.BucketRange{Start: interval.Start, End: interval.End},
		Granularity: granularityRaw,
		Unit:        "seconds",
		Series:      series,
	}, nil
}

func optionalTrimmedString(value *string) string {
	if value == nil {
		return ""
	}

	return strings.TrimSpace(*value)
}
