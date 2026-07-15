package memory

import (
	"context"
	"slices"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

// GetUser implements [repository.UserRepository].
func (m *MemoryStore) GetUser(ctx context.Context, id uuid.UUID) (model.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	idx := slices.IndexFunc(m.users, func(u model.User) bool { return u.Id == id })
	if idx == -1 {
		return model.User{}, model.ErrNotFound
	}

	return m.users[idx], nil
}

// GetUserBySub implements [repository.UserRepository].
func (m *MemoryStore) GetUserBySub(ctx context.Context, sub string) (model.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	id, ok := m.subToID[sub]
	if !ok {
		return model.User{}, model.ErrNotFound
	}

	idx := slices.IndexFunc(m.users, func(u model.User) bool { return u.Id == id })
	if idx == -1 {
		return model.User{}, model.ErrNotFound
	}

	return m.users[idx], nil
}

// CreateUser implements [repository.UserRepository].
func (m *MemoryStore) CreateUser(ctx context.Context, user model.User) (model.User, error) {
	if user.Sub == "" || user.Email == "" {
		return model.User{}, model.ErrInvalidArgument
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for duplicate sub
	if _, exists := m.subToID[user.Sub]; exists {
		return model.User{}, model.ErrAlreadyExists
	}

	m.users = append(m.users, user)
	m.subToID[user.Sub] = user.Id

	return user, nil
}

// UpdateUser implements [repository.UserRepository].
func (m *MemoryStore) UpdateUser(ctx context.Context, user model.User) (model.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	idx := slices.IndexFunc(m.users, func(u model.User) bool { return u.Id == user.Id })
	if idx == -1 {
		return model.User{}, model.ErrNotFound
	}

	user.Sub = m.users[idx].Sub // Sub cannot be updated
	m.users[idx] = user

	return user, nil
}
