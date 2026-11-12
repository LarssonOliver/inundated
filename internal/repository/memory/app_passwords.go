package memory

import (
	"context"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

type storedAppPassword struct {
	password  model.AppPassword
	tokenHash string
}

// ListAppPasswords implements [repository.AppPasswordRepository].
func (m *MemoryStore) ListAppPasswords(ctx context.Context, scope model.OwnerScope) ([]model.AppPassword, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	passwords := []model.AppPassword{}
	for _, stored := range m.appPasswords {
		if matchesScope(stored.password.UserId, scope) {
			passwords = append(passwords, stored.password)
		}
	}
	slices.SortStableFunc(passwords, func(a, b model.AppPassword) int { return b.CreatedAt.Compare(a.CreatedAt) })
	return passwords, nil
}

// CreateAppPassword implements [repository.AppPasswordRepository].
func (m *MemoryStore) CreateAppPassword(ctx context.Context, scope model.OwnerScope, password model.AppPassword, token string) (model.AppPassword, error) {
	if token == "" {
		return model.AppPassword{}, model.ErrInvalidArgument
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if password.Id == uuid.Nil {
		password.Id = uuid.New()
	}
	if password.CreatedAt.IsZero() {
		password.CreatedAt = time.Now()
	}
	password.UserId = scope.UserID()
	password.LastUsedAt = nil

	hash := hashHex(token)
	for _, stored := range m.appPasswords {
		if stored.password.Id == password.Id || stored.tokenHash == hash {
			return model.AppPassword{}, model.ErrAlreadyExists
		}
	}

	m.appPasswords = append(m.appPasswords, storedAppPassword{password: password, tokenHash: hash})
	return password, nil
}

// DeleteAppPassword implements [repository.AppPasswordRepository].
func (m *MemoryStore) DeleteAppPassword(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	idx := slices.IndexFunc(m.appPasswords, func(s storedAppPassword) bool {
		return s.password.Id == id && matchesScope(s.password.UserId, scope)
	})
	if idx == -1 {
		return model.ErrNotFound
	}

	m.appPasswords = slices.Delete(m.appPasswords, idx, idx+1)
	return nil
}

// GetAppPasswordByToken implements [repository.AppPasswordRepository].
func (m *MemoryStore) GetAppPasswordByToken(ctx context.Context, token string) (model.AppPassword, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	hash := hashHex(token)
	idx := slices.IndexFunc(m.appPasswords, func(s storedAppPassword) bool { return s.tokenHash == hash })
	if idx == -1 {
		return model.AppPassword{}, model.ErrNotFound
	}
	return m.appPasswords[idx].password, nil
}

// TouchAppPassword implements [repository.AppPasswordRepository].
func (m *MemoryStore) TouchAppPassword(ctx context.Context, id uuid.UUID, usedAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	idx := slices.IndexFunc(m.appPasswords, func(s storedAppPassword) bool { return s.password.Id == id })
	if idx == -1 {
		return model.ErrNotFound
	}
	m.appPasswords[idx].password.LastUsedAt = &usedAt
	return nil
}
