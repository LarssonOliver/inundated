package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

func (s *ServiceImpl) GetProject(ctx context.Context, id uuid.UUID, includes *ProjectServiceGetIncludes) (model.Project, error) {
	scope := ownerScope(ctx)

	project, err := s.repository.GetProject(ctx, scope, id)

	if err != nil {
		return model.Project{}, model.ErrNotFound
	}

	if includes != nil {
		if includes.TotalTime && project.TagIds != nil && len(project.TagIds) > 0 {
			totalTime, err := s.repository.GetTotalDurationByTags(ctx, scope, project.TagIds)
			if err != nil {
				return model.Project{}, model.ErrNotFound
			}
			project.TotalTime = &totalTime
		}
	}

	return project, nil
}

func (s *ServiceImpl) ListProjects(ctx context.Context, params model.PaginationParams) (model.Page[model.Project], error) {
	scope := ownerScope(ctx)
	return s.repository.ListProjects(ctx, scope, params)
}

func (s *ServiceImpl) CreateProject(ctx context.Context, project model.Project) (model.Project, error) {
	scope := ownerScope(ctx)
	project.Id = uuid.New()
	return s.repository.CreateProject(ctx, scope, project)
}

func (s *ServiceImpl) UpdateProject(ctx context.Context, project model.Project) (model.Project, error) {
	scope := ownerScope(ctx)
	return s.repository.UpdateProject(ctx, scope, project)
}

func (s *ServiceImpl) DeleteProject(ctx context.Context, id uuid.UUID) error {
	scope := ownerScope(ctx)
	return s.repository.DeleteProject(ctx, scope, id)
}
