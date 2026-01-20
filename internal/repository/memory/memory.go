package memory

import "github.com/larssonoliver/inundated/internal/repository"

type MemoryStore struct {
	TagStore
	ProjectStore
	TimeSpanStore
}

var _ repository.Repository = (*MemoryStore)(nil)

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		TagStore:      *NewTagStore(),
		ProjectStore:  *NewProjectStore(),
		TimeSpanStore: *NewTimeSpanStore(),
	}
}
