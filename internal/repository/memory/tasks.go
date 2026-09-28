package memory

import (
	"bytes"
	"cmp"
	"context"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/utils"
)

// GetTask implements [repository.TaskRepository].
func (t *MemoryStore) GetTask(ctx context.Context, scope model.OwnerScope, id uuid.UUID) (model.Task, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	idx := t.taskIndex(scope, id)
	if idx == -1 {
		return model.Task{}, model.ErrNotFound
	}
	return copyTask(t.tasks[idx]), nil
}

// ListTasks implements [repository.TaskRepository].
func (t *MemoryStore) ListTasks(ctx context.Context, scope model.OwnerScope, params model.TaskListParams) (model.Page[model.Task], error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	all := make([]model.Task, 0, len(t.tasks))
	for _, task := range t.tasks {
		if !matchesScope(task.UserId, scope) || (!params.IncludeClosed && task.Closed()) {
			continue
		}
		if params.ParentId != nil && (task.ParentId == nil || *task.ParentId != *params.ParentId) {
			continue
		}
		if params.TagId != nil && !slices.Contains(task.TagIds, *params.TagId) {
			continue
		}
		if params.DueFrom != nil && (task.DueDate == nil || task.DueDate.Before(*params.DueFrom)) {
			continue
		}
		if params.DueTo != nil && (task.DueDate == nil || task.DueDate.After(*params.DueTo)) {
			continue
		}
		all = append(all, copyTask(task))
	}

	// Top-level tasks first, then grouped by parent, then by rank - the
	// order Postgres gives for "ORDER BY parent_id NULLS FIRST, rank, id".
	slices.SortStableFunc(all, func(a, b model.Task) int {
		return cmp.Or(compareParents(a.ParentId, b.ParentId), strings.Compare(a.Rank, b.Rank), bytes.Compare(a.Id[:], b.Id[:]))
	})

	total := len(all)
	start := min(params.Offset, total)
	end := min(start+params.Limit, total)

	return model.Page[model.Task]{
		Data:       all[start:end],
		TotalCount: total,
		Limit:      params.Limit,
		Offset:     params.Offset,
	}, nil
}

// CreateTask implements [repository.TaskRepository].
func (t *MemoryStore) CreateTask(ctx context.Context, scope model.OwnerScope, task model.Task) (model.Task, error) {
	if task.Name == "" {
		return model.Task{}, model.ErrInvalidArgument
	}
	tagIds := utils.DedupeUUIDs(task.TagIds)

	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.tagsUsable(scope, tagIds, nil) || t.anyTaskTag(tagIds) {
		return model.Task{}, model.ErrInvalidReference
	}
	if task.ParentId != nil {
		idx := t.taskIndex(scope, *task.ParentId)
		if idx == -1 {
			return model.Task{}, model.ErrInvalidReference
		}
		if t.tasks[idx].Closed() {
			return model.Task{}, model.ErrInvalidArgument
		}
	}

	siblings := t.siblings(scope, task.ParentId, uuid.Nil)
	last := ""
	if len(siblings) > 0 {
		last = t.tasks[siblings[len(siblings)-1]].Rank
	}
	rank, err := utils.RankBetween(last, "")
	if err != nil || len(rank) > utils.MaxRankLength {
		t.respace(siblings)
		rank, err = utils.RankBetween(t.tasks[siblings[len(siblings)-1]].Rank, "")
		if err != nil {
			return model.Task{}, err
		}
	}

	taskId := uuid.New()
	tag := model.Tag{
		Id:     uuid.New(),
		Name:   task.Name,
		Color:  model.DefaultTaskTagColor,
		UserId: scope.UserID(),
		TaskId: &taskId,
	}
	created := model.Task{
		Id:       taskId,
		Name:     task.Name,
		TagId:    tag.Id,
		ParentId: task.ParentId,
		TagIds:   append([]uuid.UUID{}, tagIds...),
		DueDate:  task.DueDate,
		Estimate: task.Estimate,
		Rank:     rank,
		UserId:   scope.UserID(),
	}

	t.tags = append(t.tags, tag)
	t.tasks = append(t.tasks, created)
	return copyTask(created), nil
}

// UpdateTask implements [repository.TaskRepository].
func (t *MemoryStore) UpdateTask(ctx context.Context, scope model.OwnerScope, task model.Task) (model.Task, error) {
	if task.Name == "" || (task.CloseReason != nil && !task.CloseReason.Valid()) {
		return model.Task{}, model.ErrInvalidArgument
	}

	tagIds := utils.DedupeUUIDs(task.TagIds)

	t.mu.Lock()
	defer t.mu.Unlock()

	idx := t.taskIndex(scope, task.Id)
	if idx == -1 {
		return model.Task{}, model.ErrNotFound
	}
	if !t.tagsUsable(scope, tagIds, t.tasks[idx].TagIds) || t.anyTaskTag(tagIds) {
		return model.Task{}, model.ErrInvalidReference
	}

	stored := &t.tasks[idx]
	wasClosed := stored.Closed()
	stored.Name = task.Name
	stored.TagIds = append([]uuid.UUID{}, tagIds...)
	stored.DueDate = task.DueDate
	stored.Estimate = task.Estimate

	now := time.Now().UTC()
	switch {
	case !wasClosed && task.CloseReason != nil:
		reason := *task.CloseReason
		stored.CloseReason = &reason
		stored.ClosedAt = &now
		for _, i := range t.descendants(scope, stored.Id) {
			if !t.tasks[i].Closed() {
				t.tasks[i].CloseReason = &reason
				t.tasks[i].ClosedAt = &now
			}
		}
	case wasClosed && task.CloseReason == nil:
		stored.CloseReason = nil
		stored.ClosedAt = nil
		for parent := stored.ParentId; parent != nil; {
			i := t.taskIndex(scope, *parent)
			if i == -1 {
				break
			}
			t.tasks[i].CloseReason = nil
			t.tasks[i].ClosedAt = nil
			parent = t.tasks[i].ParentId
		}
	case task.CloseReason != nil:
		reason := *task.CloseReason
		stored.CloseReason = &reason
	}

	if tagIdx := slices.IndexFunc(t.tags, func(tag model.Tag) bool { return tag.Id == stored.TagId }); tagIdx != -1 {
		t.tags[tagIdx].Name = stored.Name
	}

	return copyTask(t.tasks[idx]), nil
}

// MoveTask implements [repository.TaskRepository].
func (t *MemoryStore) MoveTask(ctx context.Context, scope model.OwnerScope, id uuid.UUID, parentId *uuid.UUID, afterId *uuid.UUID) (model.Task, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	idx := t.taskIndex(scope, id)
	if idx == -1 {
		return model.Task{}, model.ErrNotFound
	}

	if parentId != nil {
		parentIdx := t.taskIndex(scope, *parentId)
		if parentIdx == -1 {
			return model.Task{}, model.ErrInvalidArgument
		}
		if t.tasks[parentIdx].Closed() && !t.tasks[idx].Closed() {
			return model.Task{}, model.ErrInvalidArgument
		}
		// The new parent must not be the task itself or sit below it.
		seen := map[uuid.UUID]bool{}
		for ancestor := parentId; ancestor != nil && !seen[*ancestor]; {
			if *ancestor == id {
				return model.Task{}, model.ErrInvalidArgument
			}
			seen[*ancestor] = true
			i := t.taskIndex(scope, *ancestor)
			if i == -1 {
				break
			}
			ancestor = t.tasks[i].ParentId
		}
	}

	siblings := t.siblings(scope, parentId, id)
	pos := 0
	if afterId != nil {
		pos = slices.IndexFunc(siblings, func(i int) bool { return t.tasks[i].Id == *afterId })
		if pos == -1 {
			return model.Task{}, model.ErrInvalidArgument
		}
		pos++
	}

	lo, hi := "", ""
	if pos > 0 {
		lo = t.tasks[siblings[pos-1]].Rank
	}
	if pos < len(siblings) {
		hi = t.tasks[siblings[pos]].Rank
	}
	rank, err := utils.RankBetween(lo, hi)

	t.tasks[idx].ParentId = parentId
	if err != nil || len(rank) > utils.MaxRankLength {
		t.respace(slices.Insert(siblings, pos, idx))
	} else {
		t.tasks[idx].Rank = rank
	}

	return copyTask(t.tasks[idx]), nil
}

// DeleteTask implements [repository.TaskRepository].
func (t *MemoryStore) DeleteTask(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	idx := t.taskIndex(scope, id)
	if idx == -1 {
		return model.ErrNotFound
	}

	taskIds := map[uuid.UUID]bool{id: true}
	tagIds := map[uuid.UUID]bool{t.tasks[idx].TagId: true}
	for _, i := range t.descendants(scope, id) {
		taskIds[t.tasks[i].Id] = true
		tagIds[t.tasks[i].TagId] = true
	}

	for _, span := range t.timespans {
		if !matchesScope(span.UserId, scope) {
			continue
		}
		if slices.ContainsFunc(span.TagIds, func(tagId uuid.UUID) bool { return tagIds[tagId] }) {
			return model.ErrConflict
		}
	}

	t.tasks = slices.DeleteFunc(t.tasks, func(task model.Task) bool { return taskIds[task.Id] })
	t.tags = slices.DeleteFunc(t.tags, func(tag model.Tag) bool { return tagIds[tag.Id] })
	for i := range t.projects {
		t.projects[i].TagIds = slices.DeleteFunc(t.projects[i].TagIds, func(tagId uuid.UUID) bool { return tagIds[tagId] })
	}
	return nil
}

// taskIndex returns the index of the task with id in scope, or -1. Callers
// must hold t.mu.
func (t *MemoryStore) taskIndex(scope model.OwnerScope, id uuid.UUID) int {
	return slices.IndexFunc(t.tasks, func(task model.Task) bool {
		return task.Id == id && matchesScope(task.UserId, scope)
	})
}

// siblings returns the indexes of the tasks under parentId (nil for the top
// level), ordered by rank, leaving out exclude. Callers must hold t.mu.
func (t *MemoryStore) siblings(scope model.OwnerScope, parentId *uuid.UUID, exclude uuid.UUID) []int {
	var idxs []int
	for i, task := range t.tasks {
		if task.Id == exclude || !matchesScope(task.UserId, scope) || compareParents(task.ParentId, parentId) != 0 {
			continue
		}
		idxs = append(idxs, i)
	}
	slices.SortFunc(idxs, func(a, b int) int { return strings.Compare(t.tasks[a].Rank, t.tasks[b].Rank) })
	return idxs
}

// descendants returns the indexes of every task below id. Callers must hold
// t.mu.
func (t *MemoryStore) descendants(scope model.OwnerScope, id uuid.UUID) []int {
	var out []int
	seen := map[uuid.UUID]bool{id: true}
	queue := []uuid.UUID{id}
	for len(queue) > 0 {
		parent := queue[0]
		queue = queue[1:]
		for i, task := range t.tasks {
			if task.ParentId != nil && *task.ParentId == parent && matchesScope(task.UserId, scope) && !seen[task.Id] {
				seen[task.Id] = true
				out = append(out, i)
				queue = append(queue, task.Id)
			}
		}
	}
	return out
}

// respace gives the tasks at idxs fresh, evenly spread ranks in the order
// given. Callers must hold t.mu.
func (t *MemoryStore) respace(idxs []int) {
	for i, rank := range utils.EvenRanks(len(idxs)) {
		t.tasks[idxs[i]].Rank = rank
	}
}

// anyTaskTag reports whether any of ids names a task tag. Callers must hold
// t.mu.
func (t *MemoryStore) anyTaskTag(ids []uuid.UUID) bool {
	return slices.ContainsFunc(ids, t.isTaskTag)
}

func compareParents(a, b *uuid.UUID) int {
	switch {
	case a == nil && b == nil:
		return 0
	case a == nil:
		return -1
	case b == nil:
		return 1
	}
	return bytes.Compare(a[:], b[:])
}

func copyTask(task model.Task) model.Task {
	task.TagIds = append([]uuid.UUID{}, task.TagIds...)
	return task
}
