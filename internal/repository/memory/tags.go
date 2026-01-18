package memory

import (
	"context"
	"sync"

	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
)

type TagStore struct {
	sync.RWMutex
	data map[model.Uuid]model.Tag
}

var _ repository.TagRepository = (*TagStore)(nil)

func NewTagStore() *TagStore {
	return &TagStore{
		data: make(map[model.Uuid]model.Tag),
	}
}

// CreateTag implements [repository.TagRepository].
func (t *TagStore) CreateTag(ctx context.Context, tag *model.Tag) (*model.Tag, error) {
	panic("unimplemented")
}

// DeleteTag implements [repository.TagRepository].
func (t *TagStore) DeleteTag(ctx context.Context, id model.Uuid) error {
	panic("unimplemented")
}

// GetTag implements [repository.TagRepository].
func (t *TagStore) GetTag(ctx context.Context, id model.Uuid) (*model.Tag, error) {
	panic("unimplemented")
}

// ListTags implements [repository.TagRepository].
func (t *TagStore) ListTags(ctx context.Context) ([]*model.Tag, error) {
	panic("unimplemented")
}

// UpdateTag implements [repository.TagRepository].
func (t *TagStore) UpdateTag(ctx context.Context, tag *model.Tag) (*model.Tag, error) {
	panic("unimplemented")
}
