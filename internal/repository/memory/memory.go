package memory

import (
	"sync"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
)

type MemoryStore struct {
	mu        sync.RWMutex
	projects  map[uuid.UUID]model.Project
	tags      map[uuid.UUID]model.Tag
	timespans map[uuid.UUID]model.Timespan
}

var _ repository.Repository = (*MemoryStore)(nil)

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		mu:        sync.RWMutex{},
		projects:  make(map[uuid.UUID]model.Project),
		tags:      make(map[uuid.UUID]model.Tag),
		timespans: make(map[uuid.UUID]model.Timespan),
	}
}
