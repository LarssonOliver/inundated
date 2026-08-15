package memory

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

// CreateLoginState implements [repository.LoginStateRepository].
func (t *MemoryStore) CreateLoginState(ctx context.Context, loginState model.LoginState) (model.LoginState, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if loginState.Id == uuid.Nil {
		loginState.Id = uuid.New()
	}

	for _, ls := range t.loginStates {
		if ls.Id == loginState.Id {
			return model.LoginState{}, model.ErrAlreadyExists
		}
	}

	t.loginStates = append(t.loginStates, loginState)
	return loginState, nil
}

// DeleteLoginState implements [repository.LoginStateRepository].
func (t *MemoryStore) DeleteLoginState(ctx context.Context, id uuid.UUID) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	for i, ls := range t.loginStates {
		if ls.Id == id {
			t.loginStates = append(t.loginStates[:i], t.loginStates[i+1:]...)
			return nil
		}
	}

	return model.ErrNotFound
}

// GetLoginState implements [repository.LoginStateRepository].
func (t *MemoryStore) GetLoginState(ctx context.Context, id uuid.UUID) (model.LoginState, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, ls := range t.loginStates {
		if ls.Id == id {
			return ls, nil
		}
	}

	return model.LoginState{}, model.ErrNotFound
}

// DeleteAllExpiredLoginStates implements [repository.LoginStateRepository].
func (t *MemoryStore) DeleteAllExpiredLoginStates(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	for i, ls := range t.loginStates {
		if ls.ExpiresAt.Before(time.Now()) {
			t.loginStates = append(t.loginStates[:i], t.loginStates[i+1:]...)
		}
	}

	return nil
}
