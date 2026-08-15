package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/larssonoliver/inundated/internal/model"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── CreateSession ────────────────────────────────────────────────────────────

func TestCreateSession_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newSessionMock(t)
	session := aSession()

	mock.ExpectQuery(`INSERT INTO sessions`).
		WithArgs(session.Id, session.UserId, session.Sub, session.ExpiresAt).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "sub", "expires_at"}).
			AddRow(session.Id, session.UserId, session.Sub, session.ExpiresAt))

	got, err := repo.CreateSession(ctx, session)
	require.NoError(t, err)
	assert.Equal(t, session, got)
}

func TestCreateSession_GeneratesIdWhenNil(t *testing.T) {
	ctx := context.Background()
	repo, mock := newSessionMock(t)
	session := aSession()
	session.Id = uuid.Nil

	mock.ExpectQuery(`INSERT INTO sessions`).
		WithArgs(pgxmock.AnyArg(), session.UserId, session.Sub, session.ExpiresAt).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "sub", "expires_at"}).
			AddRow(uuid.New(), session.UserId, session.Sub, session.ExpiresAt))

	got, err := repo.CreateSession(ctx, session)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, got.Id)
}

func TestCreateSession_DuplicateId(t *testing.T) {
	ctx := context.Background()
	repo, mock := newSessionMock(t)
	session := aSession()

	mock.ExpectQuery(`INSERT INTO sessions`).
		WithArgs(session.Id, session.UserId, session.Sub, session.ExpiresAt).
		WillReturnError(&pgconn.PgError{Code: "23505"})

	_, err := repo.CreateSession(ctx, session)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrAlreadyExists))
}

func TestCreateSession_NilUserId(t *testing.T) {
	repo, _ := newSessionMock(t)
	session := aSession()
	session.UserId = uuid.Nil

	_, err := repo.CreateSession(context.Background(), session)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

func TestCreateSession_EmptySub(t *testing.T) {
	repo, _ := newSessionMock(t)
	session := aSession()
	session.Sub = ""

	_, err := repo.CreateSession(context.Background(), session)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

// ── GetSession ───────────────────────────────────────────────────────────────

func TestGetSession_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newSessionMock(t)
	session := aSession()

	mock.ExpectQuery(`SELECT id, user_id, sub, expires_at FROM sessions WHERE id = \$1`).
		WithArgs(session.Id).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "sub", "expires_at"}).
			AddRow(session.Id, session.UserId, session.Sub, session.ExpiresAt))

	got, err := repo.GetSession(ctx, session.Id)
	require.NoError(t, err)
	assert.Equal(t, session, got)
}

func TestGetSession_NotFound(t *testing.T) {
	ctx := context.Background()
	repo, mock := newSessionMock(t)
	id := uuid.New()

	mock.ExpectQuery(`SELECT id, user_id, sub, expires_at FROM sessions WHERE id = \$1`).
		WithArgs(id).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "sub", "expires_at"}))

	_, err := repo.GetSession(ctx, id)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestGetSession_NilId(t *testing.T) {
	repo, _ := newSessionMock(t)

	_, err := repo.GetSession(context.Background(), uuid.Nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

// ── TouchSession ────────────────────────────────────────────────────────────

func TestTouchSession_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newSessionMock(t)

	session := aSession()
	session.ExpiresAt = time.Now().Add(2 * time.Hour).UTC()

	mock.ExpectQuery(`UPDATE sessions .+ WHERE id = \$1`).
		WithArgs(session.Id, session.ExpiresAt).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "sub", "expires_at"}).
			AddRow(session.Id, session.UserId, session.Sub, session.ExpiresAt))

	got, err := repo.TouchSession(ctx, session.Id, session.ExpiresAt)
	require.NoError(t, err)
	assert.Equal(t, session, got)
}

func TestTouchSession_NotFound(t *testing.T) {
	ctx := context.Background()
	repo, mock := newSessionMock(t)
	session := aSession()

	mock.ExpectQuery(`UPDATE sessions .+ WHERE id = \$1`).
		WithArgs(session.Id, session.ExpiresAt).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "sub", "expires_at"}))

	_, err := repo.TouchSession(ctx, session.Id, session.ExpiresAt)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestTouchSession_NilId(t *testing.T) {
	repo, _ := newSessionMock(t)
	session := aSession()
	session.Id = uuid.Nil

	_, err := repo.TouchSession(context.Background(), session.Id, session.ExpiresAt)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

// ── DeleteSession ────────────────────────────────────────────────────────────

func TestDeleteSession_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newSessionMock(t)
	id := uuid.New()

	mock.ExpectExec(`DELETE FROM sessions WHERE id = \$1`).
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	require.NoError(t, repo.DeleteSession(ctx, id))
}

func TestDeleteSession_NotFound(t *testing.T) {
	ctx := context.Background()
	repo, mock := newSessionMock(t)
	id := uuid.New()

	mock.ExpectExec(`DELETE FROM sessions WHERE id = \$1`).
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	err := repo.DeleteSession(ctx, id)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestDeleteSession_NilId(t *testing.T) {
	repo, _ := newSessionMock(t)

	err := repo.DeleteSession(context.Background(), uuid.Nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

func TestDeleteAllExpiredSessions_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newSessionMock(t)

	mock.ExpectExec(`DELETE FROM sessions WHERE expires_at < NOW()`).
		WillReturnResult(pgxmock.NewResult("DELETE", 2))

	err := repo.DeleteAllExpiredSessions(ctx)
	require.NoError(t, err)
}
