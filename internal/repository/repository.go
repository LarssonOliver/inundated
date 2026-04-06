package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

type Repository interface {
	TagRepository
	ProjectRepository
	TimespanRepository
}

type TagRepository interface {
	GetTag(ctx context.Context, id uuid.UUID) (model.Tag, error)
	ListTags(ctx context.Context) ([]model.Tag, error)
	CreateTag(ctx context.Context, tag model.Tag) (model.Tag, error)
	UpdateTag(ctx context.Context, tag model.Tag) (model.Tag, error)
	DeleteTag(ctx context.Context, id uuid.UUID) error
}

type ProjectRepository interface {
	GetProject(ctx context.Context, id uuid.UUID) (model.Project, error)
	ListProjects(ctx context.Context) ([]model.Project, error)
	CreateProject(ctx context.Context, project model.Project) (model.Project, error)
	UpdateProject(ctx context.Context, project model.Project) (model.Project, error)
	DeleteProject(ctx context.Context, id uuid.UUID) error
}

type TimespanRepository interface {
	GetTimespan(ctx context.Context, id uuid.UUID) (model.Timespan, error)
	ListTimespans(ctx context.Context) ([]model.Timespan, error)
	CreateTimespan(ctx context.Context, timespan model.Timespan) (model.Timespan, error)
	UpdateTimespan(ctx context.Context, timespan model.Timespan) (model.Timespan, error)
	DeleteTimespan(ctx context.Context, id uuid.UUID) error
	GetTotalDurationByTags(ctx context.Context, tagIds []uuid.UUID) (time.Duration, error)
	// ListTimespansByTagsAndTimeRange returns timespans that have ANY tag in tagIds
	// AND overlap with the time range [start, end)
	ListTimespansByTagsAndTimeRange(ctx context.Context, tagIds []uuid.UUID, start, end time.Time) ([]model.Timespan, error)
}
