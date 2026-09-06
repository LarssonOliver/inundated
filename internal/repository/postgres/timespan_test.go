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

	mock.ExpectQuery(`SELECT id, name, start_time, end_time FROM timespans WHERE id = \$1 .* deleted_at IS NULL`).
		WithArgs(ts.Id).
		WillReturnRows(pgxmock.NewRows(timespanCols).
			AddRow(ts.Id, ts.Name, ts.StartTime, ts.EndTime))
	expectTimespanTagsQuery(mock, ts.Id, ts.TagIds)

	got, err := repo.GetTimespan(ctx, testScope, ts.Id)
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

	mock.ExpectQuery(`SELECT id, name, start_time, end_time FROM timespans WHERE id = \$1 .* deleted_at IS NULL`).
		WithArgs(id).
		WillReturnRows(pgxmock.NewRows(timespanCols))

	_, err := repo.GetTimespan(ctx, testScope, id)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestGetTimespan_NilId(t *testing.T) {
	repo, _ := newMock(t)
	_, err := repo.GetTimespan(context.Background(), testScope, uuid.Nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

// ── ListTimespans ────────────────────────────────────────────────────────────

func TestListTimespans_ReturnsAll(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)

	ts1, ts2 := aTimespan(), aTimespan()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM timespans WHERE deleted_at IS NULL`).
		WillReturnRows(
			pgxmock.NewRows([]string{"count"}).
				AddRow(2),
		)

	mock.ExpectQuery(`SELECT id, name, start_time, end_time FROM timespans WHERE deleted_at IS NULL ORDER BY start_time DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(25, 0).
		WillReturnRows(
			pgxmock.NewRows(timespanCols).
				AddRow(ts1.Id, ts1.Name, ts1.StartTime, ts1.EndTime).
				AddRow(ts2.Id, ts2.Name, ts2.StartTime, ts2.EndTime),
		)

	expectTimespanTagsQuery(mock, ts1.Id, ts1.TagIds)
	expectTimespanTagsQuery(mock, ts2.Id, ts2.TagIds)

	page, err := repo.ListTimespans(ctx, testScope, model.DefaultPaginationParams())

	require.NoError(t, err)

	assert.Len(t, page.Data, 2)
	assert.Equal(t, 2, page.TotalCount)
}

func TestListTimespans_WithPaginationParams(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)

	ts := aTimespan()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM timespans WHERE deleted_at IS NULL`).
		WillReturnRows(
			pgxmock.NewRows([]string{"count"}).
				AddRow(3),
		)

	mock.ExpectQuery(`SELECT id, name, start_time, end_time FROM timespans WHERE deleted_at IS NULL ORDER BY start_time DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(1, 1).
		WillReturnRows(
			pgxmock.NewRows(timespanCols).
				AddRow(ts.Id, ts.Name, ts.StartTime, ts.EndTime),
		)

	expectTimespanTagsQuery(mock, ts.Id, ts.TagIds)

	page, err := repo.ListTimespans(ctx, testScope, model.PaginationParams{
		Limit:  1,
		Offset: 1,
	})

	require.NoError(t, err)

	assert.Len(t, page.Data, 1)
	assert.Equal(t, 3, page.TotalCount)
	assert.Equal(t, 1, page.Limit)
	assert.Equal(t, 1, page.Offset)
}

func TestListTimespans_Empty(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM timespans WHERE deleted_at IS NULL`).
		WillReturnRows(
			pgxmock.NewRows([]string{"count"}).
				AddRow(0),
		)

	mock.ExpectQuery(`SELECT id, name, start_time, end_time FROM timespans WHERE deleted_at IS NULL ORDER BY start_time DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(25, 0).
		WillReturnRows(
			pgxmock.NewRows(timespanCols),
		)

	page, err := repo.ListTimespans(ctx, testScope, model.DefaultPaginationParams())

	require.NoError(t, err)

	assert.Empty(t, page.Data)
	assert.Equal(t, 0, page.TotalCount)
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

	got, err := repo.CreateTimespan(ctx, testScope, ts)
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

	got, err := repo.CreateTimespan(ctx, testScope, ts)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, got.Id)
}

func TestCreateTimespan_ZeroStartTime(t *testing.T) {
	repo, _ := newMock(t)
	ts := aTimespan()
	ts.StartTime = time.Time{}
	_, err := repo.CreateTimespan(context.Background(), testScope, ts)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

func TestCreateTimespan_EndTimeBeforeStartTime(t *testing.T) {
	repo, _ := newMock(t)
	ts := aTimespan()
	ts.EndTime = ts.StartTime.Add(-1 * time.Minute)
	_, err := repo.CreateTimespan(context.Background(), testScope, ts)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

func TestCreateTimespan_EndTimeEqualToStartTime(t *testing.T) {
	repo, _ := newMock(t)
	ts := aTimespan()
	ts.EndTime = ts.StartTime
	_, err := repo.CreateTimespan(context.Background(), testScope, ts)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

func TestCreateTimespan_ZeroEndTimeAllowed(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	ts := aTimespan()
	ts.EndTime = time.Time{}

	mock.ExpectQuery(`INSERT INTO timespans`).
		WithArgs(ts.Id, ts.Name, ts.StartTime, ts.EndTime).
		WillReturnRows(pgxmock.NewRows(timespanCols).
			AddRow(ts.Id, ts.Name, ts.StartTime, ts.EndTime))
	expectSetTimespanTags(mock, ts.Id, ts.TagIds)

	_, err := repo.CreateTimespan(ctx, testScope, ts)
	require.NoError(t, err)
}

// ── UpdateTimespan ───────────────────────────────────────────────────────────

func TestUpdateTimespan_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	ts := aTimespan()
	ts.Name = "renamed session"

	mock.ExpectQuery(`UPDATE timespans .* WHERE id = \$1 AND deleted_at IS NULL`).
		WithArgs(ts.Id, ts.Name, ts.StartTime, ts.EndTime).
		WillReturnRows(pgxmock.NewRows(timespanCols).
			AddRow(ts.Id, ts.Name, ts.StartTime, ts.EndTime))
	expectSetTimespanTags(mock, ts.Id, ts.TagIds)

	got, err := repo.UpdateTimespan(ctx, testScope, ts)
	require.NoError(t, err)
	assert.Equal(t, "renamed session", got.Name)
	assert.ElementsMatch(t, ts.TagIds, got.TagIds)
}

func TestUpdateTimespan_NotFound(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	ts := aTimespan()

	mock.ExpectQuery(`UPDATE timespans .* WHERE id = \$1 AND deleted_at IS NULL`).
		WithArgs(ts.Id, ts.Name, ts.StartTime, ts.EndTime).
		WillReturnError(pgx.ErrNoRows)

	_, err := repo.UpdateTimespan(ctx, testScope, ts)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestUpdateTimespan_NilId(t *testing.T) {
	repo, _ := newMock(t)
	ts := aTimespan()
	ts.Id = uuid.Nil
	_, err := repo.UpdateTimespan(context.Background(), testScope, ts)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

func TestUpdateTimespan_ZeroStartTime(t *testing.T) {
	repo, _ := newMock(t)
	ts := aTimespan()
	ts.StartTime = time.Time{}
	_, err := repo.UpdateTimespan(context.Background(), testScope, ts)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

func TestUpdateTimespan_EndTimeBeforeStartTime(t *testing.T) {
	repo, _ := newMock(t)
	ts := aTimespan()
	ts.EndTime = ts.StartTime.Add(-time.Second)
	_, err := repo.UpdateTimespan(context.Background(), testScope, ts)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

// ── DeleteTimespan ───────────────────────────────────────────────────────────

func TestDeleteTimespan_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	id := uuid.New()

	mock.ExpectExec(`UPDATE timespans SET deleted_at = now\(\) WHERE id = \$1 AND deleted_at IS NULL`).
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	require.NoError(t, repo.DeleteTimespan(ctx, testScope, id))
}

func TestDeleteTimespan_NotFound(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	id := uuid.New()

	mock.ExpectExec(`UPDATE timespans SET deleted_at = now\(\) WHERE id = \$1 AND deleted_at IS NULL`).
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	err := repo.DeleteTimespan(ctx, testScope, id)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestDeleteTimespan_NilId(t *testing.T) {
	repo, _ := newMock(t)
	err := repo.DeleteTimespan(context.Background(), testScope, uuid.Nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrInvalidArgument))
}

func TestGetTotalDurationByTags_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)

	ids := []uuid.UUID{uuid.New(), uuid.New()}
	mock.ExpectQuery("SELECT .* FROM timespans t .* t.deleted_at IS NULL").
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"total_time"}).AddRow(dur(2 * time.Hour)))

	result, err := repo.GetTotalDurationByTags(ctx, testScope, ids)
	require.NoError(t, err)
	require.Equal(t, 2*time.Hour, result)
}

func TestGetTotalDurationByTags_InvalidTag(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)

	id := []uuid.UUID{uuid.New()}

	mock.ExpectQuery("SELECT .+ FROM timespans t .* t.deleted_at IS NULL").
		WithArgs(id).
		WillReturnRows(pgxmock.NewRows([]string{"total_time"}).AddRow(nil))

	_, err := repo.GetTotalDurationByTags(ctx, testScope, id)
	require.NoError(t, err)
}

func TestGetTotalDurationByTags_EmptyList(t *testing.T) {
	ctx := context.Background()
	repo, _ := newMock(t)

	result, err := repo.GetTotalDurationByTags(ctx, testScope, []uuid.UUID{})
	require.NoError(t, err)
	require.Equal(t, 0*time.Hour, result)
}

func TestGetTotalDurationByTags_NilList(t *testing.T) {
	ctx := context.Background()
	repo, _ := newMock(t)

	out, err := repo.GetTotalDurationByTags(ctx, testScope, nil)
	require.NoError(t, err)
	require.Equal(t, 0*time.Hour, out)
}

func TestAggregateTimeSpentByTagsAndBuckets_Success(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)

	tagIDs := []uuid.UUID{uuid.New(), uuid.New()}
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	buckets := []model.BucketRange{
		{Start: base, End: base.Add(1 * time.Hour)},
		{Start: base.Add(1 * time.Hour), End: base.Add(2 * time.Hour)},
	}

	mock.ExpectQuery(`WITH input_buckets AS`).
		WithArgs(tagIDs, []time.Time{buckets[0].Start, buckets[1].Start}, []time.Time{buckets[0].End, buckets[1].End}).
		WillReturnRows(pgxmock.NewRows([]string{"bucket_start", "bucket_end", "value_seconds"}).
			AddRow(buckets[0].Start, buckets[0].End, float64(45*60)).
			AddRow(buckets[1].Start, buckets[1].End, float64(75*60)))

	got, err := repo.AggregateTimeSpentByTagsAndBuckets(ctx, testScope, tagIDs, buckets)
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, buckets[0], got[0].Bucket)
	require.Equal(t, buckets[1], got[1].Bucket)
	require.InDelta(t, 45*60, got[0].Value, 0.0001)
	require.InDelta(t, 75*60, got[1].Value, 0.0001)
}

func TestAggregateTimeSpentByTagsAndBuckets_InvalidBucket(t *testing.T) {
	repo, _ := newMock(t)

	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	buckets := []model.BucketRange{
		{Start: base, End: base},
	}

	_, err := repo.AggregateTimeSpentByTagsAndBuckets(context.Background(), testScope, []uuid.UUID{uuid.New()}, buckets)
	require.Error(t, err)
	require.ErrorIs(t, err, model.ErrInvalidArgument)
}

func TestAggregateTimeSpentByTagsAndBuckets_EmptyTagsReturnsZeroPerBucket(t *testing.T) {
	repo, _ := newMock(t)

	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	buckets := []model.BucketRange{
		{Start: base.Add(1 * time.Hour), End: base.Add(2 * time.Hour)},
		{Start: base, End: base.Add(1 * time.Hour)},
	}

	got, err := repo.AggregateTimeSpentByTagsAndBuckets(context.Background(), testScope, nil, buckets)
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, buckets[0], got[0].Bucket)
	require.Equal(t, buckets[1], got[1].Bucket)
	require.InDelta(t, 0.0, got[0].Value, 0.0001)
	require.InDelta(t, 0.0, got[1].Value, 0.0001)
}

func TestAggregateTimeSpentByTagsAndBuckets_QueryError(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)

	tagIDs := []uuid.UUID{uuid.New()}
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	buckets := []model.BucketRange{
		{Start: base, End: base.Add(1 * time.Hour)},
	}

	mock.ExpectQuery(`WITH input_buckets AS`).
		WithArgs(tagIDs, []time.Time{buckets[0].Start}, []time.Time{buckets[0].End}).
		WillReturnError(errors.New("db down"))

	_, err := repo.AggregateTimeSpentByTagsAndBuckets(ctx, testScope, tagIDs, buckets)
	require.Error(t, err)
}

func TestAggregateTimeSpentByTagsAndBuckets_UsesCoarseBucketWindowPrefilter(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)

	tagIDs := []uuid.UUID{uuid.New()}
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	buckets := []model.BucketRange{
		{Start: base, End: base.Add(1 * time.Hour)},
	}

	mock.ExpectQuery(`(?s)WITH input_buckets AS.*bucket_window AS.*t.start_time < bw.max_end.*t.end_time > bw.min_start`).
		WithArgs(tagIDs, []time.Time{buckets[0].Start}, []time.Time{buckets[0].End}).
		WillReturnRows(pgxmock.NewRows([]string{"bucket_start", "bucket_end", "value_seconds"}).
			AddRow(buckets[0].Start, buckets[0].End, float64(0)))

	_, err := repo.AggregateTimeSpentByTagsAndBuckets(ctx, testScope, tagIDs, buckets)
	require.NoError(t, err)
}
