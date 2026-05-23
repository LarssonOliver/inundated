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

// projectCols is the ordered column list returned by project queries.
var projectCols = []string{"id", "name", "color", "time_budget"}

// expectProjectTagsQuery registers the expectation for the secondary tag-fetch
// query that all Get/List/Create/Update calls issue after the main query.
func expectProjectTagsQuery(mock pgxmock.PgxPoolIface, projectId uuid.UUID, tagIds []uuid.UUID) {
	rows := pgxmock.NewRows([]string{"tag_id"})
	for _, tid := range tagIds {
		rows.AddRow(tid)
	}
	mock.ExpectQuery(`SELECT tag_id FROM project_tags WHERE project_id = \$1`).
		WithArgs(projectId).
		WillReturnRows(rows)
}

// expectSetProjectTags registers the delete + insert expectations produced by
// setProjectTags for the given tag list.
func expectSetProjectTags(mock pgxmock.PgxPoolIface, projectId uuid.UUID, tagIds []uuid.UUID) {
	mock.ExpectExec(`DELETE FROM project_tags WHERE project_id = \$1`).
		WithArgs(projectId).
		WillReturnResult(pgxmock.NewResult("DELETE", int64(len(tagIds))))
	for _, tid := range tagIds {
		mock.ExpectExec(`INSERT INTO project_tags`).
			WithArgs(projectId, tid).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))
	}
}

// ── GetProject ───────────────────────────────────────────────────────────────

func TestGetProject_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	p := aProject()

	mock.ExpectQuery(`SELECT .* FROM projects WHERE id = \$1 AND deleted_at IS NULL`).
		WithArgs(p.Id).
		WillReturnRows(pgxmock.NewRows(projectCols).
			AddRow(p.Id, p.Name, p.Color, p.TimeBudget))
	expectProjectTagsQuery(mock, p.Id, p.TagIds)

	got, err := repo.GetProject(ctx, p.Id)
	require.NoError(t, err)
	assert.Equal(t, p.Id, got.Id)
	assert.Equal(t, p.Name, got.Name)
	assert.Equal(t, p.Color, got.Color)
	assert.Equal(t, p.TimeBudget, got.TimeBudget)
	assert.ElementsMatch(t, p.TagIds, got.TagIds)
}

func TestGetProject_NilTimeBudget(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	p := aProject()
	p.TimeBudget = nil

	mock.ExpectQuery(`SELECT .* FROM projects WHERE id = \$1 AND deleted_at IS NULL`).
		WithArgs(p.Id).
		WillReturnRows(pgxmock.NewRows(projectCols).
			AddRow(p.Id, p.Name, p.Color, nil))
	expectProjectTagsQuery(mock, p.Id, nil)

	got, err := repo.GetProject(ctx, p.Id)
	require.NoError(t, err)
	assert.Nil(t, got.TimeBudget)
}

func TestGetProject_NotFound(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	id := uuid.New()

	mock.ExpectQuery(`SELECT .* FROM projects WHERE id = \$1 AND deleted_at IS NULL`).
		WithArgs(id).
		WillReturnRows(pgxmock.NewRows(projectCols))

	_, err := repo.GetProject(ctx, id)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestGetProject_NilId(t *testing.T) {
	repo, _ := newMock(t)
	_, err := repo.GetProject(context.Background(), uuid.Nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

// ── ListProjects ─────────────────────────────────────────────────────────────

func TestListProjects_ReturnsAll(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	p1, p2 := aProject(), aProject()

	mock.ExpectQuery(`SELECT id, name, color, time_budget, COUNT\(\*\) OVER\(\) AS total_count FROM projects WHERE deleted_at IS NULL ORDER BY name LIMIT \$1 OFFSET \$2`).
		WithArgs(25, 0).
		WillReturnRows(pgxmock.NewRows(append(projectCols, "total_count")).
			AddRow(p1.Id, p1.Name, p1.Color, p1.TimeBudget, 2).
			AddRow(p2.Id, p2.Name, p2.Color, p2.TimeBudget, 2))
	expectProjectTagsQuery(mock, p1.Id, p1.TagIds)
	expectProjectTagsQuery(mock, p2.Id, p2.TagIds)

	page, err := repo.ListProjects(ctx, model.DefaultPaginationParams())
	require.NoError(t, err)
	assert.Len(t, page.Data, 2)
	assert.Equal(t, 2, page.TotalCount)
}

func TestListProjects_WithPaginationParams(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	p := aProject()

	mock.ExpectQuery(`SELECT id, name, color, time_budget, COUNT\(\*\) OVER\(\) AS total_count FROM projects WHERE deleted_at IS NULL ORDER BY name LIMIT \$1 OFFSET \$2`).
		WithArgs(1, 1).
		WillReturnRows(pgxmock.NewRows(append(projectCols, "total_count")).
			AddRow(p.Id, p.Name, p.Color, p.TimeBudget, 3))
	expectProjectTagsQuery(mock, p.Id, p.TagIds)

	page, err := repo.ListProjects(ctx, model.PaginationParams{Limit: 1, Offset: 1})
	require.NoError(t, err)
	assert.Len(t, page.Data, 1)
	assert.Equal(t, 3, page.TotalCount)
	assert.Equal(t, 1, page.Limit)
	assert.Equal(t, 1, page.Offset)
}

func TestListProjects_Empty(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)

	mock.ExpectQuery(`SELECT id, name, color, time_budget, COUNT\(\*\) OVER\(\) AS total_count FROM projects WHERE deleted_at IS NULL ORDER BY name LIMIT \$1 OFFSET \$2`).
		WithArgs(25, 0).
		WillReturnRows(pgxmock.NewRows(append(projectCols, "total_count")))

	page, err := repo.ListProjects(ctx, model.DefaultPaginationParams())
	require.NoError(t, err)
	assert.Empty(t, page.Data)
	assert.Equal(t, 0, page.TotalCount)
}

// ── CreateProject ────────────────────────────────────────────────────────────

func TestCreateProject_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	p := aProject()

	mock.ExpectQuery(`INSERT INTO projects`).
		WithArgs(p.Id, p.Name, p.Color, p.TimeBudget).
		WillReturnRows(pgxmock.NewRows(projectCols).
			AddRow(p.Id, p.Name, p.Color, p.TimeBudget))
	expectSetProjectTags(mock, p.Id, p.TagIds)

	got, err := repo.CreateProject(ctx, p)
	require.NoError(t, err)
	assert.Equal(t, p.Id, got.Id)
	assert.ElementsMatch(t, p.TagIds, got.TagIds)
}

func TestCreateProject_NoTags(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	p := aProject()
	p.TagIds = nil

	mock.ExpectQuery(`INSERT INTO projects`).
		WithArgs(p.Id, p.Name, p.Color, p.TimeBudget).
		WillReturnRows(pgxmock.NewRows(projectCols).
			AddRow(p.Id, p.Name, p.Color, p.TimeBudget))
	expectSetProjectTags(mock, p.Id, nil)

	got, err := repo.CreateProject(ctx, p)
	require.NoError(t, err)
	assert.Empty(t, got.TagIds)
}

func TestCreateProject_EmptyName(t *testing.T) {
	repo, _ := newMock(t)
	p := aProject()
	p.Name = ""
	_, err := repo.CreateProject(context.Background(), p)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

func TestCreateProject_GeneratesIdWhenNil(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	p := aProject()
	p.Id = uuid.Nil

	generatedId := uuid.New()
	mock.ExpectQuery(`INSERT INTO projects`).
		WithArgs(pgxmock.AnyArg(), p.Name, p.Color, p.TimeBudget).
		WillReturnRows(pgxmock.NewRows(projectCols).
			AddRow(generatedId, p.Name, p.Color, p.TimeBudget))
	expectSetProjectTags(mock, generatedId, p.TagIds)

	got, err := repo.CreateProject(ctx, p)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, got.Id)
}

// ── UpdateProject ────────────────────────────────────────────────────────────

func TestUpdateProject_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	p := aProject()
	p.Name = "Renamed"
	newBudget := 4 * time.Hour
	p.TimeBudget = &newBudget

	mock.ExpectQuery(`UPDATE projects .* WHERE id = \$1 AND deleted_at IS NULL`).
		WithArgs(p.Id, p.Name, p.Color, p.TimeBudget).
		WillReturnRows(pgxmock.NewRows(projectCols).
			AddRow(p.Id, p.Name, p.Color, p.TimeBudget))
	expectSetProjectTags(mock, p.Id, p.TagIds)

	got, err := repo.UpdateProject(ctx, p)
	require.NoError(t, err)
	assert.Equal(t, "Renamed", got.Name)
	assert.Equal(t, &newBudget, got.TimeBudget)
}

func TestUpdateProject_NotFound(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	p := aProject()

	mock.ExpectQuery(`UPDATE projects .* WHERE id = \$1 AND deleted_at IS NULL`).
		WithArgs(p.Id, p.Name, p.Color, p.TimeBudget).
		WillReturnRows(pgxmock.NewRows(projectCols))

	_, err := repo.UpdateProject(ctx, p)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestUpdateProject_NilId(t *testing.T) {
	repo, _ := newMock(t)
	p := aProject()
	p.Id = uuid.Nil
	_, err := repo.UpdateProject(context.Background(), p)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

func TestUpdateProject_EmptyName(t *testing.T) {
	repo, _ := newMock(t)
	p := aProject()
	p.Name = ""
	_, err := repo.UpdateProject(context.Background(), p)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

// ── DeleteProject ────────────────────────────────────────────────────────────

func TestDeleteProject_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	id := uuid.New()

	mock.ExpectExec(`UPDATE projects SET deleted_at = now\(\) WHERE id = \$1 AND deleted_at IS NULL`).
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	require.NoError(t, repo.DeleteProject(ctx, id))
}

func TestDeleteProject_NotFound(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	id := uuid.New()

	mock.ExpectExec(`UPDATE projects SET deleted_at = now\(\) WHERE id = \$1 AND deleted_at IS NULL`).
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	err := repo.DeleteProject(ctx, id)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestDeleteProject_NilId(t *testing.T) {
	repo, _ := newMock(t)
	err := repo.DeleteProject(context.Background(), uuid.Nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}
