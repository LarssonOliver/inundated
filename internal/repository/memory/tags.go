package memory

import (
	"context"

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

	t.tags[newTag.Id] = newTag
	return newTag, nil
}

// GetTag implements [repository.TagRepository].
func (t *MemoryStore) GetTag(ctx context.Context, id uuid.UUID) (model.Tag, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	tag, exists := t.tags[id]
	if !exists {
		return model.Tag{}, model.ErrNotFound
	}

	return tag, nil
}

// ListTags implements [repository.TagRepository].
func (t *MemoryStore) ListTags(ctx context.Context) ([]model.Tag, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	tags := make([]model.Tag, 0, len(t.tags))

	for _, tag := range t.tags {
		tags = append(tags, tag)
	}

	return tags, nil
}

// UpdateTag implements [repository.TagRepository].
func (t *MemoryStore) UpdateTag(ctx context.Context, tag model.Tag) (model.Tag, error) {
	if tag.Name == "" || tag.Color == "" || !utils.IsValidColor(tag.Color) {
		return model.Tag{}, model.ErrInvalidArgument
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	_, exists := t.tags[tag.Id]
	if !exists {
		return model.Tag{}, model.ErrNotFound
	}

	t.tags[tag.Id] = tag
	return tag, nil
}

// DeleteTag implements [repository.TagRepository].
func (t *MemoryStore) DeleteTag(ctx context.Context, id uuid.UUID) error {
	// Skip locking for write if the tag does not exist
	t.mu.RLock()
	_, exists := t.tags[id]
	t.mu.RUnlock()

	if !exists {
		return model.ErrNotFound
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// delete is a noop if the key does not exist
	// thus, it does not matter if it has been deleted by another thread before this line
	delete(t.tags, id)
	return nil
}
