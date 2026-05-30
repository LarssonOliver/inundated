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

	mock.ExpectQuery(`SELECT id, name, color FROM tags WHERE id = \$1 AND deleted_at IS NULL`).
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

	mock.ExpectQuery(`SELECT id, name, color FROM tags WHERE id = \$1 AND deleted_at IS NULL`).
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

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM tags WHERE deleted_at IS NULL`).
		WillReturnRows(
			pgxmock.NewRows([]string{"count"}).
				AddRow(2),
		)

	mock.ExpectQuery(`SELECT id, name, color FROM tags WHERE deleted_at IS NULL ORDER BY name LIMIT \$1 OFFSET \$2`).
		WithArgs(25, 0).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "name", "color"}).
				AddRow(t1.Id, t1.Name, t1.Color).
				AddRow(t2.Id, t2.Name, t2.Color),
		)

	page, err := repo.ListTags(ctx, model.DefaultPaginationParams())
	require.NoError(t, err)

	assert.Len(t, page.Data, 2)
	assert.Equal(t, 2, page.TotalCount)
	assert.Equal(t, t1.Name, page.Data[0].Name)
	assert.Equal(t, t2.Name, page.Data[1].Name)
}

func TestListTags_WithPaginationParams(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)

	tag := aTag()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM tags WHERE deleted_at IS NULL`).
		WillReturnRows(
			pgxmock.NewRows([]string{"count"}).
				AddRow(3),
		)

	mock.ExpectQuery(`SELECT id, name, color FROM tags WHERE deleted_at IS NULL ORDER BY name LIMIT \$1 OFFSET \$2`).
		WithArgs(1, 1).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "name", "color"}).
				AddRow(tag.Id, tag.Name, tag.Color),
		)

	page, err := repo.ListTags(ctx, model.PaginationParams{
		Limit:  1,
		Offset: 1,
	})

	require.NoError(t, err)

	assert.Len(t, page.Data, 1)
	assert.Equal(t, 3, page.TotalCount)
	assert.Equal(t, 1, page.Limit)
	assert.Equal(t, 1, page.Offset)
}

func TestListTags_Empty(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM tags WHERE deleted_at IS NULL`).
		WillReturnRows(
			pgxmock.NewRows([]string{"count"}).
				AddRow(0),
		)

	mock.ExpectQuery(`SELECT id, name, color FROM tags WHERE deleted_at IS NULL ORDER BY name LIMIT \$1 OFFSET \$2`).
		WithArgs(25, 0).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "name", "color"}),
		)

	page, err := repo.ListTags(ctx, model.DefaultPaginationParams())

	require.NoError(t, err)

	assert.Empty(t, page.Data)
	assert.Equal(t, 0, page.TotalCount)
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

	mock.ExpectQuery(`UPDATE tags .+ WHERE .+ deleted_at IS NULL`).
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

	mock.ExpectQuery(`UPDATE tags .+ WHERE .+ deleted_at IS NULL`).
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

	mock.ExpectExec(`UPDATE tags SET deleted_at = now\(\) WHERE .* deleted_at IS NULL`).
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	require.NoError(t, repo.DeleteTag(ctx, id))
}

func TestDeleteTag_NotFound(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	id := uuid.New()

	mock.ExpectExec(`UPDATE tags SET deleted_at = now\(\) WHERE .* deleted_at IS NULL`).
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

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
