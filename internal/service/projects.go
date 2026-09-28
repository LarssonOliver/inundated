package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

func (s *ServiceImpl) GetProject(ctx context.Context, id uuid.UUID, includes *ProjectServiceGetIncludes) (model.Project, error) {
	scope, err := ownerScope(ctx)
	if err != nil {
		return model.Project{}, err
	}

	project, err := s.repository.GetProject(ctx, scope, id)

	if errors.Is(err, model.ErrInvalidArgument) {
		// The only invalid argument a read-by-id raises is a malformed id
		// (e.g. the zero UUID); it can never name a real row, so it's a miss.
		return model.Project{}, fmt.Errorf("GetProject %s: %w", id, model.ErrNotFound)
	}
	if err != nil {
		// Propagate as-is: a genuine miss already carries model.ErrNotFound,
		// and an infrastructure failure must not be masked as a 404.
		return model.Project{}, err
	}

	if includes != nil {
		if includes.TotalTime && len(project.TagIds) > 0 {
			totalTime, err := s.repository.GetTotalDurationByTags(ctx, scope, project.TagIds)
			if err != nil {
				return model.Project{}, err
			}
			project.TotalTime = &totalTime
		}
		if includes.TaskTime {
			taskTime, err := s.projectTaskTime(ctx, scope, project.Id)
			if err != nil {
				return model.Project{}, err
			}
			project.TaskTime = &taskTime
		}
	}

	return project, nil
}

// projectTaskTime is the time logged on the project's tasks, open or
// closed. Each timespan counts once however many of the tasks it is on, and
// every such timespan also counts toward the project, because a task in the
// project carries one of its tags.
func (s *ServiceImpl) projectTaskTime(ctx context.Context, scope model.OwnerScope, projectId uuid.UUID) (time.Duration, error) {
	const pageSize = 100
	var taskTagIds []uuid.UUID
	for offset := 0; ; offset += pageSize {
		page, err := s.repository.ListTasks(ctx, scope, model.TaskListParams{
			PaginationParams: model.PaginationParams{Limit: pageSize, Offset: offset},
			IncludeClosed:    true,
			ProjectId:        &projectId,
		})
		if err != nil {
			return 0, err
		}
		for _, task := range page.Data {
			taskTagIds = append(taskTagIds, task.TagId)
		}
		if len(page.Data) < pageSize {
			break
		}
	}
	return s.repository.GetTotalDurationByTags(ctx, scope, taskTagIds)
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
