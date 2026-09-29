package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

func (s *ServiceImpl) GetTask(ctx context.Context, id uuid.UUID, includes *TaskServiceIncludes) (model.Task, error) {
	scope, err := ownerScope(ctx)
	if err != nil {
		return model.Task{}, err
	}

	task, err := s.repository.GetTask(ctx, scope, id)
	if err != nil {
		return model.Task{}, err
	}

	tasks := []model.Task{task}
	if err := s.addTaskIncludes(ctx, scope, tasks, includes); err != nil {
		return model.Task{}, err
	}
	return tasks[0], nil
}

func (s *ServiceImpl) ListTasks(ctx context.Context, params model.TaskListParams, includes *TaskServiceIncludes) (model.Page[model.Task], error) {
	scope, err := ownerScope(ctx)
	if err != nil {
		return model.Page[model.Task]{}, err
	}
	if params.DueFrom != nil && params.DueTo != nil && params.DueFrom.After(*params.DueTo) {
		return model.Page[model.Task]{}, fmt.Errorf("ListTasks: dueFrom is after dueTo: %w", model.ErrInvalidArgument)
	}
	page, err := s.repository.ListTasks(ctx, scope, params)
	if err != nil {
		return model.Page[model.Task]{}, err
	}
	if err := s.addTaskIncludes(ctx, scope, page.Data, includes); err != nil {
		return model.Page[model.Task]{}, err
	}
	return page, nil
}

// addTaskIncludes fills in the computed fields includes asks for, in place.
// A task's total time is the time on its task tag, which rolls up its
// subtasks because they inherit that tag.
func (s *ServiceImpl) addTaskIncludes(ctx context.Context, scope model.OwnerScope, tasks []model.Task, includes *TaskServiceIncludes) error {
	if includes == nil || len(tasks) == 0 {
		return nil
	}

	if includes.TotalTime {
		tagIds := make([]uuid.UUID, len(tasks))
		for i, task := range tasks {
			tagIds[i] = task.TagId
		}
		totals, err := s.repository.GetTotalDurationPerTag(ctx, scope, tagIds)
		if err != nil {
			return err
		}
		for i := range tasks {
			totalTime := totals[tasks[i].TagId]
			tasks[i].TotalTime = &totalTime
		}
	}

	if includes.ProjectIds {
		ids := make([]uuid.UUID, len(tasks))
		for i, task := range tasks {
			ids[i] = task.Id
		}
		projectIds, err := s.repository.ListTaskProjectIds(ctx, scope, ids)
		if err != nil {
			return err
		}
		for i := range tasks {
			tasks[i].ProjectIds = projectIds[tasks[i].Id]
			if tasks[i].ProjectIds == nil {
				tasks[i].ProjectIds = []uuid.UUID{}
			}
		}
	}
	return nil
}

func (s *ServiceImpl) CreateTask(ctx context.Context, task model.Task) (model.Task, error) {
	scope, err := ownerScope(ctx)
	if err != nil {
		return model.Task{}, err
	}
	if err := validateTask(task); err != nil {
		return model.Task{}, err
	}
	return s.repository.CreateTask(ctx, scope, task)
}

func (s *ServiceImpl) UpdateTask(ctx context.Context, id uuid.UUID, patch model.TaskPatch) (model.Task, error) {
	scope, err := ownerScope(ctx)
	if err != nil {
		return model.Task{}, err
	}
	if err := validateTaskPatch(patch); err != nil {
		return model.Task{}, err
	}
	return s.repository.UpdateTask(ctx, scope, id, patch)
}

func (s *ServiceImpl) MoveTask(ctx context.Context, id uuid.UUID, parentId *uuid.UUID, afterId *uuid.UUID) (model.Task, error) {
	scope, err := ownerScope(ctx)
	if err != nil {
		return model.Task{}, err
	}
	if afterId != nil && *afterId == id {
		return model.Task{}, fmt.Errorf("MoveTask: a task can't follow itself: %w", model.ErrInvalidArgument)
	}
	return s.repository.MoveTask(ctx, scope, id, parentId, afterId)
}

func (s *ServiceImpl) DeleteTask(ctx context.Context, id uuid.UUID) error {
	scope, err := ownerScope(ctx)
	if err != nil {
		return err
	}
	return s.repository.DeleteTask(ctx, scope, id)
}

// validateTask and validateTaskPatch check a task's field values. The
// repositories trust them, and check only the name (as they do for every
// resource) and what needs stored state. A new task is always open, so its
// close fields aren't checked.
func validateTask(task model.Task) error {
	if task.Estimate != nil && *task.Estimate < 0 {
		return fmt.Errorf("task estimate must not be negative: %w", model.ErrInvalidArgument)
	}
	return nil
}

func validateTaskPatch(patch model.TaskPatch) error {
	if patch.Estimate != nil && *patch.Estimate < 0 {
		return fmt.Errorf("task estimate must not be negative: %w", model.ErrInvalidArgument)
	}
	if patch.CloseReason != nil && !patch.CloseReason.Valid() {
		return fmt.Errorf("unknown close reason %q: %w", *patch.CloseReason, model.ErrInvalidArgument)
	}
	if (patch.DueDate != nil && patch.ClearDueDate) ||
		(patch.Estimate != nil && patch.ClearEstimate) ||
		(patch.CloseReason != nil && patch.Closed != nil && !*patch.Closed) {
		return fmt.Errorf("task patch both sets and clears a field: %w", model.ErrInvalidArgument)
	}
	return nil
}
