package memory

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/helpers"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
)

type TagStore struct {
	mu   sync.RWMutex
	data map[uuid.UUID]model.Tag
}

var _ repository.TagRepository = (*TagStore)(nil)

func NewTagStore() *TagStore {
	return &TagStore{
		mu:   sync.RWMutex{},
		data: make(map[uuid.UUID]model.Tag),
	}
}

// CreateTag implements [repository.TagRepository].
func (t *TagStore) CreateTag(ctx context.Context, tag model.Tag) (model.Tag, error) {
	if tag.Name == "" {
		return model.Tag{}, errors.New("tag name cannot be empty")
	}

	if tag.Color == "" {
		return model.Tag{}, errors.New("tag color cannot be empty")
	}

	if !helpers.IsValidColor(tag.Color) {
		return model.Tag{}, errors.New("tag color is not a valid hex color")
	}

	newId := uuid.New()
	newTag := model.Tag{
		Id:    newId,
		Name:  tag.Name,
		Color: tag.Color,
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.data[newId] = newTag
	return newTag, nil
}

// GetTag implements [repository.TagRepository].
func (t *TagStore) GetTag(ctx context.Context, id uuid.UUID) (model.Tag, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	tag, exists := t.data[id]
	if !exists {
		return model.Tag{}, errors.New("tag not found")
	}

	return tag, nil
}

// ListTags implements [repository.TagRepository].
func (t *TagStore) ListTags(ctx context.Context) ([]model.Tag, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	tags := make([]model.Tag, 0, len(t.data))

	for _, tag := range t.data {
		tags = append(tags, tag)
	}

	return tags, nil
}

// UpdateTag implements [repository.TagRepository].
func (t *TagStore) UpdateTag(ctx context.Context, tag model.Tag) (model.Tag, error) {
	if tag.Name == "" {
		return model.Tag{}, errors.New("tag name cannot be empty")
	}

	if tag.Color == "" {
		return model.Tag{}, errors.New("tag color cannot be empty")
	}

	if !helpers.IsValidColor(tag.Color) {
		return model.Tag{}, errors.New("tag color is not a valid hex color")
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	_, exists := t.data[tag.Id]
	if !exists {
		return model.Tag{}, errors.New("tag not found")
	}

	t.data[tag.Id] = tag
	return tag, nil
}

// DeleteTag implements [repository.TagRepository].
func (t *TagStore) DeleteTag(ctx context.Context, id uuid.UUID) error {
	// Skip locking for write if the tag does not exist
	t.mu.RLock()
	_, exists := t.data[id]
	t.mu.RUnlock()

	if !exists {
		return errors.New("tag not found")
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// delete is a noop if the key does not exist
	// thus, it does not matter if it has been deleted by another thread before this line
	delete(t.data, id)
	return nil
}
