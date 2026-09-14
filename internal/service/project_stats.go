package service

import (
	"context"

	"github.com/larssonoliver/inundated/internal/model"
)

func (s *ServiceImpl) GetProjectStats(ctx context.Context, input GetProjectStatsInput) (model.ProjectStats, error) {
	if input.Metric != model.StatsMetricTimeSpent {
		return model.ProjectStats{}, model.ErrInvalidArgument
	}

	scope, err := ownerScope(ctx)
	if err != nil {
		return model.ProjectStats{}, err
	}

	project, err := s.repository.GetProject(ctx, scope, input.ProjectID)
	if err != nil {
		return model.ProjectStats{}, err
	}

	result, err := s.computeTimeSpentSeries(ctx, scope, project.TagIds, input.IntervalRaw, input.GranularityRaw, input.TimezoneRaw, input.Now)
	if err != nil {
		return model.ProjectStats{}, err
	}

	return model.ProjectStats{
		ProjectID:   project.Id,
		Metric:      input.Metric,
		Interval:    result.Interval,
		Granularity: result.Granularity,
		Unit:        "seconds",
		Series:      result.Series,
	}, nil
}
