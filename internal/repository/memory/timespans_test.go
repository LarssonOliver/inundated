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
				_, err := ta.CreateTag(context.Background(), model.Tag{Id: tagId, Name: "Tag", Color: "#FFFFFF"})
				require.NoError(t, err)
			}

			got, gotErr := ta.CreateTimespan(context.Background(), tt.timespan)

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
				_, err := ta.CreateTag(context.Background(), model.Tag{Id: tagId, Name: "Tag", Color: "#FFFFFF"})
				require.NoError(t, err)
			}

			timespan, _ := ta.CreateTimespan(context.Background(), tt.createTimespan)
			getId := tt.getId(&timespan)
			tt.want.Id = getId

			got, gotErr := ta.GetTimespan(context.Background(), getId)
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
		name            string // description of this test case
		insertTimespans []model.Timespan
		wantErr         bool
		errType         error
	}{
		{
			name:            "Test ListTimespans with multiple entries",
			insertTimespans: []model.Timespan{{Name: "Timespan1", StartTime: baseTime, EndTime: baseTime.Add(time.Hour), TagIds: []uuid.UUID{tagIds[0]}}, {Name: "Timespan2", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(3 * time.Hour)}},
			wantErr:         false,
		},
		{
			name:            "Test ListTimespans with no entries",
			insertTimespans: []model.Timespan{},
			wantErr:         false,
		},
		{
			name:            "Test ListTimespans with one entry",
			insertTimespans: []model.Timespan{{Name: "OnlyTimespan", StartTime: baseTime, EndTime: baseTime.Add(30 * time.Minute)}},
			wantErr:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := memory.NewMemoryStore()

			for _, tagId := range tagIds {
				_, err := ta.CreateTag(context.Background(), model.Tag{Id: tagId, Name: "Tag", Color: "#FFFFFF"})
				require.NoError(t, err)
			}

			for i, timespan := range tt.insertTimespans {
				createdTimespan, _ := ta.CreateTimespan(context.Background(), timespan)
				tt.insertTimespans[i].Id = createdTimespan.Id
			}

			got, gotErr := ta.ListTimespans(context.Background())
			if tt.wantErr {
				require.Error(t, gotErr)
				if tt.errType != nil {
					require.ErrorIs(t, gotErr, tt.errType)
				}
				return
			}

			require.NoError(t, gotErr)
			require.NotNil(t, got)
			require.Len(t, got, len(tt.insertTimespans))

			for _, timespan := range tt.insertTimespans {
				found := false
				for _, gotTimespan := range got {
					if gotTimespan.Id == timespan.Id && gotTimespan.Name == timespan.Name && timespan.StartTime.Equal(gotTimespan.StartTime) && timespan.EndTime.Equal(gotTimespan.EndTime) && len(gotTimespan.TagIds) == len(timespan.TagIds) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("ListTimespans() missing expected timespan: %v", timespan)
				}
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
				_, err := ta.CreateTag(context.Background(), model.Tag{Id: tagId, Name: "Tag", Color: "#FFFFFF"})
				require.NoError(t, err)
			}

			insertedTimespan, _ := ta.CreateTimespan(context.Background(), tt.timespan)
			editId := tt.editTimespanId(&insertedTimespan)

			tt.editTimespan.Id = editId
			tt.want.Id = editId

			got, gotErr := ta.UpdateTimespan(context.Background(), tt.editTimespan)
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

			timespan, _ := ta.CreateTimespan(context.Background(), tt.insertTimespan)
			deleteId := tt.deleteId(&timespan)

			gotErr := ta.DeleteTimespan(context.Background(), deleteId)
			if tt.wantErr {
				require.Error(t, gotErr)
				if tt.errType != nil {
					require.ErrorIs(t, gotErr, tt.errType)
				}
				return
			}

			require.NoError(t, gotErr)
			timespan, err := ta.GetTimespan(context.Background(), deleteId)
			require.ErrorIs(t, err, model.ErrNotFound)
		})
	}
}

func TestMemoryStore_GetTotalDurationByTags(t *testing.T) {
	m := memory.NewMemoryStore()

	tags := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	for _, tagId := range tags {
		_, err := m.CreateTag(context.Background(), model.Tag{Id: tagId, Name: "Tag", Color: "#FFFFFF"})
		require.NoError(t, err)
	}

	baseTime := time.Now()
	ts1 := model.Timespan{Name: "T", StartTime: baseTime, EndTime: baseTime.Add(2 * time.Hour), TagIds: []uuid.UUID{tags[0], tags[1]}}
	ts2 := model.Timespan{Name: "T", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(5 * time.Hour), TagIds: []uuid.UUID{tags[1], tags[2]}}
	_, err := m.CreateTimespan(context.Background(), ts1)
	require.NoError(t, err)
	_, err = m.CreateTimespan(context.Background(), ts2)
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
			name: "Test GetTotalDurationByTags with duplicate tags",
			tagIds: []uuid.UUID{tags[1], tags[1]},
			want: 5 * time.Hour, // Should not double count the duration for tag[1]
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := m.GetTotalDurationByTags(context.Background(), tt.tagIds)
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
