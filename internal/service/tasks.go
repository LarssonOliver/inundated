package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

func (s *ServiceImpl) GetTask(ctx context.Context, id uuid.UUID) (model.Task, error) {
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
	return task, err
}

func (s *ServiceImpl) ListTasks(ctx context.Context, params model.TaskListParams) (model.Page[model.Task], error) {
	scope, err := ownerScope(ctx)
	if err != nil {
		return model.Page[model.Task]{}, err
	}
	if params.DueFrom != nil && params.DueTo != nil && params.DueFrom.After(*params.DueTo) {
		return model.Page[model.Task]{}, fmt.Errorf("ListTasks: dueFrom is after dueTo: %w", model.ErrInvalidArgument)
	}
	return s.repository.ListTasks(ctx, scope, params)
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

func (s *ServiceImpl) UpdateTask(ctx context.Context, task model.Task) (model.Task, error) {
	scope, err := ownerScope(ctx)
	if err != nil {
		return model.Task{}, err
	}
	if err := validateTask(task); err != nil {
		return model.Task{}, err
	}
	return s.repository.UpdateTask(ctx, scope, task)
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
