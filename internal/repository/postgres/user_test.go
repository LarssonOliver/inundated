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

// ── GetUser ──────────────────────────────────────────────────────────────────

func TestGetUser_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	user := aUser()

	mock.ExpectQuery(`SELECT id, sub, email, name FROM users WHERE id = \$1`).
		WithArgs(user.Id).
		WillReturnRows(pgxmock.NewRows([]string{"id", "sub", "email", "name"}).
			AddRow(user.Id, user.Sub, user.Email, user.Name))

	got, err := repo.GetUser(ctx, user.Id)
	require.NoError(t, err)
	assert.Equal(t, user, got)
}

func TestGetUser_NotFound(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	id := uuid.New()

	mock.ExpectQuery(`SELECT id, sub, email, name FROM users WHERE id = \$1`).
		WithArgs(id).
		WillReturnRows(pgxmock.NewRows([]string{"id", "sub", "email", "name"}))

	_, err := repo.GetUser(ctx, id)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestGetUser_NilId(t *testing.T) {
	repo, _ := newMock(t)
	_, err := repo.GetUser(context.Background(), uuid.Nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

// ── GetUserBySub ─────────────────────────────────────────────────────────────

func TestGetUserBySub_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	user := aUser()

	mock.ExpectQuery(`SELECT id, sub, email, name FROM users WHERE sub = \$1`).
		WithArgs(user.Sub).
		WillReturnRows(pgxmock.NewRows([]string{"id", "sub", "email", "name"}).
			AddRow(user.Id, user.Sub, user.Email, user.Name))

	got, err := repo.GetUserBySub(ctx, user.Sub)
	require.NoError(t, err)
	assert.Equal(t, user, got)
}

func TestGetUserBySub_NotFound(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)

	mock.ExpectQuery(`SELECT id, sub, email, name FROM users WHERE sub = \$1`).
		WithArgs("ghost|999").
		WillReturnRows(pgxmock.NewRows([]string{"id", "sub", "email", "name"}))

	_, err := repo.GetUserBySub(ctx, "ghost|999")
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestGetUserBySub_EmptySub(t *testing.T) {
	repo, _ := newMock(t)
	_, err := repo.GetUserBySub(context.Background(), "")
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

// ── CreateUser ───────────────────────────────────────────────────────────────

func TestCreateUser_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	user := aUser()

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(user.Id, user.Sub, user.Email, user.Name).
		WillReturnRows(pgxmock.NewRows([]string{"id", "sub", "email", "name"}).
			AddRow(user.Id, user.Sub, user.Email, user.Name))

	got, err := repo.CreateUser(ctx, user)
	require.NoError(t, err)
	assert.Equal(t, user, got)
}

func TestCreateUser_GeneratesIdWhenNil(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	user := aUser()
	user.Id = uuid.Nil

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(pgxmock.AnyArg(), user.Sub, user.Email, user.Name).
		WillReturnRows(pgxmock.NewRows([]string{"id", "sub", "email", "name"}).
			AddRow(uuid.New(), user.Sub, user.Email, user.Name))

	got, err := repo.CreateUser(ctx, user)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, got.Id)
}

func TestCreateUser_EmptySub(t *testing.T) {
	repo, _ := newMock(t)
	user := aUser()
	user.Sub = ""
	_, err := repo.CreateUser(context.Background(), user)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

func TestCreateUser_EmptyEmail(t *testing.T) {
	repo, _ := newMock(t)
	user := aUser()
	user.Email = ""
	_, err := repo.CreateUser(context.Background(), user)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

// ── CreateUserAdoptingOrphans ────────────────────────────────────────────────

func adoptionRows(id uuid.UUID, sub, email, name string, projects, tags, timespans int) *pgxmock.Rows {
	return pgxmock.NewRows([]string{"id", "sub", "email", "name", "projects", "tags", "timespans"}).
		AddRow(id, sub, email, name, projects, tags, timespans)
}

func TestCreateUserAdoptingOrphans_FirstUserReportsAdoptedCounts(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	user := aUser()

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(user.Id, user.Sub, user.Email, user.Name).
		WillReturnRows(adoptionRows(user.Id, user.Sub, user.Email, user.Name, 3, 2, 5))

	got, adoption, err := repo.CreateUserAdoptingOrphans(ctx, user)
	require.NoError(t, err)
	assert.Equal(t, user, got)
	assert.Equal(t, model.OrphanAdoption{Projects: 3, Tags: 2, Timespans: 5}, adoption)
}

func TestCreateUserAdoptingOrphans_GeneratesIdWhenNil(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	user := aUser()
	user.Id = uuid.Nil

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(pgxmock.AnyArg(), user.Sub, user.Email, user.Name).
		WillReturnRows(adoptionRows(uuid.New(), user.Sub, user.Email, user.Name, 0, 0, 0))

	got, _, err := repo.CreateUserAdoptingOrphans(ctx, user)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, got.Id)
}

func TestCreateUserAdoptingOrphans_DuplicateSub(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	user := aUser()

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(user.Id, user.Sub, user.Email, user.Name).
		WillReturnError(&pgconn.PgError{Code: "23505"})

	_, _, err := repo.CreateUserAdoptingOrphans(ctx, user)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrAlreadyExists))
}

func TestCreateUserAdoptingOrphans_EmptySub(t *testing.T) {
	repo, _ := newMock(t)
	user := aUser()
	user.Sub = ""
	_, _, err := repo.CreateUserAdoptingOrphans(context.Background(), user)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

func TestCreateUserAdoptingOrphans_EmptyEmail(t *testing.T) {
	repo, _ := newMock(t)
	user := aUser()
	user.Email = ""
	_, _, err := repo.CreateUserAdoptingOrphans(context.Background(), user)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

// ── UpdateUser ───────────────────────────────────────────────────────────────

func TestUpdateUser_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	user := aUser()
	user.Email = "updated@example.com"

	mock.ExpectQuery(`UPDATE users .+ WHERE .+`).
		WithArgs(user.Id, user.Email, user.Name).
		WillReturnRows(pgxmock.NewRows([]string{"id", "sub", "email", "name"}).
			AddRow(user.Id, user.Sub, user.Email, user.Name))

	got, err := repo.UpdateUser(ctx, user)
	require.NoError(t, err)
	assert.Equal(t, user, got)
}

func TestUpdateUser_NotFound(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	user := aUser()

	mock.ExpectQuery(`UPDATE users .+ WHERE .+`).
		WithArgs(user.Id, user.Email, user.Name).
		WillReturnRows(pgxmock.NewRows([]string{"id", "sub", "email", "name"}))

	_, err := repo.UpdateUser(ctx, user)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestUpdateUser_NilId(t *testing.T) {
	repo, _ := newMock(t)
	user := aUser()
	user.Id = uuid.Nil
	_, err := repo.UpdateUser(context.Background(), user)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

func TestUpdateUser_EmptyEmail(t *testing.T) {
	repo, _ := newMock(t)
	user := aUser()
	user.Email = ""
	_, err := repo.UpdateUser(context.Background(), user)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}
