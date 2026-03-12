package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/larssonoliver/inundated/internal/model"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var timespanCols = []string{"id", "name", "start_time", "end_time"}

// expectTimespanTagsQuery registers the secondary tag-fetch expectation.
func expectTimespanTagsQuery(mock pgxmock.PgxPoolIface, timespanId uuid.UUID, tagIds []uuid.UUID) {
	rows := pgxmock.NewRows([]string{"tag_id"})
	for _, tid := range tagIds {
		rows.AddRow(tid)
	}
	mock.ExpectQuery(`SELECT tag_id FROM timespan_tags WHERE timespan_id = \$1`).
		WithArgs(timespanId).
		WillReturnRows(rows)
}

// expectSetTimespanTags registers the delete + insert expectations.
func expectSetTimespanTags(mock pgxmock.PgxPoolIface, timespanId uuid.UUID, tagIds []uuid.UUID) {
	mock.ExpectExec(`DELETE FROM timespan_tags WHERE timespan_id = \$1`).
		WithArgs(timespanId).
		WillReturnResult(pgxmock.NewResult("DELETE", int64(len(tagIds))))
	for _, tid := range tagIds {
		mock.ExpectExec(`INSERT INTO timespan_tags`).
			WithArgs(timespanId, tid).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))
	}
}

// ── GetTimespan ──────────────────────────────────────────────────────────────

func TestGetTimespan_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	ts := aTimespan()

	mock.ExpectQuery(`SELECT id, name, start_time, end_time FROM timespans WHERE id = \$1`).
		WithArgs(ts.Id).
		WillReturnRows(pgxmock.NewRows(timespanCols).
			AddRow(ts.Id, ts.Name, ts.StartTime, ts.EndTime))
	expectTimespanTagsQuery(mock, ts.Id, ts.TagIds)

	got, err := repo.GetTimespan(ctx, ts.Id)
	require.NoError(t, err)
	assert.Equal(t, ts.Id, got.Id)
	assert.Equal(t, ts.Name, got.Name)
	assert.True(t, ts.StartTime.Equal(got.StartTime))
	assert.True(t, ts.EndTime.Equal(got.EndTime))
	assert.ElementsMatch(t, ts.TagIds, got.TagIds)
}

func TestGetTimespan_NotFound(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	id := uuid.New()

	mock.ExpectQuery(`SELECT id, name, start_time, end_time FROM timespans WHERE id = \$1`).
		WithArgs(id).
		WillReturnRows(pgxmock.NewRows(timespanCols))

	_, err := repo.GetTimespan(ctx, id)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestGetTimespan_NilId(t *testing.T) {
	repo, _ := newMock(t)
	_, err := repo.GetTimespan(context.Background(), uuid.Nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

// ── ListTimespans ────────────────────────────────────────────────────────────

func TestListTimespans_ReturnsAll(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	ts1, ts2 := aTimespan(), aTimespan()

	mock.ExpectQuery(`SELECT id, name, start_time, end_time FROM timespans ORDER BY start_time DESC`).
		WillReturnRows(pgxmock.NewRows(timespanCols).
			AddRow(ts1.Id, ts1.Name, ts1.StartTime, ts1.EndTime).
			AddRow(ts2.Id, ts2.Name, ts2.StartTime, ts2.EndTime))
	expectTimespanTagsQuery(mock, ts1.Id, ts1.TagIds)
	expectTimespanTagsQuery(mock, ts2.Id, ts2.TagIds)

	got, err := repo.ListTimespans(ctx)
	require.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestListTimespans_Empty(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)

	mock.ExpectQuery(`SELECT id, name, start_time, end_time FROM timespans ORDER BY start_time DESC`).
		WillReturnRows(pgxmock.NewRows(timespanCols))

	got, err := repo.ListTimespans(ctx)
	require.NoError(t, err)
	assert.Empty(t, got)
}

// ── CreateTimespan ───────────────────────────────────────────────────────────

func TestCreateTimespan_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	ts := aTimespan()

	mock.ExpectQuery(`INSERT INTO timespans`).
		WithArgs(ts.Id, ts.Name, ts.StartTime, ts.EndTime).
		WillReturnRows(pgxmock.NewRows(timespanCols).
			AddRow(ts.Id, ts.Name, ts.StartTime, ts.EndTime))
	expectSetTimespanTags(mock, ts.Id, ts.TagIds)

	got, err := repo.CreateTimespan(ctx, ts)
	require.NoError(t, err)
	assert.Equal(t, ts.Id, got.Id)
	assert.ElementsMatch(t, ts.TagIds, got.TagIds)
}

func TestCreateTimespan_GeneratesIdWhenNil(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	ts := aTimespan()
	ts.Id = uuid.Nil

	generatedId := uuid.New()
	mock.ExpectQuery(`INSERT INTO timespans`).
		WithArgs(pgxmock.AnyArg(), ts.Name, ts.StartTime, ts.EndTime).
		WillReturnRows(pgxmock.NewRows(timespanCols).
			AddRow(generatedId, ts.Name, ts.StartTime, ts.EndTime))
	expectSetTimespanTags(mock, generatedId, ts.TagIds)

	got, err := repo.CreateTimespan(ctx, ts)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, got.Id)
}

func TestCreateTimespan_ZeroStartTime(t *testing.T) {
	repo, _ := newMock(t)
	ts := aTimespan()
	ts.StartTime = time.Time{}
	_, err := repo.CreateTimespan(context.Background(), ts)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

func TestCreateTimespan_EndTimeBeforeStartTime(t *testing.T) {
	repo, _ := newMock(t)
	ts := aTimespan()
	ts.EndTime = ts.StartTime.Add(-1 * time.Minute) // end before start
	_, err := repo.CreateTimespan(context.Background(), ts)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

func TestCreateTimespan_EndTimeEqualToStartTime(t *testing.T) {
	repo, _ := newMock(t)
	ts := aTimespan()
	ts.EndTime = ts.StartTime // equal is also invalid
	_, err := repo.CreateTimespan(context.Background(), ts)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

func TestCreateTimespan_ZeroEndTimeAllowed(t *testing.T) {
	// A zero EndTime represents an open/in-progress span — should be accepted.
	ctx := context.Background()
	repo, mock := newMock(t)
	ts := aTimespan()
	ts.EndTime = time.Time{}

	mock.ExpectQuery(`INSERT INTO timespans`).
		WithArgs(ts.Id, ts.Name, ts.StartTime, ts.EndTime).
		WillReturnRows(pgxmock.NewRows(timespanCols).
			AddRow(ts.Id, ts.Name, ts.StartTime, ts.EndTime))
	expectSetTimespanTags(mock, ts.Id, ts.TagIds)

	_, err := repo.CreateTimespan(ctx, ts)
	require.NoError(t, err)
}

// ── UpdateTimespan ───────────────────────────────────────────────────────────

func TestUpdateTimespan_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	ts := aTimespan()
	ts.Name = "renamed session"

	mock.ExpectQuery(`UPDATE timespans`).
		WithArgs(ts.Id, ts.Name, ts.StartTime, ts.EndTime).
		WillReturnRows(pgxmock.NewRows(timespanCols).
			AddRow(ts.Id, ts.Name, ts.StartTime, ts.EndTime))
	expectSetTimespanTags(mock, ts.Id, ts.TagIds)

	got, err := repo.UpdateTimespan(ctx, ts)
	require.NoError(t, err)
	assert.Equal(t, "renamed session", got.Name)
	assert.ElementsMatch(t, ts.TagIds, got.TagIds)
}

func TestUpdateTimespan_NotFound(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	ts := aTimespan()

	mock.ExpectQuery(`UPDATE timespans`).
		WithArgs(ts.Id, ts.Name, ts.StartTime, ts.EndTime).
		WillReturnError(pgx.ErrNoRows)

	_, err := repo.UpdateTimespan(ctx, ts)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestUpdateTimespan_NilId(t *testing.T) {
	repo, _ := newMock(t)
	ts := aTimespan()
	ts.Id = uuid.Nil
	_, err := repo.UpdateTimespan(context.Background(), ts)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

func TestUpdateTimespan_ZeroStartTime(t *testing.T) {
	repo, _ := newMock(t)
	ts := aTimespan()
	ts.StartTime = time.Time{}
	_, err := repo.UpdateTimespan(context.Background(), ts)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

func TestUpdateTimespan_EndTimeBeforeStartTime(t *testing.T) {
	repo, _ := newMock(t)
	ts := aTimespan()
	ts.EndTime = ts.StartTime.Add(-time.Second)
	_, err := repo.UpdateTimespan(context.Background(), ts)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

// ── DeleteTimespan ───────────────────────────────────────────────────────────

func TestDeleteTimespan_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	id := uuid.New()

	mock.ExpectExec(`DELETE FROM timespans WHERE id = \$1`).
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	require.NoError(t, repo.DeleteTimespan(ctx, id))
}

func TestDeleteTimespan_NotFound(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	id := uuid.New()

	mock.ExpectExec(`DELETE FROM timespans WHERE id = \$1`).
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	err := repo.DeleteTimespan(ctx, id)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestDeleteTimespan_NilId(t *testing.T) {
	repo, _ := newMock(t)
	err := repo.DeleteTimespan(context.Background(), uuid.Nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

func TestGetTotalDurationByTags_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)

	ids := []uuid.UUID{uuid.New(), uuid.New()}
	mock.ExpectQuery("SELECT SUM").
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"total_time"}).AddRow(dur(2 * time.Hour)))

	result, err := repo.GetTotalDurationByTags(ctx, ids)
	require.NoError(t, err)
	require.Equal(t, 2*time.Hour, result)
}

func TestGetTotalDurationByTags_InvalidTag(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)

	id := []uuid.UUID{uuid.New()}

	mock.ExpectQuery("SELECT SUM").
		WithArgs(id).
		WillReturnRows(pgxmock.NewRows([]string{"total_time"}).AddRow(nil))

	_, err := repo.GetTotalDurationByTags(ctx, id)
	require.NoError(t, err)
}

func TestGetTotalDurationByTags_EmptyList(t *testing.T) {
	ctx := context.Background()
	repo, _ := newMock(t)

	result, err := repo.GetTotalDurationByTags(ctx, []uuid.UUID{})
	require.NoError(t, err)
	require.Equal(t, 0*time.Hour, result)
}

func TestGetTotalDurationByTags_NilList(t *testing.T) {
	ctx := context.Background()
	repo, _ := newMock(t)

	out, err := repo.GetTotalDurationByTags(ctx, nil)
	require.NoError(t, err)
	require.Equal(t, 0*time.Hour, out)
}
