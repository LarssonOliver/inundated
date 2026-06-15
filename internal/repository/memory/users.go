package memory

import (
	"context"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

// GetByID implements [repository.UserRepository].
func (m *MemoryStore) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	idx := slices.IndexFunc(m.users, func(u *model.User) bool { return u.ID == id })
	if idx == -1 {
		return nil, model.ErrNotFound
	}

	return m.users[idx], nil
}

// GetBySub implements [repository.UserRepository].
func (m *MemoryStore) GetBySub(ctx context.Context, sub string) (*model.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	id, ok := m.subToID[sub]
	if !ok {
		return nil, model.ErrNotFound
	}

	idx := slices.IndexFunc(m.users, func(u *model.User) bool { return u.ID == id })
	if idx == -1 {
		return nil, model.ErrNotFound
	}

	return m.users[idx], nil
}

// Create implements [repository.UserRepository].
func (m *MemoryStore) Create(ctx context.Context, user *model.User) error {
	if user.Sub == "" || user.Email == "" {
		return model.ErrInvalidArgument
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for duplicate sub
	if _, exists := m.subToID[user.Sub]; exists {
		return model.ErrAlreadyExists
	}

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	m.users = append(m.users, user)
	m.subToID[user.Sub] = user.ID

	return nil
}

// Update implements [repository.UserRepository].
func (m *MemoryStore) Update(ctx context.Context, id uuid.UUID, update *model.UpdateUser) (*model.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	idx := slices.IndexFunc(m.users, func(u *model.User) bool { return u.ID == id })
	if idx == -1 {
		return nil, model.ErrNotFound
	}

	user := m.users[idx]

	if update.Email != nil {
		user.Email = *update.Email
	}
	if update.Name != nil {
		user.Name = *update.Name
	}
	user.UpdatedAt = time.Now()

	return user, nil
}
