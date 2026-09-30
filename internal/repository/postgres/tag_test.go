package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tagCols is the ordered column list returned by CreateTag, which does not
// report archived_at (a newly created tag is never archived).
var tagCols = []string{"id", "name", "color", "user_id"}

// tagColsArchived is the ordered column list returned by Get/List/Update,
// which additionally report the archived_at column.
var tagColsArchived = []string{"id", "name", "color", "user_id", "archived_at"}

// tagColsRead is the ordered column list returned by Get/List, which read
// through the tag_owners join and also report the owner's kind and id.
var tagColsRead = []string{"id", "name", "color", "user_id", "archived_at", "owner_kind", "owner_id"}

// ── GetTag ───────────────────────────────────────────────────────────────────

func TestGetTag_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	tag := aTag()

	mock.ExpectQuery(`SELECT t\.id, t\.name, .+ FROM tags t LEFT JOIN tag_owners o ON o\.tag_id = t\.id WHERE t\.id = \$1 AND t\.deleted_at IS NULL AND t\.user_id = \$2`).
		WithArgs(tag.Id, *testScope.UserID()).
		WillReturnRows(pgxmock.NewRows(tagColsRead).
			AddRow(tag.Id, tag.Name, tag.Color, tag.UserId, nil, nil, nil))

	got, err := repo.GetTag(ctx, testScope, tag.Id)
	require.NoError(t, err)
	assert.Equal(t, tag, got)
}

func TestGetTag_Owned(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	tag := aTag()
	taskId := uuid.New()

	mock.ExpectQuery(`SELECT t\.id, t\.name, .+ FROM tags t LEFT JOIN tag_owners o ON o\.tag_id = t\.id WHERE t\.id = \$1`).
		WithArgs(tag.Id, *testScope.UserID()).
		WillReturnRows(pgxmock.NewRows(tagColsRead).
			AddRow(tag.Id, tag.Name, tag.Color, tag.UserId, nil, new("task"), &taskId))

	got, err := repo.GetTag(ctx, testScope, tag.Id)
	require.NoError(t, err)
	assert.Equal(t, &model.TagOwner{Kind: model.TagOwnerTask, Id: taskId}, got.Owner)
}

func TestGetTag_Archived(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	tag := aTag()
	archivedAt := time.Now().UTC()

	mock.ExpectQuery(`SELECT t\.id, t\.name, .+ FROM tags t LEFT JOIN tag_owners o ON o\.tag_id = t\.id WHERE t\.id = \$1 AND t\.deleted_at IS NULL AND t\.user_id = \$2`).
		WithArgs(tag.Id, *testScope.UserID()).
		WillReturnRows(pgxmock.NewRows(tagColsRead).
			AddRow(tag.Id, tag.Name, tag.Color, tag.UserId, &archivedAt, nil, nil))

	got, err := repo.GetTag(ctx, testScope, tag.Id)
	require.NoError(t, err)
	assert.True(t, got.Archived)
}

func TestGetTag_NotFound(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	id := uuid.New()

	mock.ExpectQuery(`SELECT t\.id, t\.name, .+ FROM tags t LEFT JOIN tag_owners o ON o\.tag_id = t\.id WHERE t\.id = \$1 AND t\.deleted_at IS NULL AND t\.user_id = \$2`).
		WithArgs(id, *testScope.UserID()).
		WillReturnRows(pgxmock.NewRows(tagColsRead))

	_, err := repo.GetTag(ctx, testScope, id)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestGetTag_NilId(t *testing.T) {
	repo, _ := newMock(t)
	_, err := repo.GetTag(context.Background(), testScope, uuid.Nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

// ── ListTags ─────────────────────────────────────────────────────────────────

func TestListTags_ReturnsSorted(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)

	t1, t2 := aTag(), aTag()
	t1.Name, t2.Name = "aaa", "zzz"

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM tags t LEFT JOIN tag_owners o ON o\.tag_id = t\.id WHERE t\.deleted_at IS NULL AND CASE WHEN o\.tag_id IS NULL THEN t\.archived_at ELSE o\.archived_at END IS NULL AND o\.tag_id IS NULL AND t\.user_id = \$1`).
		WithArgs(*testScope.UserID()).
		WillReturnRows(
			pgxmock.NewRows([]string{"count"}).
				AddRow(2),
		)

	mock.ExpectQuery(`SELECT t\.id, t\.name, .+ FROM tags t LEFT JOIN tag_owners o ON o\.tag_id = t\.id WHERE t\.deleted_at IS NULL AND CASE WHEN o\.tag_id IS NULL THEN t\.archived_at ELSE o\.archived_at END IS NULL AND o\.tag_id IS NULL AND t\.user_id = \$3 ORDER BY row_number\(\) OVER \(PARTITION BY o\.kind ORDER BY lower\(t\.name\) COLLATE "C", t\.name COLLATE "C", t\.id\), o\.kind NULLS FIRST LIMIT \$1 OFFSET \$2`).
		WithArgs(25, 0, *testScope.UserID()).
		WillReturnRows(
			pgxmock.NewRows(tagColsRead).
				AddRow(t1.Id, t1.Name, t1.Color, t1.UserId, nil, nil, nil).
				AddRow(t2.Id, t2.Name, t2.Color, t2.UserId, nil, nil, nil),
		)

	page, err := repo.ListTags(ctx, testScope, model.TagListParams{PaginationParams: model.DefaultPaginationParams()})
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

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM tags t LEFT JOIN tag_owners o ON o\.tag_id = t\.id WHERE t\.deleted_at IS NULL AND CASE WHEN o\.tag_id IS NULL THEN t\.archived_at ELSE o\.archived_at END IS NULL AND o\.tag_id IS NULL AND t\.user_id = \$1`).
		WithArgs(*testScope.UserID()).
		WillReturnRows(
			pgxmock.NewRows([]string{"count"}).
				AddRow(3),
		)

	mock.ExpectQuery(`SELECT t\.id, t\.name, .+ FROM tags t LEFT JOIN tag_owners o ON o\.tag_id = t\.id WHERE t\.deleted_at IS NULL AND CASE WHEN o\.tag_id IS NULL THEN t\.archived_at ELSE o\.archived_at END IS NULL AND o\.tag_id IS NULL AND t\.user_id = \$3 ORDER BY row_number\(\) OVER \(PARTITION BY o\.kind ORDER BY lower\(t\.name\) COLLATE "C", t\.name COLLATE "C", t\.id\), o\.kind NULLS FIRST LIMIT \$1 OFFSET \$2`).
		WithArgs(1, 1, *testScope.UserID()).
		WillReturnRows(
			pgxmock.NewRows(tagColsRead).
				AddRow(tag.Id, tag.Name, tag.Color, tag.UserId, nil, nil, nil),
		)

	page, err := repo.ListTags(ctx, testScope, model.TagListParams{PaginationParams: model.PaginationParams{
		Limit:  1,
		Offset: 1,
	}})

	require.NoError(t, err)

	assert.Len(t, page.Data, 1)
	assert.Equal(t, 3, page.TotalCount)
	assert.Equal(t, 1, page.Limit)
	assert.Equal(t, 1, page.Offset)
}

func TestListTags_Empty(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM tags t LEFT JOIN tag_owners o ON o\.tag_id = t\.id WHERE t\.deleted_at IS NULL AND CASE WHEN o\.tag_id IS NULL THEN t\.archived_at ELSE o\.archived_at END IS NULL AND o\.tag_id IS NULL AND t\.user_id = \$1`).
		WithArgs(*testScope.UserID()).
		WillReturnRows(
			pgxmock.NewRows([]string{"count"}).
				AddRow(0),
		)

	mock.ExpectQuery(`SELECT t\.id, t\.name, .+ FROM tags t LEFT JOIN tag_owners o ON o\.tag_id = t\.id WHERE t\.deleted_at IS NULL AND CASE WHEN o\.tag_id IS NULL THEN t\.archived_at ELSE o\.archived_at END IS NULL AND o\.tag_id IS NULL AND t\.user_id = \$3 ORDER BY row_number\(\) OVER \(PARTITION BY o\.kind ORDER BY lower\(t\.name\) COLLATE "C", t\.name COLLATE "C", t\.id\), o\.kind NULLS FIRST LIMIT \$1 OFFSET \$2`).
		WithArgs(25, 0, *testScope.UserID()).
		WillReturnRows(
			pgxmock.NewRows(tagColsRead),
		)

	page, err := repo.ListTags(ctx, testScope, model.TagListParams{PaginationParams: model.DefaultPaginationParams()})

	require.NoError(t, err)

	assert.Empty(t, page.Data)
	assert.Equal(t, 0, page.TotalCount)
}

func TestListTags_UnownedScope(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)

	tag := aTag()
	tag.UserId = nil

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM tags t LEFT JOIN tag_owners o ON o\.tag_id = t\.id WHERE t\.deleted_at IS NULL AND CASE WHEN o\.tag_id IS NULL THEN t\.archived_at ELSE o\.archived_at END IS NULL AND o\.tag_id IS NULL AND t\.user_id IS NULL`).
		WillReturnRows(
			pgxmock.NewRows([]string{"count"}).
				AddRow(1),
		)

	mock.ExpectQuery(`SELECT t\.id, t\.name, .+ FROM tags t LEFT JOIN tag_owners o ON o\.tag_id = t\.id WHERE t\.deleted_at IS NULL AND CASE WHEN o\.tag_id IS NULL THEN t\.archived_at ELSE o\.archived_at END IS NULL AND o\.tag_id IS NULL AND t\.user_id IS NULL ORDER BY row_number\(\) OVER \(PARTITION BY o\.kind ORDER BY lower\(t\.name\) COLLATE "C", t\.name COLLATE "C", t\.id\), o\.kind NULLS FIRST LIMIT \$1 OFFSET \$2`).
		WithArgs(25, 0).
		WillReturnRows(
			pgxmock.NewRows(tagColsRead).
				AddRow(tag.Id, tag.Name, tag.Color, tag.UserId, nil, nil, nil),
		)

	page, err := repo.ListTags(ctx, model.UnownedScope(), model.TagListParams{PaginationParams: model.DefaultPaginationParams()})
	require.NoError(t, err)

	assert.Len(t, page.Data, 1)
	assert.Equal(t, 1, page.TotalCount)
}

func TestListTags_IncludeArchived(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)

	tag := aTag()
	archivedAt := time.Now().UTC()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM tags t LEFT JOIN tag_owners o ON o\.tag_id = t\.id WHERE t\.deleted_at IS NULL AND o\.tag_id IS NULL AND t\.user_id = \$1`).
		WithArgs(*testScope.UserID()).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT t\.id, t\.name, .+ FROM tags t LEFT JOIN tag_owners o ON o\.tag_id = t\.id WHERE t\.deleted_at IS NULL AND o\.tag_id IS NULL AND t\.user_id = \$3 ORDER BY row_number\(\) OVER \(PARTITION BY o\.kind ORDER BY lower\(t\.name\) COLLATE "C", t\.name COLLATE "C", t\.id\), o\.kind NULLS FIRST LIMIT \$1 OFFSET \$2`).
		WithArgs(25, 0, *testScope.UserID()).
		WillReturnRows(pgxmock.NewRows(tagColsRead).
			AddRow(tag.Id, tag.Name, tag.Color, tag.UserId, &archivedAt, nil, nil))

	params := model.DefaultPaginationParams()
	params.IncludeArchived = true
	page, err := repo.ListTags(ctx, testScope, model.TagListParams{PaginationParams: params})
	require.NoError(t, err)

	assert.Len(t, page.Data, 1)
	assert.True(t, page.Data[0].Archived)
}

// ── CreateTag ────────────────────────────────────────────────────────────────

func TestCreateTag_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	tag := aTag()

	mock.ExpectQuery(`INSERT INTO tags \(id, name, color, user_id\) VALUES \(\$1, \$2, \$3, \$4\) RETURNING id, name, color, user_id`).
		WithArgs(tag.Id, tag.Name, tag.Color, testScope.UserID()).
		WillReturnRows(pgxmock.NewRows(tagCols).
			AddRow(tag.Id, tag.Name, tag.Color, tag.UserId))

	got, err := repo.CreateTag(ctx, testScope, tag)
	require.NoError(t, err)
	assert.Equal(t, tag, got)
}

func TestCreateTag_GeneratesIdWhenNil(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)

	tag := aTag()
	tag.Id = uuid.Nil

	mock.ExpectQuery(`INSERT INTO tags \(id, name, color, user_id\) VALUES \(\$1, \$2, \$3, \$4\) RETURNING id, name, color, user_id`).
		WithArgs(pgxmock.AnyArg(), tag.Name, tag.Color, testScope.UserID()).
		WillReturnRows(pgxmock.NewRows(tagCols).
			AddRow(uuid.New(), tag.Name, tag.Color, tag.UserId))

	got, err := repo.CreateTag(ctx, testScope, tag)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, got.Id)
}

func TestCreateTag_EmptyName(t *testing.T) {
	repo, _ := newMock(t)
	tag := aTag()
	tag.Name = ""
	_, err := repo.CreateTag(context.Background(), testScope, tag)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

// ── UpdateTag ────────────────────────────────────────────────────────────────

func TestUpdateTag_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	tag := aTag()
	tag.Name = "updated-name"

	mock.ExpectQuery(`UPDATE tags .+ WHERE .+ deleted_at IS NULL AND user_id = \$5 RETURNING id, name, color, user_id, archived_at`).
		WithArgs(tag.Id, tag.Name, tag.Color, tag.Archived, *testScope.UserID()).
		WillReturnRows(pgxmock.NewRows(tagColsArchived).
			AddRow(tag.Id, tag.Name, tag.Color, tag.UserId, nil))

	got, err := repo.UpdateTag(ctx, testScope, tag)
	require.NoError(t, err)
	assert.Equal(t, tag, got)
}

func TestUpdateTag_Archive(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	tag := aTag()
	tag.Archived = true
	archivedAt := time.Now().UTC()

	mock.ExpectQuery(`UPDATE tags .+ WHERE .+ deleted_at IS NULL AND user_id = \$5 RETURNING id, name, color, user_id, archived_at`).
		WithArgs(tag.Id, tag.Name, tag.Color, tag.Archived, *testScope.UserID()).
		WillReturnRows(pgxmock.NewRows(tagColsArchived).
			AddRow(tag.Id, tag.Name, tag.Color, tag.UserId, &archivedAt))

	got, err := repo.UpdateTag(ctx, testScope, tag)
	require.NoError(t, err)
	assert.True(t, got.Archived)
}

func TestUpdateTag_NotFound(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	tag := aTag()

	mock.ExpectQuery(`UPDATE tags .+ WHERE .+ deleted_at IS NULL AND user_id = \$5 RETURNING id, name, color, user_id, archived_at`).
		WithArgs(tag.Id, tag.Name, tag.Color, tag.Archived, *testScope.UserID()).
		WillReturnRows(pgxmock.NewRows(tagColsArchived))
	expectTagWriteMiss(mock, tag.Id, false)

	_, err := repo.UpdateTag(ctx, testScope, tag)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestUpdateTag_TaskTag(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	tag := aTag()

	mock.ExpectQuery(`UPDATE tags .+ WHERE id = \$1 AND NOT EXISTS \(SELECT 1 FROM tag_owners o WHERE o\.tag_id = tags\.id\) AND deleted_at IS NULL AND user_id = \$5 RETURNING`).
		WithArgs(tag.Id, tag.Name, tag.Color, tag.Archived, *testScope.UserID()).
		WillReturnRows(pgxmock.NewRows(tagColsArchived))
	expectTagWriteMiss(mock, tag.Id, true)

	_, err := repo.UpdateTag(ctx, testScope, tag)
	require.ErrorIs(t, err, model.ErrInvalidArgument)
}

func TestUpdateTag_NilId(t *testing.T) {
	repo, _ := newMock(t)
	tag := aTag()
	tag.Id = uuid.Nil
	_, err := repo.UpdateTag(context.Background(), testScope, tag)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestUpdateTag_EmptyName(t *testing.T) {
	repo, _ := newMock(t)
	tag := aTag()
	tag.Name = ""
	_, err := repo.UpdateTag(context.Background(), testScope, tag)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

// ── DeleteTag ────────────────────────────────────────────────────────────────

func TestDeleteTag_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	id := uuid.New()

	mock.ExpectExec(`WITH locked AS \( SELECT id FROM tags WHERE .* deleted_at IS NULL AND user_id = \$2 FOR UPDATE \) UPDATE tags SET deleted_at = now\(\) WHERE id IN \(SELECT id FROM locked\)`).
		WithArgs(id, *testScope.UserID()).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	require.NoError(t, repo.DeleteTag(ctx, testScope, id))
}

func TestDeleteTag_NotFound(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	id := uuid.New()

	mock.ExpectExec(`WITH locked AS \( SELECT id FROM tags WHERE .* deleted_at IS NULL AND user_id = \$2 FOR UPDATE \) UPDATE tags SET deleted_at = now\(\) WHERE id IN \(SELECT id FROM locked\)`).
		WithArgs(id, *testScope.UserID()).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))
	expectTagWriteMiss(mock, id, false)

	err := repo.DeleteTag(ctx, testScope, id)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestDeleteTag_TaskTag(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	id := uuid.New()

	mock.ExpectExec(`WITH locked AS \( SELECT id FROM tags WHERE id = \$1 AND NOT EXISTS \(SELECT 1 FROM tag_owners o WHERE o\.tag_id = tags\.id\) AND deleted_at IS NULL AND user_id = \$2 FOR UPDATE \)`).
		WithArgs(id, *testScope.UserID()).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))
	expectTagWriteMiss(mock, id, true)

	require.ErrorIs(t, repo.DeleteTag(ctx, testScope, id), model.ErrInvalidArgument)
}

func TestDeleteTag_NilId(t *testing.T) {
	repo, _ := newMock(t)
	err := repo.DeleteTag(context.Background(), testScope, uuid.Nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

// expectTagWriteMiss expects the lookup UpdateTag and DeleteTag make after
// matching no row, which tells a task tag from a missing one.
func expectTagWriteMiss(mock pgxmock.PgxPoolIface, id uuid.UUID, taskTag bool) {
	mock.ExpectQuery(`SELECT EXISTS \( SELECT 1 FROM tags t JOIN tag_owners o ON o\.tag_id = t\.id WHERE t\.id = \$1 AND t\.deleted_at IS NULL AND t\.user_id = \$2 \)`).
		WithArgs(id, *testScope.UserID()).
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(taskTag))
}
