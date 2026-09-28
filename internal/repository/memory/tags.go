package memory

import (
	"cmp"
	"context"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/utils"
)

// tagsExist reports whether every id refers to a live tag owned by scope.
// An archived tag is only acceptable if it's already in alreadyAssociated
// (i.e. it was attached to this project/timespan before this call) - that
// keeps existing associations with a since-archived tag intact across
// unrelated edits, while still blocking a fresh attachment of an archived
// tag that a picker would never surface.
func (t *MemoryStore) tagsExist(ctx context.Context, scope model.OwnerScope, tagIds []uuid.UUID, alreadyAssociated []uuid.UUID) bool {
	allowedArchived := make(map[uuid.UUID]bool, len(alreadyAssociated))
	for _, id := range alreadyAssociated {
		allowedArchived[id] = true
	}

	for _, tagId := range tagIds {
		tag, err := t.GetTag(ctx, scope, tagId)
		if err != nil {
			return false
		}
		if tag.Archived && !allowedArchived[tagId] {
			return false
		}
	}
	return true
}

// CreateTag implements [repository.TagRepository].
func (t *MemoryStore) CreateTag(ctx context.Context, scope model.OwnerScope, tag model.Tag) (model.Tag, error) {
	if tag.Name == "" || tag.Color == "" || !utils.IsValidColor(tag.Color) {
		return model.Tag{}, model.ErrInvalidArgument
	}

	if tag.Id == uuid.Nil {
		tag.Id = uuid.New()
	}

	newTag := model.Tag{
		Id:     tag.Id,
		Name:   tag.Name,
		Color:  tag.Color,
		UserId: scope.UserID(),
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.tags = append(t.tags, newTag)
	return newTag, nil
}

// GetTag implements [repository.TagRepository].
func (t *MemoryStore) GetTag(ctx context.Context, scope model.OwnerScope, id uuid.UUID) (model.Tag, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	idx := slices.IndexFunc(t.tags, func(tag model.Tag) bool { return tag.Id == id })
	if idx == -1 || !matchesScope(t.tags[idx].UserId, scope) {
		return model.Tag{}, model.ErrNotFound
	}

	return t.viewTag(t.tags[idx]), nil
}

// viewTag returns tag as readers see it: a task tag takes its archived state
// and color from its task. Callers must hold t.mu.
func (t *MemoryStore) viewTag(tag model.Tag) model.Tag {
	if tag.TaskId == nil {
		return tag
	}
	idx := slices.IndexFunc(t.tasks, func(task model.Task) bool { return task.Id == *tag.TaskId })
	if idx == -1 {
		return tag
	}
	task := t.tasks[idx]
	tag.Archived = task.Closed()
	tag.Color = model.DefaultTaskTagColor
	var first *model.Tag
	for _, id := range task.TagIds {
		i := slices.IndexFunc(t.tags, func(regular model.Tag) bool { return regular.Id == id })
		if i == -1 {
			continue
		}
		candidate := t.tags[i]
		if first == nil || candidate.Name < first.Name ||
			(candidate.Name == first.Name && strings.Compare(candidate.Id.String(), first.Id.String()) < 0) {
			first = &candidate
		}
	}
	if first != nil {
		tag.Color = first.Color
	}
	return tag
}

// ListTags implements [repository.TagRepository].
func (t *MemoryStore) ListTags(ctx context.Context, scope model.OwnerScope, params model.TagListParams) (model.Page[model.Tag], error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	query := strings.ToLower(params.Query)
	all := make([]model.Tag, 0, len(t.tags))
	for _, stored := range t.tags {
		tag := t.viewTag(stored)
		if !matchesScope(tag.UserId, scope) || (!params.IncludeArchived && tag.Archived) {
			continue
		}
		switch params.Kind {
		case model.TagKindAll:
		case model.TagKindTask:
			if tag.TaskId == nil {
				continue
			}
		default:
			if tag.TaskId != nil {
				continue
			}
		}
		if query != "" && !strings.Contains(strings.ToLower(tag.Name), query) {
			continue
		}
		all = append(all, tag)
	}

	// Regular tags first, then by name, as the Postgres store orders them.
	slices.SortStableFunc(all, func(a, b model.Tag) int {
		return cmp.Or(
			cmp.Compare(boolRank(a.TaskId != nil), boolRank(b.TaskId != nil)),
			strings.Compare(a.Name, b.Name),
		)
	})

	total := len(all)
	start := min(params.Offset, total)
	end := min(start+params.Limit, total)

	return model.Page[model.Tag]{
		Data:       all[start:end],
		TotalCount: total,
		Limit:      params.Limit,
		Offset:     params.Offset,
	}, nil
}

// UpdateTag implements [repository.TagRepository].
func (t *MemoryStore) UpdateTag(ctx context.Context, scope model.OwnerScope, tag model.Tag) (model.Tag, error) {
	if tag.Name == "" || tag.Color == "" || !utils.IsValidColor(tag.Color) {
		return model.Tag{}, model.ErrInvalidArgument
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	idx := slices.IndexFunc(t.tags, func(existing model.Tag) bool {
		return existing.Id == tag.Id && matchesScope(existing.UserId, scope)
	})
	if idx == -1 {
		return model.Tag{}, model.ErrNotFound
	}

	tag.UserId = t.tags[idx].UserId
	tag.TaskId = t.tags[idx].TaskId
	t.tags[idx] = tag
	return t.viewTag(tag), nil
}

// DeleteTag implements [repository.TagRepository].
func (t *MemoryStore) DeleteTag(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	idx := slices.IndexFunc(t.tags, func(existing model.Tag) bool {
		return existing.Id == id && matchesScope(existing.UserId, scope)
	})
	if idx == -1 {
		return model.ErrNotFound
	}

	t.tags = slices.Delete(t.tags, idx, idx+1)
	return nil
}

func boolRank(b bool) int {
	if b {
		return 1
	}
	return 0
}

// isTaskTag reports whether id names a task tag. Callers must hold t.mu.
func (t *MemoryStore) isTaskTag(id uuid.UUID) bool {
	idx := slices.IndexFunc(t.tags, func(tag model.Tag) bool { return tag.Id == id })
	return idx != -1 && t.tags[idx].TaskId != nil
}
