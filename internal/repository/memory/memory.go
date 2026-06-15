package memory

import (
	"sync"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
)

type MemoryStore struct {
	mu        sync.RWMutex
	users     []*model.User
	subToID   map[string]uuid.UUID // mapping from sub to user ID for efficient lookups
	projects  []model.Project
	tags      []model.Tag
	timespans []model.Timespan
}

var _ repository.Repository = (*MemoryStore)(nil)

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		mu:        sync.RWMutex{},
		users:     []*model.User{},
		subToID:   make(map[string]uuid.UUID),
		projects:  []model.Project{},
		tags:      []model.Tag{},
		timespans: []model.Timespan{},
	}
}
