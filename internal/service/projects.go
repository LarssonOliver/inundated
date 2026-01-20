package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
)

type ProjectServiceImpl struct {
	repository repository.ProjectRepository
}

var _ ProjectService = (*ProjectServiceImpl)(nil)

func NewProjectService(repo repository.ProjectRepository) *ProjectServiceImpl {
	return &ProjectServiceImpl{
		repository: repo,
	}
}

func (s *ProjectServiceImpl) GetProject(ctx context.Context, id uuid.UUID) (model.Project, error) {
	return s.repository.GetProject(ctx, id)
}

func (s *ProjectServiceImpl) ListProjects(ctx context.Context) ([]model.Project, error) {
	return s.repository.ListProjects(ctx)
}

func (s *ProjectServiceImpl) CreateProject(ctx context.Context, project model.Project) (model.Project, error) {
	return s.repository.CreateProject(ctx, project)
}

func (s *ProjectServiceImpl) UpdateProject(ctx context.Context, project model.Project) (model.Project, error) {
	return s.repository.UpdateProject(ctx, project)
}

func (s *ProjectServiceImpl) DeleteProject(ctx context.Context, id uuid.UUID) error {
	return s.repository.DeleteProject(ctx, id)
}
