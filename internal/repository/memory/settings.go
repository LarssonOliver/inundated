package memory

import (
	"context"
	"slices"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

// GetSettings implements [repository.SettingsRepository].
func (m *MemoryStore) GetSettings(ctx context.Context, scope model.OwnerScope) (model.Settings, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	idx := slices.IndexFunc(m.settings, func(s model.Settings) bool { return matchesScope(s.UserId, scope) })
	if idx == -1 {
		return model.Settings{}, model.ErrNotFound
	}

	return m.settings[idx], nil
}

// CreateSettings implements [repository.SettingsRepository].
func (m *MemoryStore) CreateSettings(ctx context.Context, scope model.OwnerScope, settings model.Settings) (model.Settings, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if slices.ContainsFunc(m.settings, func(s model.Settings) bool { return matchesScope(s.UserId, scope) }) {
		return model.Settings{}, model.ErrAlreadyExists
	}

	if settings.Id == uuid.Nil {
		settings.Id = uuid.New()
	}
	settings.UserId = scope.UserID()

	m.settings = append(m.settings, settings)
	return settings, nil
}

// UpdateSettings implements [repository.SettingsRepository].
func (m *MemoryStore) UpdateSettings(ctx context.Context, scope model.OwnerScope, settings model.Settings) (model.Settings, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	idx := slices.IndexFunc(m.settings, func(s model.Settings) bool { return matchesScope(s.UserId, scope) })
	if idx == -1 {
		return model.Settings{}, model.ErrNotFound
	}

	settings.Id = m.settings[idx].Id
	settings.UserId = m.settings[idx].UserId
	m.settings[idx] = settings
	return settings, nil
}
