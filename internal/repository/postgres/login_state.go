package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/larssonoliver/inundated/internal/model"
)

// CreateLoginState implements [repository.LoginStateRepository].
func (r *PostgresStore) CreateLoginState(ctx context.Context, loginState model.LoginState) (model.LoginState, error) {
	if loginState.RedirectUri == "" {
		return model.LoginState{}, fmt.Errorf("CreateLoginState: redirect_uri must not be empty: %w", model.ErrInvalidArgument)
	}
	if loginState.CodeVerifier == "" {
		return model.LoginState{}, fmt.Errorf("CreateLoginState: code_verifier must not be empty: %w", model.ErrInvalidArgument)
	}
	if loginState.Id == uuid.Nil {
		loginState.Id = uuid.New()
	}

	const q = `
		INSERT INTO login_states (id, redirect_uri, code_verifier, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, redirect_uri, code_verifier, expires_at`

	var created model.LoginState
	err := r.db.QueryRow(ctx, q, loginState.Id, loginState.RedirectUri, loginState.CodeVerifier, loginState.ExpiresAt).
		Scan(&created.Id, &created.RedirectUri, &created.CodeVerifier, &created.ExpiresAt)
	if isUniqueViolation(err) {
		return model.LoginState{}, fmt.Errorf("CreateLoginState %s: %w", loginState.Id, model.ErrAlreadyExists)
	}
	if err != nil {
		return model.LoginState{}, fmt.Errorf("CreateLoginState: %w", err)
	}
	return created, nil
}

// DeleteLoginState implements [repository.LoginStateRepository].
func (r *PostgresStore) DeleteLoginState(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("DeleteLoginState: id: %w", model.ErrInvalidArgument)
	}

	const q = `DELETE FROM login_states WHERE id = $1`
	res, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("DeleteLoginState: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("DeleteLoginState %s: %w", id, model.ErrNotFound)
	}
	return nil
}

// GetLoginState implements [repository.LoginStateRepository].
func (r *PostgresStore) GetLoginState(ctx context.Context, id uuid.UUID) (model.LoginState, error) {
	if id == uuid.Nil {
		return model.LoginState{}, fmt.Errorf("GetLoginState: id: %w", model.ErrInvalidArgument)
	}

	const q = `
		SELECT id, redirect_uri, code_verifier, expires_at
		FROM login_states
		WHERE id = $1`

	var ls model.LoginState
	err := r.db.QueryRow(ctx, q, id).Scan(&ls.Id, &ls.RedirectUri, &ls.CodeVerifier, &ls.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.LoginState{}, fmt.Errorf("GetLoginState %s: %w", id, model.ErrNotFound)
	}
	if err != nil {
		return model.LoginState{}, fmt.Errorf("GetLoginState: %w", err)
	}
	return ls, nil
}

// DeleteAllExpiredLoginStates implements [repository.LoginStateRepository].
func (r *PostgresStore) DeleteAllExpiredLoginStates(ctx context.Context) error {
	const q = "DELETE FROM login_states WHERE expires_at < NOW()"

	_, err := r.db.Exec(ctx, q)

	if err != nil {
		return fmt.Errorf("DeleteAllExpiredLoginStates: %w", err)
	}

	return nil
}
