package memory

import (
	"context"
	"slices"

	"github.com/larssonoliver/inundated/internal/model"
)

// GetTimespanByCaldavName implements [repository.TimespanRepository].
func (t *MemoryStore) GetTimespanByCaldavName(ctx context.Context, scope model.OwnerScope, name string) (model.Timespan, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	idx := t.timespanNameIndex(scope, &name)
	if idx == -1 {
		return model.Timespan{}, model.ErrNotFound
	}
	return t.timespans[idx], nil
}

// GetTaskByCaldavName implements [repository.TaskRepository].
func (t *MemoryStore) GetTaskByCaldavName(ctx context.Context, scope model.OwnerScope, name string) (model.Task, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	idx := t.taskNameIndex(scope, &name)
	if idx == -1 {
		return model.Task{}, model.ErrNotFound
	}
	return copyTask(t.tasks[idx]), nil
}

// timespanNameIndex returns the index of the timespan in scope created
// under name, or -1 (always for a nil name). Callers must hold t.mu.
func (t *MemoryStore) timespanNameIndex(scope model.OwnerScope, name *string) int {
	if name == nil {
		return -1
	}
	return slices.IndexFunc(t.timespans, func(timespan model.Timespan) bool {
		return timespan.CaldavName != nil && *timespan.CaldavName == *name && matchesScope(timespan.UserId, scope)
	})
}

// taskNameIndex is [MemoryStore.timespanNameIndex] for tasks.
func (t *MemoryStore) taskNameIndex(scope model.OwnerScope, name *string) int {
	if name == nil {
		return -1
	}
	return slices.IndexFunc(t.tasks, func(task model.Task) bool {
		return task.CaldavName != nil && *task.CaldavName == *name && matchesScope(task.UserId, scope)
	})
}
