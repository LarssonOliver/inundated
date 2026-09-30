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

// projectCols is the ordered column list returned by CreateProject, which
// does not report archived_at (a newly created project is never archived).
var projectCols = []string{"id", "name", "color", "time_budget", "user_id", "tag_id"}

// projectColsArchived is the ordered column list returned by Get/List/Update,
// which additionally report the archived_at column.
var projectColsArchived = []string{"id", "name", "color", "time_budget", "user_id", "archived_at", "tag_id"}

// expectProjectTagsQuery registers the expectation for the secondary tag-fetch
// query that all Get/List/Create/Update calls issue after the main query.
func expectProjectTagsQuery(mock pgxmock.PgxPoolIface, projectId uuid.UUID, tagIds []uuid.UUID) {
	rows := pgxmock.NewRows([]string{"project_id", "tag_id"})
	for _, tid := range tagIds {
		rows.AddRow(projectId, tid)
	}
	mock.ExpectQuery(`SELECT l\.project_id, l\.tag_id FROM project_tags l JOIN tags t ON t\.id = l\.tag_id AND t\.deleted_at IS NULL WHERE l\.project_id = ANY\(\$1\)`).
		WithArgs([]uuid.UUID{projectId}).
		WillReturnRows(rows)
}

// expectTagsInScope registers the tag-ownership check that Create/Update issue
// as the first statement inside their transaction. All returned tags are
// reported as not archived.
func expectTagsInScope(mock pgxmock.PgxPoolIface, holder model.TagHolder, tagIds []uuid.UUID) {
	if len(tagIds) == 0 {
		return
	}
	rows := pgxmock.NewRows([]string{"id", "archived_at"})
	for _, tid := range tagIds {
		rows.AddRow(tid, nil)
	}
	mock.ExpectQuery(`SELECT t\.id, CASE WHEN o\.tag_id IS NULL THEN t\.archived_at ELSE o\.archived_at END FROM tags t LEFT JOIN tag_owners o ON o\.tag_id = t\.id WHERE t\.id = ANY\(\$1\) AND t\.deleted_at IS NULL AND \(o\.tag_id IS NULL OR o\.kind = ANY\(\$2\)\) AND t\.user_id = \$3`).
		WithArgs(tagIds, carriedKinds(holder), *testScope.UserID()).
		WillReturnRows(rows)
}

// carriedKinds is the owner kinds tagsInScope lets holder carry, as bound.
func carriedKinds(holder model.TagHolder) []string {
	kinds := []string{}
	for _, kind := range holder.CarriedOwnerKinds() {
		kinds = append(kinds, string(kind))
	}
	return kinds
}

// expectSetProjectTags registers the delete + insert expectations produced by
// setLinkedTags for the given tag list.
func expectSetProjectTags(mock pgxmock.PgxPoolIface, projectId uuid.UUID, tagIds []uuid.UUID) {
	mock.ExpectExec(`DELETE FROM project_tags WHERE project_id = \$1`).
		WithArgs(projectId).
		WillReturnResult(pgxmock.NewResult("DELETE", int64(len(tagIds))))
	if len(tagIds) == 0 {
		return
	}
	mock.ExpectExec(`INSERT INTO project_tags \(project_id, tag_id\) SELECT \$1, unnest\(\$2::uuid\[\]\)`).
		WithArgs(projectId, tagIds).
		WillReturnResult(pgxmock.NewResult("INSERT", int64(len(tagIds))))
}

// expectCreateProjectTag registers the insert of p's project tag that
// CreateProject issues before the project itself.
func expectCreateProjectTag(mock pgxmock.PgxPoolIface, p model.Project) {
	mock.ExpectExec(`INSERT INTO tags \(id, name, color, user_id\) VALUES \(\$1, \$2, \$3, \$4\)`).
		WithArgs(pgxmock.AnyArg(), p.Name, p.Color, testScope.UserID()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
}

// expectRenameProjectTag registers UpdateProject's write of p's name and
// color to its project tag.
func expectRenameProjectTag(mock pgxmock.PgxPoolIface, p model.Project) {
	mock.ExpectExec(`UPDATE tags SET name = \$2, color = \$3 WHERE id = \$1 AND \(name, color\) IS DISTINCT FROM \(\$2, \$3\)`).
		WithArgs(p.TagId, p.Name, p.Color).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
}

// ── GetProject ───────────────────────────────────────────────────────────────

func TestGetProject_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	p := aProject()

	mock.ExpectQuery(`SELECT .* FROM projects WHERE id = \$1 AND deleted_at IS NULL AND user_id = \$2`).
		WithArgs(p.Id, *testScope.UserID()).
		WillReturnRows(pgxmock.NewRows(projectColsArchived).
			AddRow(p.Id, p.Name, p.Color, p.TimeBudget, testScope.UserID(), nil, p.TagId))
	expectProjectTagsQuery(mock, p.Id, p.TagIds)

	got, err := repo.GetProject(ctx, testScope, p.Id)
	require.NoError(t, err)
	assert.Equal(t, p.Id, got.Id)
	assert.Equal(t, p.Name, got.Name)
	assert.Equal(t, p.Color, got.Color)
	assert.Equal(t, p.TimeBudget, got.TimeBudget)
	assert.False(t, got.Archived)
	assert.ElementsMatch(t, p.TagIds, got.TagIds)
}

func TestGetProject_NilTimeBudget(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	p := aProject()
	p.TimeBudget = nil

	mock.ExpectQuery(`SELECT .* FROM projects WHERE id = \$1 AND deleted_at IS NULL AND user_id = \$2`).
		WithArgs(p.Id, *testScope.UserID()).
		WillReturnRows(pgxmock.NewRows(projectColsArchived).
			AddRow(p.Id, p.Name, p.Color, nil, testScope.UserID(), nil, p.TagId))
	expectProjectTagsQuery(mock, p.Id, nil)

	got, err := repo.GetProject(ctx, testScope, p.Id)
	require.NoError(t, err)
	assert.Nil(t, got.TimeBudget)
}

func TestGetProject_Archived(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	p := aProject()
	archivedAt := time.Now().UTC()
	archivedAtPtr := &archivedAt

	mock.ExpectQuery(`SELECT .* FROM projects WHERE id = \$1 AND deleted_at IS NULL AND user_id = \$2`).
		WithArgs(p.Id, *testScope.UserID()).
		WillReturnRows(pgxmock.NewRows(projectColsArchived).
			AddRow(p.Id, p.Name, p.Color, p.TimeBudget, testScope.UserID(), archivedAtPtr, p.TagId))
	expectProjectTagsQuery(mock, p.Id, p.TagIds)

	got, err := repo.GetProject(ctx, testScope, p.Id)
	require.NoError(t, err)
	assert.True(t, got.Archived)
}

func TestGetProject_NotFound(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	id := uuid.New()

	mock.ExpectQuery(`SELECT .* FROM projects WHERE id = \$1 AND deleted_at IS NULL AND user_id = \$2`).
		WithArgs(id, *testScope.UserID()).
		WillReturnRows(pgxmock.NewRows(projectColsArchived))

	_, err := repo.GetProject(ctx, testScope, id)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestGetProject_NilId(t *testing.T) {
	repo, _ := newMock(t)
	_, err := repo.GetProject(context.Background(), testScope, uuid.Nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

// ── ListProjects ─────────────────────────────────────────────────────────────

func TestListProjects_ReturnsAll(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)

	p1, p2 := aProject(), aProject()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM projects WHERE deleted_at IS NULL AND archived_at IS NULL AND user_id = \$1`).
		WithArgs(*testScope.UserID()).
		WillReturnRows(
			pgxmock.NewRows([]string{"count"}).
				AddRow(2),
		)

	mock.ExpectQuery(`SELECT id, name, color, time_budget, user_id, archived_at, tag_id FROM projects WHERE deleted_at IS NULL AND archived_at IS NULL AND user_id = \$3 ORDER BY name LIMIT \$1 OFFSET \$2`).
		WithArgs(25, 0, *testScope.UserID()).
		WillReturnRows(
			pgxmock.NewRows(projectColsArchived).
				AddRow(p1.Id, p1.Name, p1.Color, p1.TimeBudget, testScope.UserID(), nil, p1.TagId).
				AddRow(p2.Id, p2.Name, p2.Color, p2.TimeBudget, testScope.UserID(), nil, p2.TagId),
		)

	expectProjectTagsQuery(mock, p1.Id, p1.TagIds)
	expectProjectTagsQuery(mock, p2.Id, p2.TagIds)

	page, err := repo.ListProjects(ctx, testScope, model.DefaultPaginationParams())

	require.NoError(t, err)

	assert.Len(t, page.Data, 2)
	assert.Equal(t, 2, page.TotalCount)
}

func TestListProjects_WithPaginationParams(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)

	p := aProject()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM projects WHERE deleted_at IS NULL AND archived_at IS NULL AND user_id = \$1`).
		WithArgs(*testScope.UserID()).
		WillReturnRows(
			pgxmock.NewRows([]string{"count"}).
				AddRow(3),
		)

	mock.ExpectQuery(`SELECT id, name, color, time_budget, user_id, archived_at, tag_id FROM projects WHERE deleted_at IS NULL AND archived_at IS NULL AND user_id = \$3 ORDER BY name LIMIT \$1 OFFSET \$2`).
		WithArgs(1, 1, *testScope.UserID()).
		WillReturnRows(
			pgxmock.NewRows(projectColsArchived).
				AddRow(p.Id, p.Name, p.Color, p.TimeBudget, testScope.UserID(), nil, p.TagId),
		)

	expectProjectTagsQuery(mock, p.Id, p.TagIds)

	page, err := repo.ListProjects(ctx, testScope, model.PaginationParams{
		Limit:  1,
		Offset: 1,
	})

	require.NoError(t, err)

	assert.Len(t, page.Data, 1)
	assert.Equal(t, 3, page.TotalCount)
	assert.Equal(t, 1, page.Limit)
	assert.Equal(t, 1, page.Offset)
}

func TestListProjects_Empty(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM projects WHERE deleted_at IS NULL AND archived_at IS NULL AND user_id = \$1`).
		WithArgs(*testScope.UserID()).
		WillReturnRows(
			pgxmock.NewRows([]string{"count"}).
				AddRow(0),
		)

	mock.ExpectQuery(`SELECT id, name, color, time_budget, user_id, archived_at, tag_id FROM projects WHERE deleted_at IS NULL AND archived_at IS NULL AND user_id = \$3 ORDER BY name LIMIT \$1 OFFSET \$2`).
		WithArgs(25, 0, *testScope.UserID()).
		WillReturnRows(
			pgxmock.NewRows(projectColsArchived),
		)

	page, err := repo.ListProjects(ctx, testScope, model.DefaultPaginationParams())

	require.NoError(t, err)

	assert.Empty(t, page.Data)
	assert.Equal(t, 0, page.TotalCount)
}

func TestListProjects_UnownedScope(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)

	p := aProject()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM projects WHERE deleted_at IS NULL AND archived_at IS NULL AND user_id IS NULL`).
		WillReturnRows(
			pgxmock.NewRows([]string{"count"}).
				AddRow(1),
		)

	mock.ExpectQuery(`SELECT id, name, color, time_budget, user_id, archived_at, tag_id FROM projects WHERE deleted_at IS NULL AND archived_at IS NULL AND user_id IS NULL ORDER BY name LIMIT \$1 OFFSET \$2`).
		WithArgs(25, 0).
		WillReturnRows(
			pgxmock.NewRows(projectColsArchived).
				AddRow(p.Id, p.Name, p.Color, p.TimeBudget, nil, nil, p.TagId),
		)

	expectProjectTagsQuery(mock, p.Id, p.TagIds)

	page, err := repo.ListProjects(ctx, model.UnownedScope(), model.DefaultPaginationParams())
	require.NoError(t, err)

	assert.Len(t, page.Data, 1)
	assert.Equal(t, 1, page.TotalCount)
}

func TestListProjects_IncludeArchived(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)

	p := aProject()
	archivedAt := time.Now().UTC()
	archivedAtPtr := &archivedAt

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM projects WHERE deleted_at IS NULL AND user_id = \$1`).
		WithArgs(*testScope.UserID()).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT id, name, color, time_budget, user_id, archived_at, tag_id FROM projects WHERE deleted_at IS NULL AND user_id = \$3 ORDER BY name LIMIT \$1 OFFSET \$2`).
		WithArgs(25, 0, *testScope.UserID()).
		WillReturnRows(pgxmock.NewRows(projectColsArchived).
			AddRow(p.Id, p.Name, p.Color, p.TimeBudget, testScope.UserID(), archivedAtPtr, p.TagId))

	expectProjectTagsQuery(mock, p.Id, p.TagIds)

	params := model.DefaultPaginationParams()
	params.IncludeArchived = true
	page, err := repo.ListProjects(ctx, testScope, params)
	require.NoError(t, err)

	assert.Len(t, page.Data, 1)
	assert.True(t, page.Data[0].Archived)
}

// ── CreateProject ────────────────────────────────────────────────────────────

func TestCreateProject_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	p := aProject()

	mock.ExpectBegin()
	expectTagsInScope(mock, model.TagHolderProject, p.TagIds)
	expectCreateProjectTag(mock, p)
	mock.ExpectQuery(`INSERT INTO projects \(id, name, color, time_budget, user_id, tag_id\) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6\) RETURNING id, name, color, time_budget, user_id, tag_id`).
		WithArgs(p.Id, p.Name, p.Color, p.TimeBudget, testScope.UserID(), pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows(projectCols).
			AddRow(p.Id, p.Name, p.Color, p.TimeBudget, testScope.UserID(), uuid.New()))
	expectSetProjectTags(mock, p.Id, p.TagIds)
	mock.ExpectCommit()

	got, err := repo.CreateProject(ctx, testScope, p)
	require.NoError(t, err)
	assert.Equal(t, p.Id, got.Id)
	assert.ElementsMatch(t, p.TagIds, got.TagIds)
}

func TestCreateProject_NoTags(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	p := aProject()
	p.TagIds = nil

	mock.ExpectBegin()
	expectTagsInScope(mock, model.TagHolderProject, p.TagIds)
	expectCreateProjectTag(mock, p)
	mock.ExpectQuery(`INSERT INTO projects \(id, name, color, time_budget, user_id, tag_id\) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6\) RETURNING id, name, color, time_budget, user_id, tag_id`).
		WithArgs(p.Id, p.Name, p.Color, p.TimeBudget, testScope.UserID(), pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows(projectCols).
			AddRow(p.Id, p.Name, p.Color, p.TimeBudget, testScope.UserID(), uuid.New()))
	expectSetProjectTags(mock, p.Id, nil)
	mock.ExpectCommit()

	got, err := repo.CreateProject(ctx, testScope, p)
	require.NoError(t, err)
	assert.Empty(t, got.TagIds)
}

func TestCreateProject_EmptyName(t *testing.T) {
	repo, _ := newMock(t)
	p := aProject()
	p.Name = ""
	_, err := repo.CreateProject(context.Background(), testScope, p)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

func TestCreateProject_GeneratesIdWhenNil(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	p := aProject()
	p.Id = uuid.Nil

	generatedId := uuid.New()
	mock.ExpectBegin()
	expectTagsInScope(mock, model.TagHolderProject, p.TagIds)
	expectCreateProjectTag(mock, p)
	mock.ExpectQuery(`INSERT INTO projects \(id, name, color, time_budget, user_id, tag_id\) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6\) RETURNING id, name, color, time_budget, user_id, tag_id`).
		WithArgs(pgxmock.AnyArg(), p.Name, p.Color, p.TimeBudget, testScope.UserID(), pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows(projectCols).
			AddRow(generatedId, p.Name, p.Color, p.TimeBudget, testScope.UserID(), uuid.New()))
	expectSetProjectTags(mock, generatedId, p.TagIds)
	mock.ExpectCommit()

	got, err := repo.CreateProject(ctx, testScope, p)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, got.Id)
}

func TestCreateProject_ForeignTagRejected(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	p := aProject()

	mock.ExpectBegin()
	// tagsInScope finds fewer live, in-scope tags than requested; the parent
	// INSERT never runs and the transaction rolls back.
	rows := pgxmock.NewRows([]string{"id", "archived_at"})
	for _, tid := range p.TagIds[:len(p.TagIds)-1] {
		rows.AddRow(tid, nil)
	}
	mock.ExpectQuery(`SELECT t\.id, CASE WHEN o\.tag_id IS NULL THEN t\.archived_at ELSE o\.archived_at END FROM tags t LEFT JOIN tag_owners o ON o\.tag_id = t\.id WHERE t\.id = ANY\(\$1\) AND t\.deleted_at IS NULL AND \(o\.tag_id IS NULL OR o\.kind = ANY\(\$2\)\) AND t\.user_id = \$3`).
		WithArgs(p.TagIds, carriedKinds(model.TagHolderProject), *testScope.UserID()).
		WillReturnRows(rows)
	mock.ExpectRollback()

	_, err := repo.CreateProject(ctx, testScope, p)
	require.ErrorIs(t, err, model.ErrInvalidReference)
}

// ── UpdateProject ────────────────────────────────────────────────────────────

func TestUpdateProject_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	p := aProject()
	p.Name = "Renamed"
	newBudget := 4 * time.Hour
	p.TimeBudget = &newBudget

	mock.ExpectBegin()
	expectTagsInScope(mock, model.TagHolderProject, p.TagIds)
	mock.ExpectQuery(`UPDATE projects .* WHERE id = \$1 AND deleted_at IS NULL AND user_id = \$6 RETURNING id, name, color, time_budget, user_id, archived_at, tag_id`).
		WithArgs(p.Id, p.Name, p.Color, p.TimeBudget, p.Archived, *testScope.UserID()).
		WillReturnRows(pgxmock.NewRows(projectColsArchived).
			AddRow(p.Id, p.Name, p.Color, p.TimeBudget, testScope.UserID(), nil, p.TagId))
	expectRenameProjectTag(mock, p)
	expectSetProjectTags(mock, p.Id, p.TagIds)
	mock.ExpectCommit()

	got, err := repo.UpdateProject(ctx, testScope, p)
	require.NoError(t, err)
	assert.Equal(t, "Renamed", got.Name)
	assert.Equal(t, &newBudget, got.TimeBudget)
}

func TestUpdateProject_Archive(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	p := aProject()
	p.Archived = true
	archivedAtNow := time.Now().UTC()

	mock.ExpectBegin()
	expectTagsInScope(mock, model.TagHolderProject, p.TagIds)
	mock.ExpectQuery(`UPDATE projects .* WHERE id = \$1 AND deleted_at IS NULL AND user_id = \$6 RETURNING id, name, color, time_budget, user_id, archived_at, tag_id`).
		WithArgs(p.Id, p.Name, p.Color, p.TimeBudget, p.Archived, *testScope.UserID()).
		WillReturnRows(pgxmock.NewRows(projectColsArchived).
			AddRow(p.Id, p.Name, p.Color, p.TimeBudget, testScope.UserID(), &archivedAtNow, p.TagId))
	expectRenameProjectTag(mock, p)
	expectSetProjectTags(mock, p.Id, p.TagIds)
	mock.ExpectCommit()

	got, err := repo.UpdateProject(ctx, testScope, p)
	require.NoError(t, err)
	assert.True(t, got.Archived)
}

func TestUpdateProject_NotFound(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	p := aProject()

	mock.ExpectBegin()
	expectTagsInScope(mock, model.TagHolderProject, p.TagIds)
	mock.ExpectQuery(`UPDATE projects .* WHERE id = \$1 AND deleted_at IS NULL AND user_id = \$6 RETURNING id, name, color, time_budget, user_id, archived_at, tag_id`).
		WithArgs(p.Id, p.Name, p.Color, p.TimeBudget, p.Archived, *testScope.UserID()).
		WillReturnRows(pgxmock.NewRows(projectColsArchived))
	mock.ExpectRollback()

	_, err := repo.UpdateProject(ctx, testScope, p)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestUpdateProject_NilId(t *testing.T) {
	repo, _ := newMock(t)
	p := aProject()
	p.Id = uuid.Nil
	_, err := repo.UpdateProject(context.Background(), testScope, p)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestUpdateProject_EmptyName(t *testing.T) {
	repo, _ := newMock(t)
	p := aProject()
	p.Name = ""
	_, err := repo.UpdateProject(context.Background(), testScope, p)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

// ── DeleteProject ────────────────────────────────────────────────────────────

// expectDeleteProjectLocks expects DeleteProject to lock the project and its
// project tag, then check for time attributed to the tag, answering
// attributed.
func expectDeleteProjectLocks(mock pgxmock.PgxPoolIface, id, tagId uuid.UUID, attributed bool) {
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT tag_id FROM projects WHERE id = \$1 AND deleted_at IS NULL AND user_id = \$2 FOR UPDATE`).
		WithArgs(id, *testScope.UserID()).
		WillReturnRows(pgxmock.NewRows([]string{"tag_id"}).AddRow(tagId))
	mock.ExpectExec(`SELECT id FROM tags WHERE id = \$1 FOR UPDATE`).
		WithArgs(tagId).
		WillReturnResult(pgxmock.NewResult("SELECT", 1))
	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs([]uuid.UUID{tagId}, *testScope.UserID()).
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(attributed))
}

func TestDeleteProject_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	id := uuid.New()
	tagId := uuid.New()

	expectDeleteProjectLocks(mock, id, tagId, false)
	mock.ExpectExec(`UPDATE projects SET deleted_at = now\(\) WHERE id = \$1`).
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectExec(`UPDATE tags SET deleted_at = now\(\) WHERE id = \$1 AND deleted_at IS NULL`).
		WithArgs(tagId).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectCommit()

	require.NoError(t, repo.DeleteProject(ctx, testScope, id))
}

func TestDeleteProject_AttributedTime(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	id := uuid.New()

	expectDeleteProjectLocks(mock, id, uuid.New(), true)
	mock.ExpectRollback()

	require.ErrorIs(t, repo.DeleteProject(ctx, testScope, id), model.ErrConflict)
}

func TestDeleteProject_NotFound(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	id := uuid.New()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT tag_id FROM projects WHERE id = \$1 AND deleted_at IS NULL AND user_id = \$2 FOR UPDATE`).
		WithArgs(id, *testScope.UserID()).
		WillReturnRows(pgxmock.NewRows([]string{"tag_id"}))
	mock.ExpectRollback()

	err := repo.DeleteProject(ctx, testScope, id)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestDeleteProject_NilId(t *testing.T) {
	repo, _ := newMock(t)
	err := repo.DeleteProject(context.Background(), testScope, uuid.Nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}
