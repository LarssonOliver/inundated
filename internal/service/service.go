package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
)

type ServiceImpl struct {
	TagServiceImpl
	ProjectServiceImpl
	TimespanServiceImpl
}

var _ Service = (*ServiceImpl)(nil)

func NewService(repository repository.Repository) *ServiceImpl {
	return &ServiceImpl{
		*NewTagService(repository),
		*NewProjectService(repository),
		*NewTimespanService(repository),
	}
}

type Service interface {
	TagService
	ProjectService
	TimespanService
}

type TagService interface {
	GetTag(ctx context.Context, id uuid.UUID) (model.Tag, error)
	ListTags(ctx context.Context) ([]model.Tag, error)
	CreateTag(ctx context.Context, tag model.Tag) (model.Tag, error)
	UpdateTag(ctx context.Context, tag model.Tag) (model.Tag, error)
	DeleteTag(ctx context.Context, id uuid.UUID) error
}

type ProjectService interface {
	GetProject(ctx context.Context, id uuid.UUID) (model.Project, error)
	ListProjects(ctx context.Context) ([]model.Project, error)
	CreateProject(ctx context.Context, project model.Project) (model.Project, error)
	UpdateProject(ctx context.Context, project model.Project) (model.Project, error)
	DeleteProject(ctx context.Context, id uuid.UUID) error
}

type TimespanService interface {
	GetTimespan(ctx context.Context, id uuid.UUID) (model.Timespan, error)
	ListTimespans(ctx context.Context) ([]model.Timespan, error)
	CreateTimespan(ctx context.Context, timespan model.Timespan) (model.Timespan, error)
	UpdateTimespan(ctx context.Context, timespan model.Timespan) (model.Timespan, error)
	DeleteTimespan(ctx context.Context, id uuid.UUID) error
}
