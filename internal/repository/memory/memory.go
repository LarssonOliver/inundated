package memory

import "github.com/larssonoliver/inundated/internal/repository"

type MemoryStore struct {
	*TagStore
	*ProjectStore
	*TimeSpanStore
}

var _ repository.Repository = (*MemoryStore)(nil)

func NewMemoryStore() *MemoryStore {
	tagStore := NewTagStore()
	return &MemoryStore{
		TagStore:      tagStore,
		ProjectStore:  NewProjectStore(tagStore),
		TimeSpanStore: NewTimeSpanStore(tagStore),
	}
}
