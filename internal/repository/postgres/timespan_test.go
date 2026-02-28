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

var timeSpanCols = []string{"id", "name", "start_time", "end_time"}

// expectTimeSpanTagsQuery registers the secondary tag-fetch expectation.
func expectTimeSpanTagsQuery(mock pgxmock.PgxPoolIface, timeSpanId uuid.UUID, tagIds []uuid.UUID) {
	rows := pgxmock.NewRows([]string{"tag_id"})
	for _, tid := range tagIds {
		rows.AddRow(tid)
	}
	mock.ExpectQuery(`SELECT tag_id FROM time_span_tags WHERE time_span_id = \$1`).
		WithArgs(timeSpanId).
		WillReturnRows(rows)
}

// expectSetTimeSpanTags registers the delete + insert expectations.
func expectSetTimeSpanTags(mock pgxmock.PgxPoolIface, timeSpanId uuid.UUID, tagIds []uuid.UUID) {
	mock.ExpectExec(`DELETE FROM time_span_tags WHERE time_span_id = \$1`).
		WithArgs(timeSpanId).
		WillReturnResult(pgxmock.NewResult("DELETE", int64(len(tagIds))))
	for _, tid := range tagIds {
		mock.ExpectExec(`INSERT INTO time_span_tags`).
			WithArgs(timeSpanId, tid).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))
	}
}

// ── GetTimeSpan ──────────────────────────────────────────────────────────────

func TestGetTimeSpan_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	ts := aTimeSpan()

	mock.ExpectQuery(`SELECT id, name, start_time, end_time FROM time_spans WHERE id = \$1`).
		WithArgs(ts.Id).
		WillReturnRows(pgxmock.NewRows(timeSpanCols).
			AddRow(ts.Id, ts.Name, ts.StartTime, ts.EndTime))
	expectTimeSpanTagsQuery(mock, ts.Id, ts.TagIds)

	got, err := repo.GetTimeSpan(ctx, ts.Id)
	require.NoError(t, err)
	assert.Equal(t, ts.Id, got.Id)
	assert.Equal(t, ts.Name, got.Name)
	assert.True(t, ts.StartTime.Equal(got.StartTime))
	assert.True(t, ts.EndTime.Equal(got.EndTime))
	assert.ElementsMatch(t, ts.TagIds, got.TagIds)
}

func TestGetTimeSpan_NotFound(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	id := uuid.New()

	mock.ExpectQuery(`SELECT id, name, start_time, end_time FROM time_spans WHERE id = \$1`).
		WithArgs(id).
		WillReturnRows(pgxmock.NewRows(timeSpanCols))

	_, err := repo.GetTimeSpan(ctx, id)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestGetTimeSpan_NilId(t *testing.T) {
	repo, _ := newMock(t)
	_, err := repo.GetTimeSpan(context.Background(), uuid.Nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

// ── ListTimeSpans ────────────────────────────────────────────────────────────

func TestListTimeSpans_ReturnsAll(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	ts1, ts2 := aTimeSpan(), aTimeSpan()

	mock.ExpectQuery(`SELECT id, name, start_time, end_time FROM time_spans ORDER BY start_time DESC`).
		WillReturnRows(pgxmock.NewRows(timeSpanCols).
			AddRow(ts1.Id, ts1.Name, ts1.StartTime, ts1.EndTime).
			AddRow(ts2.Id, ts2.Name, ts2.StartTime, ts2.EndTime))
	expectTimeSpanTagsQuery(mock, ts1.Id, ts1.TagIds)
	expectTimeSpanTagsQuery(mock, ts2.Id, ts2.TagIds)

	got, err := repo.ListTimeSpans(ctx)
	require.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestListTimeSpans_Empty(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)

	mock.ExpectQuery(`SELECT id, name, start_time, end_time FROM time_spans ORDER BY start_time DESC`).
		WillReturnRows(pgxmock.NewRows(timeSpanCols))

	got, err := repo.ListTimeSpans(ctx)
	require.NoError(t, err)
	assert.Empty(t, got)
}

// ── CreateTimeSpan ───────────────────────────────────────────────────────────

func TestCreateTimeSpan_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	ts := aTimeSpan()

	mock.ExpectQuery(`INSERT INTO time_spans`).
		WithArgs(ts.Id, ts.Name, ts.StartTime, ts.EndTime).
		WillReturnRows(pgxmock.NewRows(timeSpanCols).
			AddRow(ts.Id, ts.Name, ts.StartTime, ts.EndTime))
	expectSetTimeSpanTags(mock, ts.Id, ts.TagIds)

	got, err := repo.CreateTimeSpan(ctx, ts)
	require.NoError(t, err)
	assert.Equal(t, ts.Id, got.Id)
	assert.ElementsMatch(t, ts.TagIds, got.TagIds)
}

func TestCreateTimeSpan_GeneratesIdWhenNil(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	ts := aTimeSpan()
	ts.Id = uuid.Nil

	generatedId := uuid.New()
	mock.ExpectQuery(`INSERT INTO time_spans`).
		WithArgs(pgxmock.AnyArg(), ts.Name, ts.StartTime, ts.EndTime).
		WillReturnRows(pgxmock.NewRows(timeSpanCols).
			AddRow(generatedId, ts.Name, ts.StartTime, ts.EndTime))
	expectSetTimeSpanTags(mock, generatedId, ts.TagIds)

	got, err := repo.CreateTimeSpan(ctx, ts)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, got.Id)
}

func TestCreateTimeSpan_ZeroStartTime(t *testing.T) {
	repo, _ := newMock(t)
	ts := aTimeSpan()
	ts.StartTime = time.Time{}
	_, err := repo.CreateTimeSpan(context.Background(), ts)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

func TestCreateTimeSpan_EndTimeBeforeStartTime(t *testing.T) {
	repo, _ := newMock(t)
	ts := aTimeSpan()
	ts.EndTime = ts.StartTime.Add(-1 * time.Minute) // end before start
	_, err := repo.CreateTimeSpan(context.Background(), ts)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

func TestCreateTimeSpan_EndTimeEqualToStartTime(t *testing.T) {
	repo, _ := newMock(t)
	ts := aTimeSpan()
	ts.EndTime = ts.StartTime // equal is also invalid
	_, err := repo.CreateTimeSpan(context.Background(), ts)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

func TestCreateTimeSpan_ZeroEndTimeAllowed(t *testing.T) {
	// A zero EndTime represents an open/in-progress span — should be accepted.
	ctx := context.Background()
	repo, mock := newMock(t)
	ts := aTimeSpan()
	ts.EndTime = time.Time{}

	mock.ExpectQuery(`INSERT INTO time_spans`).
		WithArgs(ts.Id, ts.Name, ts.StartTime, ts.EndTime).
		WillReturnRows(pgxmock.NewRows(timeSpanCols).
			AddRow(ts.Id, ts.Name, ts.StartTime, ts.EndTime))
	expectSetTimeSpanTags(mock, ts.Id, ts.TagIds)

	_, err := repo.CreateTimeSpan(ctx, ts)
	require.NoError(t, err)
}

// ── UpdateTimeSpan ───────────────────────────────────────────────────────────

func TestUpdateTimeSpan_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	ts := aTimeSpan()
	ts.Name = "renamed session"

	mock.ExpectQuery(`UPDATE time_spans`).
		WithArgs(ts.Id, ts.Name, ts.StartTime, ts.EndTime).
		WillReturnRows(pgxmock.NewRows(timeSpanCols).
			AddRow(ts.Id, ts.Name, ts.StartTime, ts.EndTime))
	expectSetTimeSpanTags(mock, ts.Id, ts.TagIds)

	got, err := repo.UpdateTimeSpan(ctx, ts)
	require.NoError(t, err)
	assert.Equal(t, "renamed session", got.Name)
	assert.ElementsMatch(t, ts.TagIds, got.TagIds)
}

func TestUpdateTimeSpan_NotFound(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	ts := aTimeSpan()

	mock.ExpectQuery(`UPDATE time_spans`).
		WithArgs(ts.Id, ts.Name, ts.StartTime, ts.EndTime).
		WillReturnError(pgx.ErrNoRows)

	_, err := repo.UpdateTimeSpan(ctx, ts)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestUpdateTimeSpan_NilId(t *testing.T) {
	repo, _ := newMock(t)
	ts := aTimeSpan()
	ts.Id = uuid.Nil
	_, err := repo.UpdateTimeSpan(context.Background(), ts)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

func TestUpdateTimeSpan_ZeroStartTime(t *testing.T) {
	repo, _ := newMock(t)
	ts := aTimeSpan()
	ts.StartTime = time.Time{}
	_, err := repo.UpdateTimeSpan(context.Background(), ts)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

func TestUpdateTimeSpan_EndTimeBeforeStartTime(t *testing.T) {
	repo, _ := newMock(t)
	ts := aTimeSpan()
	ts.EndTime = ts.StartTime.Add(-time.Second)
	_, err := repo.UpdateTimeSpan(context.Background(), ts)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

// ── DeleteTimeSpan ───────────────────────────────────────────────────────────

func TestDeleteTimeSpan_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	id := uuid.New()

	mock.ExpectExec(`DELETE FROM time_spans WHERE id = \$1`).
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	require.NoError(t, repo.DeleteTimeSpan(ctx, id))
}

func TestDeleteTimeSpan_NotFound(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	id := uuid.New()

	mock.ExpectExec(`DELETE FROM time_spans WHERE id = \$1`).
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	err := repo.DeleteTimeSpan(ctx, id)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestDeleteTimeSpan_NilId(t *testing.T) {
	repo, _ := newMock(t)
	err := repo.DeleteTimeSpan(context.Background(), uuid.Nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}
