package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

func (s *ServiceImpl) GetTagStats(ctx context.Context, input GetTagStatsInput) (model.TagStats, error) {
	if input.Metric != model.StatsMetricTimeSpent {
		return model.TagStats{}, model.ErrInvalidArgument
	}

	scope, err := ownerScope(ctx)
	if err != nil {
		return model.TagStats{}, err
	}

	tag, err := s.GetTag(ctx, input.TagID, nil)
	if err != nil {
		return model.TagStats{}, err
	}

	result, err := s.computeTimeSpentSeries(ctx, scope, []uuid.UUID{tag.Id}, input.IntervalRaw, input.GranularityRaw, input.TimezoneRaw, input.Now)
	if err != nil {
		return model.TagStats{}, err
	}

	return model.TagStats{
		TagID:       tag.Id,
		Metric:      input.Metric,
		Interval:    result.Interval,
		Granularity: result.Granularity,
		Unit:        "seconds",
		Series:      result.Series,
	}, nil
}
