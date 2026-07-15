package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/larssonoliver/inundated/internal/model"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── CreateLoginState ─────────────────────────────────────────────────────────

func TestCreateLoginState_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newLoginStateMock(t)
	loginState := aLoginState()

	mock.ExpectQuery(`INSERT INTO login_states`).
		WithArgs(loginState.Id, loginState.RedirectUri, loginState.CodeVerifier, loginState.ExpiresAt).
		WillReturnRows(pgxmock.NewRows([]string{"id", "redirect_uri", "code_verifier", "expires_at"}).
			AddRow(loginState.Id, loginState.RedirectUri, loginState.CodeVerifier, loginState.ExpiresAt))

	got, err := repo.CreateLoginState(ctx, loginState)
	require.NoError(t, err)
	assert.Equal(t, loginState, got)
}

func TestCreateLoginState_GeneratesIdWhenNil(t *testing.T) {
	ctx := context.Background()
	repo, mock := newLoginStateMock(t)
	loginState := aLoginState()
	loginState.Id = uuid.Nil

	mock.ExpectQuery(`INSERT INTO login_states`).
		WithArgs(pgxmock.AnyArg(), loginState.RedirectUri, loginState.CodeVerifier, loginState.ExpiresAt).
		WillReturnRows(pgxmock.NewRows([]string{"id", "redirect_uri", "code_verifier", "expires_at"}).
			AddRow(uuid.New(), loginState.RedirectUri, loginState.CodeVerifier, loginState.ExpiresAt))

	got, err := repo.CreateLoginState(ctx, loginState)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, got.Id)
}

func TestCreateLoginState_DuplicateId(t *testing.T) {
	ctx := context.Background()
	repo, mock := newLoginStateMock(t)
	loginState := aLoginState()

	mock.ExpectQuery(`INSERT INTO login_states`).
		WithArgs(loginState.Id, loginState.RedirectUri, loginState.CodeVerifier, loginState.ExpiresAt).
		WillReturnError(&pgconn.PgError{Code: "23505"})

	_, err := repo.CreateLoginState(ctx, loginState)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrAlreadyExists))
}

func TestCreateLoginState_EmptyRedirectUri(t *testing.T) {
	repo, _ := newLoginStateMock(t)
	loginState := aLoginState()
	loginState.RedirectUri = ""

	_, err := repo.CreateLoginState(context.Background(), loginState)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

func TestCreateLoginState_EmptyCodeVerifier(t *testing.T) {
	repo, _ := newLoginStateMock(t)
	loginState := aLoginState()
	loginState.CodeVerifier = ""

	_, err := repo.CreateLoginState(context.Background(), loginState)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

// ── GetLoginState ────────────────────────────────────────────────────────────

func TestGetLoginState_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newLoginStateMock(t)
	loginState := aLoginState()

	mock.ExpectQuery(`SELECT id, redirect_uri, code_verifier, expires_at FROM login_states WHERE id = \$1`).
		WithArgs(loginState.Id).
		WillReturnRows(pgxmock.NewRows([]string{"id", "redirect_uri", "code_verifier", "expires_at"}).
			AddRow(loginState.Id, loginState.RedirectUri, loginState.CodeVerifier, loginState.ExpiresAt))

	got, err := repo.GetLoginState(ctx, loginState.Id)
	require.NoError(t, err)
	assert.Equal(t, loginState, got)
}

func TestGetLoginState_NotFound(t *testing.T) {
	ctx := context.Background()
	repo, mock := newLoginStateMock(t)
	id := uuid.New()

	mock.ExpectQuery(`SELECT id, redirect_uri, code_verifier, expires_at FROM login_states WHERE id = \$1`).
		WithArgs(id).
		WillReturnRows(pgxmock.NewRows([]string{"id", "redirect_uri", "code_verifier", "expires_at"}))

	_, err := repo.GetLoginState(ctx, id)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestGetLoginState_NilId(t *testing.T) {
	repo, _ := newLoginStateMock(t)
	_, err := repo.GetLoginState(context.Background(), uuid.Nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

// ── DeleteLoginState ─────────────────────────────────────────────────────────

func TestDeleteLoginState_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newLoginStateMock(t)
	id := uuid.New()

	mock.ExpectExec(`DELETE FROM login_states WHERE id = \$1`).
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	require.NoError(t, repo.DeleteLoginState(ctx, id))
}

func TestDeleteLoginState_NotFound(t *testing.T) {
	ctx := context.Background()
	repo, mock := newLoginStateMock(t)
	id := uuid.New()

	mock.ExpectExec(`DELETE FROM login_states WHERE id = \$1`).
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	err := repo.DeleteLoginState(ctx, id)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestDeleteLoginState_NilId(t *testing.T) {
	repo, _ := newLoginStateMock(t)
	err := repo.DeleteLoginState(context.Background(), uuid.Nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}
