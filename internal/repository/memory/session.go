package memory

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

// CreateSession implements [repository.SessionRepository].
func (t *MemoryStore) CreateSession(ctx context.Context, session model.Session) (model.Session, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if session.Id == uuid.Nil {
		session.Id = uuid.New()
	}

	for _, s := range t.sessions {
		if s.Id == session.Id {
			return model.Session{}, model.ErrAlreadyExists
		}
	}

	t.sessions = append(t.sessions, session)
	return session, nil
}

// DeleteSession implements [repository.SessionRepository].
func (t *MemoryStore) DeleteSession(ctx context.Context, id uuid.UUID) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	for i, s := range t.sessions {
		if s.Id == id {
			t.sessions = append(t.sessions[:i], t.sessions[i+1:]...)
			return nil
		}
	}

	return model.ErrNotFound
}

// GetSession implements [repository.SessionRepository].
func (t *MemoryStore) GetSession(ctx context.Context, id uuid.UUID) (model.Session, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, s := range t.sessions {
		if s.Id == id {
			return s, nil
		}
	}

	return model.Session{}, model.ErrNotFound
}

// TouchSession implements [repository.SessionRepository].
func (t *MemoryStore) TouchSession(ctx context.Context, id uuid.UUID, expiresAt time.Time) (model.Session, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for i, s := range t.sessions {
		if s.Id == id {
			t.sessions[i].ExpiresAt = expiresAt
			return t.sessions[i], nil
		}
	}

	return model.Session{}, model.ErrNotFound
}

// DeleteAllExpiredSessions implements [repository.SessionRepository].
func (t *MemoryStore) DeleteAllExpiredSessions(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	for i, s := range t.sessions {
		if s.ExpiresAt.Before(time.Now()) {
			t.sessions = append(t.sessions[:i], t.sessions[i+1:]...)
		}
	}

	return nil
}
