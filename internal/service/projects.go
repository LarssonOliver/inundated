package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
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
