package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── GetTag ───────────────────────────────────────────────────────────────────

func TestGetTag_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	tag := aTag()

	mock.ExpectQuery(`SELECT id, name, color FROM tags WHERE id = \$1`).
		WithArgs(tag.Id).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "color"}).
			AddRow(tag.Id, tag.Name, tag.Color))

	got, err := repo.GetTag(ctx, tag.Id)
	require.NoError(t, err)
	assert.Equal(t, tag, got)
}

func TestGetTag_NotFound(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	id := uuid.New()

	mock.ExpectQuery(`SELECT id, name, color FROM tags WHERE id = \$1`).
		WithArgs(id).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "color"}))

	_, err := repo.GetTag(ctx, id)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestGetTag_NilId(t *testing.T) {
	repo, _ := newMock(t)
	_, err := repo.GetTag(context.Background(), uuid.Nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

// ── ListTags ─────────────────────────────────────────────────────────────────

func TestListTags_ReturnsSorted(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)

	t1, t2 := aTag(), aTag()
	t1.Name, t2.Name = "aaa", "zzz"

	mock.ExpectQuery(`SELECT id, name, color FROM tags ORDER BY name`).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "color"}).
			AddRow(t1.Id, t1.Name, t1.Color).
			AddRow(t2.Id, t2.Name, t2.Color))

	got, err := repo.ListTags(ctx)
	require.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, t1.Name, got[0].Name)
	assert.Equal(t, t2.Name, got[1].Name)
}

func TestListTags_Empty(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)

	mock.ExpectQuery(`SELECT id, name, color FROM tags ORDER BY name`).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "color"}))

	got, err := repo.ListTags(ctx)
	require.NoError(t, err)
	assert.Empty(t, got)
}

// ── CreateTag ────────────────────────────────────────────────────────────────

func TestCreateTag_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	tag := aTag()

	mock.ExpectQuery(`INSERT INTO tags`).
		WithArgs(tag.Id, tag.Name, tag.Color).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "color"}).
			AddRow(tag.Id, tag.Name, tag.Color))

	got, err := repo.CreateTag(ctx, tag)
	require.NoError(t, err)
	assert.Equal(t, tag, got)
}

func TestCreateTag_GeneratesIdWhenNil(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)

	tag := aTag()
	tag.Id = uuid.Nil

	mock.ExpectQuery(`INSERT INTO tags`).
		WithArgs(pgxmock.AnyArg(), tag.Name, tag.Color).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "color"}).
			AddRow(uuid.New(), tag.Name, tag.Color))

	// Pass in a tag with no Id; the implementation must assign one.
	got, err := repo.CreateTag(ctx, tag)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, got.Id)
}

func TestCreateTag_EmptyName(t *testing.T) {
	repo, _ := newMock(t)
	tag := aTag()
	tag.Name = ""
	_, err := repo.CreateTag(context.Background(), tag)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

// ── UpdateTag ────────────────────────────────────────────────────────────────

func TestUpdateTag_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	tag := aTag()
	tag.Name = "updated-name"

	mock.ExpectQuery(`UPDATE tags`).
		WithArgs(tag.Id, tag.Name, tag.Color).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "color"}).
			AddRow(tag.Id, tag.Name, tag.Color))

	got, err := repo.UpdateTag(ctx, tag)
	require.NoError(t, err)
	assert.Equal(t, tag, got)
}

func TestUpdateTag_NotFound(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	tag := aTag()

	mock.ExpectQuery(`UPDATE tags`).
		WithArgs(tag.Id, tag.Name, tag.Color).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "color"}))

	_, err := repo.UpdateTag(ctx, tag)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestUpdateTag_NilId(t *testing.T) {
	repo, _ := newMock(t)
	tag := aTag()
	tag.Id = uuid.Nil
	_, err := repo.UpdateTag(context.Background(), tag)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

func TestUpdateTag_EmptyName(t *testing.T) {
	repo, _ := newMock(t)
	tag := aTag()
	tag.Name = ""
	_, err := repo.UpdateTag(context.Background(), tag)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

// ── DeleteTag ────────────────────────────────────────────────────────────────

func TestDeleteTag_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	id := uuid.New()

	mock.ExpectExec(`DELETE FROM tags WHERE id = \$1`).
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	require.NoError(t, repo.DeleteTag(ctx, id))
}

func TestDeleteTag_NotFound(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	id := uuid.New()

	mock.ExpectExec(`DELETE FROM tags WHERE id = \$1`).
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	err := repo.DeleteTag(ctx, id)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestDeleteTag_NilId(t *testing.T) {
	repo, _ := newMock(t)
	err := repo.DeleteTag(context.Background(), uuid.Nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

