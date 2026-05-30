package memory

import (
	"context"
	"slices"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/utils"
)

func (t *MemoryStore) tagsExist(ctx context.Context, tagIds []uuid.UUID) bool {
	for _, tagId := range tagIds {
		if _, err := t.GetTag(ctx, tagId); err != nil {
			return false
		}
	}
	return true
}

// CreateTag implements [repository.TagRepository].
func (t *MemoryStore) CreateTag(ctx context.Context, tag model.Tag) (model.Tag, error) {
	if tag.Name == "" || tag.Color == "" || !utils.IsValidColor(tag.Color) {
		return model.Tag{}, model.ErrInvalidArgument
	}

	if tag.Id == uuid.Nil {
		tag.Id = uuid.New()
	}

	newTag := model.Tag{
		Id:    tag.Id,
		Name:  tag.Name,
		Color: tag.Color,
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.tags = append(t.tags, newTag)
	return newTag, nil
}

// GetTag implements [repository.TagRepository].
func (t *MemoryStore) GetTag(ctx context.Context, id uuid.UUID) (model.Tag, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	idx := slices.IndexFunc(t.tags, func(tag model.Tag) bool { return tag.Id == id })
	if idx == -1 {
		return model.Tag{}, model.ErrNotFound
	}

	return t.tags[idx], nil
}

// ListTags implements [repository.TagRepository].
func (t *MemoryStore) ListTags(ctx context.Context, params model.PaginationParams) (model.Page[model.Tag], error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	all := make([]model.Tag, 0, len(t.tags))
	all = append(all, t.tags...)

	total := len(all)
	start := min(params.Offset, total)
	end := min(start + params.Limit, total)

	return model.Page[model.Tag]{
		Data:       all[start:end],
		TotalCount: total,
		Limit:      params.Limit,
		Offset:     params.Offset,
	}, nil
}

// UpdateTag implements [repository.TagRepository].
func (t *MemoryStore) UpdateTag(ctx context.Context, tag model.Tag) (model.Tag, error) {
	if tag.Name == "" || tag.Color == "" || !utils.IsValidColor(tag.Color) {
		return model.Tag{}, model.ErrInvalidArgument
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	idx := slices.IndexFunc(t.tags, func(t model.Tag) bool { return t.Id == tag.Id })
	if idx == -1 {
		return model.Tag{}, model.ErrNotFound
	}

	t.tags[idx] = tag
	return tag, nil
}

// DeleteTag implements [repository.TagRepository].
func (t *MemoryStore) DeleteTag(ctx context.Context, id uuid.UUID) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	idx := slices.IndexFunc(t.tags, func(t model.Tag) bool { return t.Id == id })
	if idx == -1 {
		return model.ErrNotFound
	}

	t.tags = slices.Delete(t.tags, idx, idx+1)
	return nil
}
