package memory

import (
	"context"
	"slices"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/utils"
)

func (t *MemoryStore) tagsExist(ctx context.Context, scope model.OwnerScope, tagIds []uuid.UUID) bool {
	for _, tagId := range tagIds {
		if _, err := t.GetTag(ctx, scope, tagId); err != nil {
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

	return t.tags[idx], nil
}

// ListTags implements [repository.TagRepository].
func (t *MemoryStore) ListTags(ctx context.Context, scope model.OwnerScope, params model.PaginationParams) (model.Page[model.Tag], error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	all := make([]model.Tag, 0, len(t.tags))
	for _, tag := range t.tags {
		if matchesScope(tag.UserId, scope) {
			all = append(all, tag)
		}
	}

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
	t.tags[idx] = tag
	return tag, nil
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
