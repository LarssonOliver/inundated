package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

type Repository interface {
	TagRepository
	ProjectRepository
	TimeSpanRepository
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

type TimeSpanRepository interface {
	GetTimeSpan(ctx context.Context, id uuid.UUID) (model.TimeSpan, error)
	ListTimeSpans(ctx context.Context) ([]model.TimeSpan, error)
	CreateTimeSpan(ctx context.Context, timeSpan model.TimeSpan) (model.TimeSpan, error)
	UpdateTimeSpan(ctx context.Context, timeSpan model.TimeSpan) (model.TimeSpan, error)
	DeleteTimeSpan(ctx context.Context, id uuid.UUID) error
}
