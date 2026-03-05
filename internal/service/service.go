package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
)

type ServiceImpl struct {
	repository repository.Repository
}

var _ Service = (*ServiceImpl)(nil)

func NewService(repository repository.Repository) *ServiceImpl {
	return &ServiceImpl{
		repository: repository,
	}
}

type Service interface {
	TagService
	ProjectService
	TimespanService
}

type TagServiceGetIncludes struct {
	TotalTime bool
}

type TagService interface {
	GetTag(ctx context.Context, id uuid.UUID, includes *TagServiceGetIncludes) (model.Tag, error)
	ListTags(ctx context.Context) ([]model.Tag, error)
	CreateTag(ctx context.Context, tag model.Tag) (model.Tag, error)
	UpdateTag(ctx context.Context, tag model.Tag) (model.Tag, error)
	DeleteTag(ctx context.Context, id uuid.UUID) error
}

type ProjectServiceGetIncludes struct {
	TotalTime bool
}

type ProjectService interface {
	GetProject(ctx context.Context, id uuid.UUID, includes *ProjectServiceGetIncludes) (model.Project, error)
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
