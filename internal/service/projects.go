package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

func (s *ServiceImpl) GetProject(ctx context.Context, id uuid.UUID, includes *ProjectServiceGetIncludes) (model.Project, error) {
	scope, err := ownerScope(ctx)
	if err != nil {
		return model.Project{}, err
	}

	project, err := s.repository.GetProject(ctx, scope, id)
	if err != nil {
		return model.Project{}, err
	}

	if includes != nil {
		tagIds := project.EffectiveTagIds()
		if includes.TotalTime && len(tagIds) > 0 {
			totalTime, err := s.repository.GetTotalDurationByTags(ctx, scope, tagIds)
			if err != nil {
				return model.Project{}, err
			}
			project.TotalTime = &totalTime
		}
		if includes.TaskTime && len(tagIds) > 0 {
			taskTime, err := s.repository.GetTaskDurationByTags(ctx, scope, tagIds)
			if err != nil {
				return model.Project{}, err
			}
			project.TaskTime = &taskTime
		}
	}

	return project, nil
}

func (s *ServiceImpl) ListProjects(ctx context.Context, params model.PaginationParams) (model.Page[model.Project], error) {
	scope, err := ownerScope(ctx)
	if err != nil {
		return model.Page[model.Project]{}, err
	}
	return s.repository.ListProjects(ctx, scope, params)
}

func (s *ServiceImpl) CreateProject(ctx context.Context, project model.Project) (model.Project, error) {
	scope, err := ownerScope(ctx)
	if err != nil {
		return model.Project{}, err
	}
	project.Id = uuid.New()
	return s.repository.CreateProject(ctx, scope, project)
}

func (s *ServiceImpl) UpdateProject(ctx context.Context, project model.Project) (model.Project, error) {
	scope, err := ownerScope(ctx)
	if err != nil {
		return model.Project{}, err
	}
	return s.repository.UpdateProject(ctx, scope, project)
}

func (s *ServiceImpl) DeleteProject(ctx context.Context, id uuid.UUID) error {
	scope, err := ownerScope(ctx)
	if err != nil {
		return err
	}
	return s.repository.DeleteProject(ctx, scope, id)
}
