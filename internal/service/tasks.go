package service

import (
	"context"
	"errors"
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
	if errors.Is(err, model.ErrInvalidArgument) {
		// The only invalid argument a read-by-id raises is a malformed id
		// (e.g. the zero UUID); it can never name a real row, so it's a miss.
		return model.Task{}, fmt.Errorf("GetTask %s: %w", id, model.ErrNotFound)
	}
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
		for i := range tasks {
			totalTime, err := s.repository.GetTotalDurationByTags(ctx, scope, []uuid.UUID{tasks[i].TagId})
			if err != nil {
				return err
			}
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
	if id == uuid.Nil {
		return model.Task{}, fmt.Errorf("UpdateTask %s: %w", id, model.ErrNotFound)
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
	if id == uuid.Nil {
		return model.Task{}, fmt.Errorf("MoveTask %s: %w", id, model.ErrNotFound)
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
	if id == uuid.Nil {
		// As in GetTask: a malformed id can never name a real row.
		return fmt.Errorf("DeleteTask %s: %w", id, model.ErrNotFound)
	}
	return s.repository.DeleteTask(ctx, scope, id)
}

func validateTask(task model.Task) error {
	if task.Name == "" {
		return fmt.Errorf("task name must not be empty: %w", model.ErrInvalidArgument)
	}
	if task.Estimate != nil && *task.Estimate < 0 {
		return fmt.Errorf("task estimate must not be negative: %w", model.ErrInvalidArgument)
	}
	if task.CloseReason != nil && !task.CloseReason.Valid() {
		return fmt.Errorf("unknown close reason %q: %w", *task.CloseReason, model.ErrInvalidArgument)
	}
	return nil
}

func validateTaskPatch(patch model.TaskPatch) error {
	if patch.Name != nil && *patch.Name == "" {
		return fmt.Errorf("task name must not be empty: %w", model.ErrInvalidArgument)
	}
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
