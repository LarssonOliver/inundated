package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository/memory"
	"github.com/stretchr/testify/require"
)

func TestTimespanStore_CreateTimespan(t *testing.T) {
	baseTime := time.Now()

	tagIds := []uuid.UUID{uuid.New(), uuid.New()}

	tests := []struct {
		name     string
		timespan model.Timespan
		want     model.Timespan
		wantErr  bool
		errType  error
	}{
		{
			name:     "Test CreateTimespan with valid input",
			timespan: model.Timespan{Name: "Morning", StartTime: baseTime, EndTime: baseTime.Add(2 * time.Hour), TagIds: []uuid.UUID{tagIds[0]}},
			want:     model.Timespan{Name: "Morning", StartTime: baseTime, EndTime: baseTime.Add(2 * time.Hour), TagIds: []uuid.UUID{tagIds[0]}},
			wantErr:  false,
		},
		{
			name:     "Test CreateTimespan with another valid input",
			timespan: model.Timespan{Name: "Afternoon", StartTime: baseTime.Add(3 * time.Hour), EndTime: baseTime.Add(5 * time.Hour), TagIds: nil},
			want:     model.Timespan{Name: "Afternoon", StartTime: baseTime.Add(3 * time.Hour), EndTime: baseTime.Add(5 * time.Hour), TagIds: []uuid.UUID{}},
			wantErr:  false,
		},
		{
			name:     "Test CreateTimespan with empty name",
			timespan: model.Timespan{Name: "", StartTime: baseTime, EndTime: baseTime.Add(1 * time.Hour)},
			want:     model.Timespan{Name: "", StartTime: baseTime, EndTime: baseTime.Add(1 * time.Hour)},
			wantErr:  false,
		},
		{
			name:     "Test CreateTimespan with EndTime before StartTime",
			timespan: model.Timespan{Name: "InvalidTime", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime},
			want:     model.Timespan{},
			wantErr:  true,
			errType:  model.ErrInvalidArgument,
		},
		{
			name:     "Test CreateTimespan with zero StartTime and EndTime",
			timespan: model.Timespan{Name: "ZeroTime", StartTime: time.Time{}, EndTime: time.Time{}},
			want:     model.Timespan{},
			wantErr:  true,
			errType:  model.ErrInvalidArgument,
		},
		{
			name:     "Test CreateTimespan with identical StartTime and EndTime",
			timespan: model.Timespan{Name: "SameTime", StartTime: baseTime, EndTime: baseTime},
			want:     model.Timespan{},
			wantErr:  true,
			errType:  model.ErrInvalidArgument,
		},
		{
			name:     "Test ensure tagIds slice is a copy",
			timespan: model.Timespan{Name: "TagCopy", StartTime: baseTime, EndTime: baseTime.Add(1 * time.Hour), TagIds: tagIds},
			want:     model.Timespan{Name: "TagCopy", StartTime: baseTime, EndTime: baseTime.Add(1 * time.Hour), TagIds: []uuid.UUID{tagIds[0], tagIds[1]}},
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := memory.NewMemoryStore()
			for _, tagId := range tagIds {
				_, err := ta.CreateTag(context.Background(), testScope, model.Tag{Id: tagId, Name: "Tag", Color: "#FFFFFF"})
				require.NoError(t, err)
			}

			got, gotErr := ta.CreateTimespan(context.Background(), testScope, tt.timespan)

			if tt.wantErr {
				require.Error(t, gotErr)
				if tt.errType != nil {
					require.ErrorIs(t, gotErr, tt.errType)
				}
				return
			}

			require.NoError(t, gotErr)
			require.Equal(t, tt.want.Name, got.Name)
			require.True(t, got.StartTime.Equal(tt.want.StartTime))
			require.True(t, got.EndTime.Equal(tt.want.EndTime))
			require.NotEqual(t, uuid.Nil, got.Id)
			require.ElementsMatch(t, tt.want.TagIds, got.TagIds)
		})
	}
}

func TestTimespanStore_GetTimespan(t *testing.T) {
	baseTime := time.Now()

	tagIds := []uuid.UUID{uuid.New(), uuid.New()}

	tests := []struct {
		name           string
		createTimespan model.Timespan
		getId          func(createdTimespan *model.Timespan) uuid.UUID
		want           model.Timespan
		wantErr        bool
		errType        error
	}{
		{
			name:           "Test GetTimespan with existing ID",
			createTimespan: model.Timespan{Name: "Timespan1", StartTime: baseTime, EndTime: baseTime.Add(1 * time.Hour), TagIds: tagIds},
			getId: func(createdTimespan *model.Timespan) uuid.UUID {
				return createdTimespan.Id
			},
			want:    model.Timespan{Name: "Timespan1", StartTime: baseTime, EndTime: baseTime.Add(1 * time.Hour), TagIds: []uuid.UUID{tagIds[0], tagIds[1]}},
			wantErr: false,
		},
		{
			name:           "Test GetTimespan with non-existing ID",
			createTimespan: model.Timespan{Name: "Timespan2", StartTime: baseTime, EndTime: baseTime.Add(2 * time.Hour)},
			getId: func(createdTimespan *model.Timespan) uuid.UUID {
				return uuid.New()
			},
			want:    model.Timespan{},
			wantErr: true,
			errType: model.ErrNotFound,
		},
		{
			name:           "Test GetTimespan with empty UUID",
			createTimespan: model.Timespan{Name: "Timespan3", StartTime: baseTime, EndTime: baseTime.Add(3 * time.Hour)},
			getId: func(createdTimespan *model.Timespan) uuid.UUID {
				return uuid.Nil
			},
			want:    model.Timespan{},
			wantErr: true,
			errType: model.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := memory.NewMemoryStore()
			for _, tagId := range tagIds {
				_, err := ta.CreateTag(context.Background(), testScope, model.Tag{Id: tagId, Name: "Tag", Color: "#FFFFFF"})
				require.NoError(t, err)
			}

			timespan, _ := ta.CreateTimespan(context.Background(), testScope, tt.createTimespan)
			getId := tt.getId(&timespan)
			tt.want.Id = getId

			got, gotErr := ta.GetTimespan(context.Background(), testScope, getId)
			if tt.wantErr {
				require.Error(t, gotErr)
				if tt.errType != nil {
					require.ErrorIs(t, gotErr, tt.errType)
				}
				return
			}

			require.NoError(t, gotErr)
			require.Equal(t, tt.want.Name, got.Name)
			require.True(t, got.StartTime.Equal(tt.want.StartTime))
			require.True(t, got.EndTime.Equal(tt.want.EndTime))
			require.Equal(t, tt.want.Id, got.Id)
			require.ElementsMatch(t, tt.want.TagIds, got.TagIds)
		})
	}
}

func TestTimespanStore_ListTimespans(t *testing.T) {
	baseTime := time.Now()
	tagIds := []uuid.UUID{uuid.New(), uuid.New()}

	tests := []struct {
		name            string
		insertTimespans []model.Timespan
		params          model.PaginationParams
		wantLen         int
		wantTotal       int
		wantErr         bool
		errType         error
	}{
		{
			name: "Test ListTimespans with multiple entries default pagination",
			insertTimespans: []model.Timespan{
				{Name: "Timespan1", StartTime: baseTime, EndTime: baseTime.Add(time.Hour), TagIds: []uuid.UUID{tagIds[0]}},
				{Name: "Timespan2", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(3 * time.Hour)},
			},
			params:    model.DefaultPaginationParams(),
			wantLen:   2,
			wantTotal: 2,
		},
		{
			name:            "Test ListTimespans with no entries",
			insertTimespans: []model.Timespan{},
			params:          model.DefaultPaginationParams(),
			wantLen:         0,
			wantTotal:       0,
		},
		{
			name:            "Test ListTimespans with one entry",
			insertTimespans: []model.Timespan{{Name: "OnlyTimespan", StartTime: baseTime, EndTime: baseTime.Add(30 * time.Minute)}},
			params:          model.DefaultPaginationParams(),
			wantLen:         1,
			wantTotal:       1,
		},
		{
			name: "Test ListTimespans with limit",
			insertTimespans: []model.Timespan{
				{Name: "Timespan1", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
				{Name: "Timespan2", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(3 * time.Hour)},
				{Name: "Timespan3", StartTime: baseTime.Add(4 * time.Hour), EndTime: baseTime.Add(5 * time.Hour)},
			},
			params:    model.PaginationParams{Limit: 2, Offset: 0},
			wantLen:   2,
			wantTotal: 3,
		},
		{
			name: "Test ListTimespans with offset",
			insertTimespans: []model.Timespan{
				{Name: "Timespan1", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
				{Name: "Timespan2", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(3 * time.Hour)},
				{Name: "Timespan3", StartTime: baseTime.Add(4 * time.Hour), EndTime: baseTime.Add(5 * time.Hour)},
			},
			params:    model.PaginationParams{Limit: 10, Offset: 2},
			wantLen:   1,
			wantTotal: 3,
		},
		{
			name: "Test ListTimespans with offset beyond end",
			insertTimespans: []model.Timespan{
				{Name: "Timespan1", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			},
			params:    model.PaginationParams{Limit: 10, Offset: 100},
			wantLen:   0,
			wantTotal: 1,
		},
		{
			name: "Test ListTimespans total count unaffected by limit",
			insertTimespans: []model.Timespan{
				{Name: "Timespan1", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
				{Name: "Timespan2", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(3 * time.Hour)},
				{Name: "Timespan3", StartTime: baseTime.Add(4 * time.Hour), EndTime: baseTime.Add(5 * time.Hour)},
			},
			params:    model.PaginationParams{Limit: 1, Offset: 0},
			wantLen:   1,
			wantTotal: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := memory.NewMemoryStore()

			for _, tagId := range tagIds {
				_, err := ta.CreateTag(context.Background(), testScope, model.Tag{Id: tagId, Name: "Tag", Color: "#FFFFFF"})
				require.NoError(t, err)
			}

			insertedIds := make(map[uuid.UUID]bool)
			for i, timespan := range tt.insertTimespans {
				createdTimespan, _ := ta.CreateTimespan(context.Background(), testScope, timespan)
				tt.insertTimespans[i].Id = createdTimespan.Id
				insertedIds[createdTimespan.Id] = true
			}

			page, gotErr := ta.ListTimespans(context.Background(), testScope, tt.params)
			if tt.wantErr {
				require.Error(t, gotErr)
				if tt.errType != nil {
					require.ErrorIs(t, gotErr, tt.errType)
				}
				return
			}

			require.NoError(t, gotErr)
			require.Len(t, page.Data, tt.wantLen)
			require.Equal(t, tt.wantTotal, page.TotalCount)
			require.Equal(t, tt.params.Limit, page.Limit)
			require.Equal(t, tt.params.Offset, page.Offset)

			for _, ts := range page.Data {
				require.True(t, insertedIds[ts.Id], "returned timespan not in inserted set")
			}
		})
	}
}

func TestTimespanStore_UpdateTimespan(t *testing.T) {
	baseTime := time.Now()

	tagIds := []uuid.UUID{uuid.New(), uuid.New()}

	tests := []struct {
		name           string
		timespan       model.Timespan
		editTimespan   model.Timespan
		editTimespanId func(createdTimespan *model.Timespan) uuid.UUID
		want           model.Timespan
		wantErr        bool
		errType        error
	}{
		{
			name:         "Test UpdateTimespan with existing ID",
			timespan:     model.Timespan{Name: "OldName", StartTime: baseTime, EndTime: baseTime.Add(time.Hour), TagIds: []uuid.UUID{tagIds[0]}},
			editTimespan: model.Timespan{Name: "NewName", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(3 * time.Hour), TagIds: tagIds},
			editTimespanId: func(createdTimespan *model.Timespan) uuid.UUID {
				return createdTimespan.Id
			},
			want:    model.Timespan{Name: "NewName", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(3 * time.Hour), TagIds: []uuid.UUID{tagIds[0], tagIds[1]}},
			wantErr: false,
		},
		{
			name:         "Test UpdateTimespan with non-existing ID",
			timespan:     model.Timespan{Name: "SomeName", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			editTimespan: model.Timespan{Name: "AnotherName", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(3 * time.Hour)},
			editTimespanId: func(createdTimespan *model.Timespan) uuid.UUID {
				return uuid.New()
			},
			want:    model.Timespan{},
			wantErr: true,
			errType: model.ErrNotFound,
		},
		{
			name:         "Test UpdateTimespan with empty UUID",
			timespan:     model.Timespan{Name: "Name1", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			editTimespan: model.Timespan{Name: "Name2", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(3 * time.Hour)},
			editTimespanId: func(createdTimespan *model.Timespan) uuid.UUID {
				return uuid.Nil
			},
			want:    model.Timespan{},
			wantErr: true,
			errType: model.ErrNotFound,
		},
		{
			name:         "Test UpdateTimespan with empty name",
			timespan:     model.Timespan{Name: "ValidName", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			editTimespan: model.Timespan{Name: "", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(3 * time.Hour)},
			editTimespanId: func(createdTimespan *model.Timespan) uuid.UUID {
				return createdTimespan.Id
			},
			want:    model.Timespan{Name: "", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(3 * time.Hour)},
			wantErr: false,
		},
		{
			name:         "Test UpdateTimespan with EndTime before StartTime",
			timespan:     model.Timespan{Name: "ValidName2", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			editTimespan: model.Timespan{Name: "InvalidTime", StartTime: baseTime.Add(3 * time.Hour), EndTime: baseTime.Add(2 * time.Hour)},
			editTimespanId: func(createdTimespan *model.Timespan) uuid.UUID {
				return createdTimespan.Id
			},
			want:    model.Timespan{},
			wantErr: true,
			errType: model.ErrInvalidArgument,
		},
		{
			name:         "Test UpdateTimespan with identical StartTime and EndTime",
			timespan:     model.Timespan{Name: "ValidName3", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			editTimespan: model.Timespan{Name: "SameTime", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(2 * time.Hour)},
			editTimespanId: func(createdTimespan *model.Timespan) uuid.UUID {
				return createdTimespan.Id
			},
			want:    model.Timespan{},
			wantErr: true,
			errType: model.ErrInvalidArgument,
		},
		{
			name:         "Test UpdateTimespan with same name and times",
			timespan:     model.Timespan{Name: "NoChange", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			editTimespan: model.Timespan{Name: "NoChange", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			editTimespanId: func(createdTimespan *model.Timespan) uuid.UUID {
				return createdTimespan.Id
			},
			want:    model.Timespan{Name: "NoChange", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			wantErr: false,
		},
		{
			name:         "Test UpdateTimespan with nil TagIds (should become empty slice)",
			timespan:     model.Timespan{Name: "TagChange", StartTime: baseTime, EndTime: baseTime.Add(time.Hour), TagIds: tagIds},
			editTimespan: model.Timespan{Name: "TagChangeUpdated", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(3 * time.Hour), TagIds: nil},
			editTimespanId: func(createdTimespan *model.Timespan) uuid.UUID {
				return createdTimespan.Id
			},
			want:    model.Timespan{Name: "TagChangeUpdated", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(3 * time.Hour), TagIds: []uuid.UUID{}},
			wantErr: false,
		},
		{
			name:         "Test UpdateTimespan ensuring TagIds slice is a copy",
			timespan:     model.Timespan{Name: "TagCopy", StartTime: baseTime, EndTime: baseTime.Add(time.Hour), TagIds: tagIds},
			editTimespan: model.Timespan{Name: "TagCopyUpdated", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(3 * time.Hour), TagIds: []uuid.UUID{tagIds[0], tagIds[1]}},
			editTimespanId: func(createdTimespan *model.Timespan) uuid.UUID {
				createdTimespan.TagIds = append(createdTimespan.TagIds, uuid.New())
				return createdTimespan.Id
			},
			want:    model.Timespan{Name: "TagCopyUpdated", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(3 * time.Hour), TagIds: []uuid.UUID{tagIds[0], tagIds[1]}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := memory.NewMemoryStore()

			for _, tagId := range tagIds {
				_, err := ta.CreateTag(context.Background(), testScope, model.Tag{Id: tagId, Name: "Tag", Color: "#FFFFFF"})
				require.NoError(t, err)
			}

			insertedTimespan, _ := ta.CreateTimespan(context.Background(), testScope, tt.timespan)
			editId := tt.editTimespanId(&insertedTimespan)

			tt.editTimespan.Id = editId
			tt.want.Id = editId

			got, gotErr := ta.UpdateTimespan(context.Background(), testScope, tt.editTimespan)
			if tt.wantErr {
				require.Error(t, gotErr)
				if tt.errType != nil {
					require.ErrorIs(t, gotErr, tt.errType)
				}
				return
			}

			require.NoError(t, gotErr)
			require.NotNil(t, got)
			require.Equal(t, tt.want.Name, got.Name)
			require.True(t, got.StartTime.Equal(tt.want.StartTime))
			require.True(t, got.EndTime.Equal(tt.want.EndTime))
			require.Equal(t, tt.want.Id, got.Id)
			require.ElementsMatch(t, tt.want.TagIds, got.TagIds)
		})
	}
}

func TestTimespanStore_DeleteTimespan(t *testing.T) {
	baseTime := time.Now()

	tests := []struct {
		name           string
		insertTimespan model.Timespan
		deleteId       func(createdTimespan *model.Timespan) uuid.UUID
		wantErr        bool
		errType        error
	}{
		{
			name:           "Test DeleteTimespan with existing ID",
			insertTimespan: model.Timespan{Name: "Timespan1", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			deleteId: func(createdTimespan *model.Timespan) uuid.UUID {
				return createdTimespan.Id
			},
			wantErr: false,
		},
		{
			name:           "Test DeleteTimespan with non-existing ID",
			insertTimespan: model.Timespan{},
			deleteId: func(createdTimespan *model.Timespan) uuid.UUID {
				return uuid.New()
			},
			wantErr: true,
			errType: model.ErrNotFound,
		},
		{
			name:           "Test DeleteTimespan with empty UUID",
			insertTimespan: model.Timespan{Name: "Timespan3", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			deleteId: func(createdTimespan *model.Timespan) uuid.UUID {
				return uuid.Nil
			},
			wantErr: true,
			errType: model.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := memory.NewMemoryStore()

			timespan, _ := ta.CreateTimespan(context.Background(), testScope, tt.insertTimespan)
			deleteId := tt.deleteId(&timespan)

			gotErr := ta.DeleteTimespan(context.Background(), testScope, deleteId)
			if tt.wantErr {
				require.Error(t, gotErr)
				if tt.errType != nil {
					require.ErrorIs(t, gotErr, tt.errType)
				}
				return
			}

			require.NoError(t, gotErr)
			timespan, err := ta.GetTimespan(context.Background(), testScope, deleteId)
			require.ErrorIs(t, err, model.ErrNotFound)
		})
	}
}

func TestMemoryStore_GetTotalDurationByTags(t *testing.T) {
	m := memory.NewMemoryStore()

	tags := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	for _, tagId := range tags {
		_, err := m.CreateTag(context.Background(), testScope, model.Tag{Id: tagId, Name: "Tag", Color: "#FFFFFF"})
		require.NoError(t, err)
	}

	baseTime := time.Now()
	ts1 := model.Timespan{Name: "T", StartTime: baseTime, EndTime: baseTime.Add(2 * time.Hour), TagIds: []uuid.UUID{tags[0], tags[1]}}
	ts2 := model.Timespan{Name: "T", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(5 * time.Hour), TagIds: []uuid.UUID{tags[1], tags[2]}}
	_, err := m.CreateTimespan(context.Background(), testScope, ts1)
	require.NoError(t, err)
	_, err = m.CreateTimespan(context.Background(), testScope, ts2)
	require.NoError(t, err)

	tests := []struct {
		name    string
		tagIds  []uuid.UUID
		want    time.Duration
		wantErr bool
		errType error
	}{
		{
			name:    "Test GetTotalDurationByTags with one tag",
			tagIds:  []uuid.UUID{tags[0]},
			want:    2 * time.Hour,
			wantErr: false,
		},
		{
			name:    "Test GetTotalDurationByTags with multiple tags",
			tagIds:  []uuid.UUID{tags[0], tags[1]},
			want:    5 * time.Hour, // ts1 has 2 hours with tag[0] and tag[1], ts2 has 3 hours with tag[1]
			wantErr: false,
		},
		{
			name:    "Test GetTotalDurationByTags with non-existing tag",
			tagIds:  []uuid.UUID{uuid.New()},
			want:    0,
			wantErr: false,
		},
		{
			name:    "Test GetTotalDurationByTags with empty tag list",
			tagIds:  []uuid.UUID{},
			want:    0,
			wantErr: false,
		},
		{
			name:    "Test GetTotalDurationByTags with nil tag list",
			tagIds:  nil,
			want:    0,
			wantErr: false,
		},
		{
			name:    "Test GetTotalDurationByTags with duplicate tags",
			tagIds:  []uuid.UUID{tags[1], tags[1]},
			want:    5 * time.Hour, // Should not double count the duration for tag[1]
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := m.GetTotalDurationByTags(context.Background(), testScope, tt.tagIds)
			if tt.wantErr {
				require.Error(t, gotErr)
				if tt.errType != nil {
					require.ErrorIs(t, gotErr, tt.errType)
				}
				return
			}
			require.NoError(t, gotErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestMemoryStore_AggregateTimeSpentByTagsAndBuckets(t *testing.T) {
	m := memory.NewMemoryStore()
	ctx := context.Background()

	tags := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	for _, tagID := range tags {
		_, err := m.CreateTag(ctx, testScope, model.Tag{Id: tagID, Name: "Tag", Color: "#FFFFFF"})
		require.NoError(t, err)
	}

	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Matches both requested tags but must be counted once per bucket.
	_, err := m.CreateTimespan(ctx, testScope, model.Timespan{
		Name:      "T1",
		StartTime: base.Add(15 * time.Minute),
		EndTime:   base.Add(75 * time.Minute),
		TagIds:    []uuid.UUID{tags[0], tags[1]},
	})
	require.NoError(t, err)

	_, err = m.CreateTimespan(ctx, testScope, model.Timespan{
		Name:      "T2",
		StartTime: base.Add(60 * time.Minute),
		EndTime:   base.Add(120 * time.Minute),
		TagIds:    []uuid.UUID{tags[1]},
	})
	require.NoError(t, err)

	// Never included in queries below.
	_, err = m.CreateTimespan(ctx, testScope, model.Timespan{
		Name:      "T3",
		StartTime: base,
		EndTime:   base.Add(2 * time.Hour),
		TagIds:    []uuid.UUID{tags[2]},
	})
	require.NoError(t, err)

	t.Run("split overlap and deduplicate by timespan", func(t *testing.T) {
		buckets := []model.BucketRange{
			{Start: base, End: base.Add(1 * time.Hour)},
			{Start: base.Add(1 * time.Hour), End: base.Add(2 * time.Hour)},
		}

		got, gotErr := m.AggregateTimeSpentByTagsAndBuckets(ctx, testScope, []uuid.UUID{tags[0], tags[1], tags[1]}, buckets)
		require.NoError(t, gotErr)
		require.Len(t, got, 2)
		require.Equal(t, buckets[0], got[0].Bucket)
		require.Equal(t, buckets[1], got[1].Bucket)
		require.InDelta(t, 45*60, got[0].Value, 0.0001)
		require.InDelta(t, 75*60, got[1].Value, 0.0001)
	})

	t.Run("no matching tags yields zeros", func(t *testing.T) {
		buckets := []model.BucketRange{
			{Start: base, End: base.Add(1 * time.Hour)},
			{Start: base.Add(1 * time.Hour), End: base.Add(2 * time.Hour)},
		}

		got, gotErr := m.AggregateTimeSpentByTagsAndBuckets(ctx, testScope, []uuid.UUID{uuid.New()}, buckets)
		require.NoError(t, gotErr)
		require.Len(t, got, 2)
		require.InDelta(t, 0.0, got[0].Value, 0.0001)
		require.InDelta(t, 0.0, got[1].Value, 0.0001)
	})

	t.Run("empty tag filter yields zeros in input order", func(t *testing.T) {
		buckets := []model.BucketRange{
			{Start: base.Add(1 * time.Hour), End: base.Add(2 * time.Hour)},
			{Start: base, End: base.Add(1 * time.Hour)},
		}

		got, gotErr := m.AggregateTimeSpentByTagsAndBuckets(ctx, testScope, nil, buckets)
		require.NoError(t, gotErr)
		require.Len(t, got, 2)
		require.Equal(t, buckets[0], got[0].Bucket)
		require.Equal(t, buckets[1], got[1].Bucket)
		require.InDelta(t, 0.0, got[0].Value, 0.0001)
		require.InDelta(t, 0.0, got[1].Value, 0.0001)
	})

	t.Run("invalid bucket returns invalid argument", func(t *testing.T) {
		buckets := []model.BucketRange{
			{Start: base.Add(2 * time.Hour), End: base.Add(2 * time.Hour)},
		}

		_, gotErr := m.AggregateTimeSpentByTagsAndBuckets(ctx, testScope, []uuid.UUID{tags[0]}, buckets)
		require.ErrorIs(t, gotErr, model.ErrInvalidArgument)
	})
}
