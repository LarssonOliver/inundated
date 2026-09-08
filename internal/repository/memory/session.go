package memory

import (
	"context"
	"encoding/hex"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

type storedSession struct {
	session   model.Session
	tokenHash string
}

func hashHex(token string) string {
	return hex.EncodeToString(model.HashSessionToken(token))
}

// CreateSession implements [repository.SessionRepository].
func (t *MemoryStore) CreateSession(ctx context.Context, session model.Session) (model.Session, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if session.Token == "" {
		return model.Session{}, model.ErrInvalidArgument
	}
	if session.Id == uuid.Nil {
		session.Id = uuid.New()
	}
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now()
	}

	hash := hashHex(session.Token)
	for _, s := range t.sessions {
		if s.session.Id == session.Id || s.tokenHash == hash {
			return model.Session{}, model.ErrAlreadyExists
		}
	}

	stored := session
	stored.Token = ""
	t.sessions = append(t.sessions, storedSession{session: stored, tokenHash: hash})
	return session, nil
}

// DeleteSession implements [repository.SessionRepository].
func (t *MemoryStore) DeleteSession(ctx context.Context, id uuid.UUID) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	for i, s := range t.sessions {
		if s.session.Id == id {
			t.sessions = append(t.sessions[:i], t.sessions[i+1:]...)
			return nil
		}
	}

	return model.ErrNotFound
}

// GetSessionByToken implements [repository.SessionRepository].
func (t *MemoryStore) GetSessionByToken(ctx context.Context, token string) (model.Session, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if token == "" {
		return model.Session{}, model.ErrInvalidArgument
	}

	hash := hashHex(token)
	for _, s := range t.sessions {
		if s.tokenHash == hash {
			return s.session, nil
		}
	}

	return model.Session{}, model.ErrNotFound
}

// TouchSession implements [repository.SessionRepository].
func (t *MemoryStore) TouchSession(ctx context.Context, id uuid.UUID, expiresAt time.Time) (model.Session, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for i, s := range t.sessions {
		if s.session.Id == id {
			t.sessions[i].session.ExpiresAt = expiresAt
			return t.sessions[i].session, nil
		}
	}

	return model.Session{}, model.ErrNotFound
}

// DeleteAllExpiredSessions implements [repository.SessionRepository].
func (t *MemoryStore) DeleteAllExpiredSessions(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	t.sessions = slices.DeleteFunc(t.sessions, func(s storedSession) bool {
		return s.session.ExpiresAt.Before(now)
	})

	return nil
}
