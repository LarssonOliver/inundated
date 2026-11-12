package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/repository/memory"
	"github.com/larssonoliver/inundated/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalendarService_ListCalendarTimespansReadsEveryPage(t *testing.T) {
	ctx := context.Background()
	store := memory.NewMemoryStore()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	const count = 501
	for i := range count {
		begin := start.Add(time.Duration(i) * time.Hour)
		_, err := store.CreateTimespan(ctx, model.UnownedScope(), model.Timespan{Id: uuid.New(), StartTime: begin, EndTime: begin.Add(time.Minute)})
		require.NoError(t, err)
	}

	timespans, err := service.NewService(store).ListCalendarTimespans(ctx, nil, nil)

	require.NoError(t, err)
	assert.Len(t, timespans, count)
}

func TestCalendarService_ListCalendarTimespansPassesRange(t *testing.T) {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)
	var got model.TimespanListParams
	repo := &repository.RepoMock{
		ListTimespanFn: func(ctx context.Context, scope model.OwnerScope, params model.TimespanListParams) (model.Page[model.Timespan], error) {
			got = params
			return model.Page[model.Timespan]{}, nil
		},
	}

	_, err := service.NewService(repo).ListCalendarTimespans(context.Background(), &from, &to)

	require.NoError(t, err)
	assert.Equal(t, &from, got.From)
	assert.Equal(t, &to, got.To)
}

func TestCalendarService_ListCalendarTasks(t *testing.T) {
	closedSince := time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)
	listed := []model.Task{{Id: uuid.New(), Name: "open"}, {Id: uuid.New(), Name: "recent"}}
	var got model.TaskListParams
	repo := &repository.RepoMock{
		ListTasksFn: func(ctx context.Context, scope model.OwnerScope, params model.TaskListParams) (model.Page[model.Task], error) {
			got = params
			return model.Page[model.Task]{Data: listed, TotalCount: len(listed)}, nil
		},
	}

	tasks, err := service.NewService(repo).ListCalendarTasks(context.Background(), closedSince)

	require.NoError(t, err)
	assert.True(t, got.IncludeClosed)
	assert.Equal(t, &closedSince, got.ClosedSince)
	assert.Equal(t, listed, tasks)
}

func TestCalendarService_ListCalendarTags(t *testing.T) {
	label := model.Tag{Id: uuid.New(), Name: "label"}
	archived := model.Tag{Id: uuid.New(), Name: "archived"}
	var got model.TagListParams
	repo := &repository.RepoMock{
		ListTagFn: func(ctx context.Context, scope model.OwnerScope, params model.TagListParams) (model.Page[model.Tag], error) {
			got = params
			return model.Page[model.Tag]{Data: []model.Tag{label, archived}, TotalCount: 2}, nil
		},
	}

	tags, err := service.NewService(repo).ListCalendarTags(context.Background())

	require.NoError(t, err)
	assert.True(t, got.IncludeArchived)
	assert.Equal(t, []model.TagKind{model.TagKindAll}, got.Kinds)
	assert.Equal(t, map[uuid.UUID]model.Tag{label.Id: label, archived.Id: archived}, tags)
}
