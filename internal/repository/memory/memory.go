package memory

import (
	"sync"

	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
)

type MemoryStore struct {
	mu        sync.RWMutex
	projects  []model.Project
	tags      []model.Tag
	timespans []model.Timespan
}

var _ repository.Repository = (*MemoryStore)(nil)

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		mu:        sync.RWMutex{},
		projects:  []model.Project{},
		tags:      []model.Tag{},
		timespans: []model.Timespan{},
	}
}
