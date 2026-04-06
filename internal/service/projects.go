package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/utils"
)

func (s *ServiceImpl) GetProject(ctx context.Context, id uuid.UUID, includes *ProjectServiceGetIncludes) (model.Project, error) {
	project, err := s.repository.GetProject(ctx, id)

	if err != nil {
		return model.Project{}, model.ErrNotFound
	}

	if includes != nil {
		if includes.TotalTime && project.TagIds != nil && len(project.TagIds) > 0 {
			totalTime, err := s.repository.GetTotalDurationByTags(ctx, project.TagIds)
			if err != nil {
				return model.Project{}, model.ErrNotFound
			}
			project.TotalTime = &totalTime
		}
	}

	return project, nil
}

func (s *ServiceImpl) ListProjects(ctx context.Context) ([]model.Project, error) {
	return s.repository.ListProjects(ctx)
}

func (s *ServiceImpl) CreateProject(ctx context.Context, project model.Project) (model.Project, error) {
	project.Id = uuid.New()
	return s.repository.CreateProject(ctx, project)
}

func (s *ServiceImpl) UpdateProject(ctx context.Context, project model.Project) (model.Project, error) {
	return s.repository.UpdateProject(ctx, project)
}

func (s *ServiceImpl) DeleteProject(ctx context.Context, id uuid.UUID) error {
	return s.repository.DeleteProject(ctx, id)
}

func (s *ServiceImpl) GetProjectStats(ctx context.Context, projectID uuid.UUID, metric string, intervalStr string, granularity string, timezoneStr string) (model.ProjectStats, error) {
	// Validate metric
	if metric != "time_spent" {
		return model.ProjectStats{}, fmt.Errorf("unsupported metric: %s", metric)
	}

	// Validate project exists and get its tags
	project, err := s.repository.GetProject(ctx, projectID)
	if err != nil {
		return model.ProjectStats{}, model.ErrNotFound
	}

	if len(project.TagIds) == 0 {
		// Project has no tags, return empty stats
		interval, err := parseInterval(intervalStr, time.Now())
		if err != nil {
			return model.ProjectStats{}, err
		}

		return model.ProjectStats{
			ProjectID:   projectID.String(),
			Metric:      metric,
			Interval:    utils.FormatIntervalAsISO8601(interval),
			Granularity: granularity,
			Unit:        "seconds",
			Series:      []model.TimeSeriesPoint{},
		}, nil
	}

	// Parse interval
	interval, err := parseInterval(intervalStr, time.Now())
	if err != nil {
		return model.ProjectStats{}, err
	}

	// Validate interval (start must be before end)
	if !interval.Start.Before(interval.End) {
		return model.ProjectStats{}, fmt.Errorf("interval start must be before end")
	}

	// Parse timezone
	tz, err := time.LoadLocation(timezoneStr)
	if err != nil {
		return model.ProjectStats{}, fmt.Errorf("invalid timezone: %w", err)
	}

	// Validate granularity
	if _, err := utils.ParseISO8601Duration(granularity); err != nil {
		return model.ProjectStats{}, fmt.Errorf("invalid granularity: %w", err)
	}

	// Fetch timespans that match the project's tags and time range
	timespans, err := s.repository.ListTimespansByTagsAndTimeRange(ctx, project.TagIds, interval.Start, interval.End)
	if err != nil {
		return model.ProjectStats{}, fmt.Errorf("failed to fetch timespans: %w", err)
	}

	// Generate buckets
	buckets, err := utils.GenerateBuckets(interval, granularity, tz)
	if err != nil {
		return model.ProjectStats{}, fmt.Errorf("failed to generate buckets: %w", err)
	}

	// Aggregate timespans into buckets
	bucketValues := make(map[int]time.Duration)

	for _, timespan := range timespans {
		// Calculate the duration of the timespan
		timespanDuration := timespan.EndTime.Sub(timespan.StartTime)

		// Split the duration across buckets
		splits := utils.SplitDurationAcrossBuckets(buckets, timespan.StartTime, timespan.EndTime, timespanDuration)

		// Add to bucket totals
		for bucketIdx, duration := range splits {
			bucketValues[bucketIdx] += duration
		}
	}

	// Convert to response format
	series := make([]model.TimeSeriesPoint, len(buckets))
	for i, bucket := range buckets {
		value := bucketValues[i]
		series[i] = model.TimeSeriesPoint{
			Interval: utils.FormatIntervalWithDuration(bucket.Start, bucket.Duration()),
			Value:    int64(value.Seconds()),
		}
	}

	return model.ProjectStats{
		ProjectID:   projectID.String(),
		Metric:      metric,
		Interval:    utils.FormatIntervalAsISO8601(interval),
		Granularity: granularity,
		Unit:        "seconds",
		Series:      series,
	}, nil
}

// parseInterval parses an interval string, handling the default case
func parseInterval(intervalStr string, now time.Time) (utils.Interval, error) {
	if intervalStr == "" {
		// Default: P30D/{now}
		intervalStr = "P30D/{now}"
	}
	return utils.ParseISO8601Interval(intervalStr, now)
}
