package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository/memory"
)

func TestTimeSpanStore_CreateTimeSpan(t *testing.T) {
	baseTime := time.Now()

	tagIds := []uuid.UUID{uuid.New(), uuid.New()}

	tests := []struct {
		name     string
		timeSpan model.TimeSpan
		want     model.TimeSpan
		wantErr  bool
	}{
		{
			name:     "Test CreateTimeSpan with valid input",
			timeSpan: model.TimeSpan{Name: "Morning", StartTime: baseTime, EndTime: baseTime.Add(2 * time.Hour), TagIds: []uuid.UUID{uuid.MustParse("53e00291-feba-4605-bc3f-fbfe2eefea1b")}},
			want:     model.TimeSpan{Name: "Morning", StartTime: baseTime, EndTime: baseTime.Add(2 * time.Hour), TagIds: []uuid.UUID{uuid.MustParse("53e00291-feba-4605-bc3f-fbfe2eefea1b")}},
			wantErr:  false,
		},
		{
			name:     "Test CreateTimeSpan with another valid input",
			timeSpan: model.TimeSpan{Name: "Afternoon", StartTime: baseTime.Add(3 * time.Hour), EndTime: baseTime.Add(5 * time.Hour), TagIds: nil},
			want:     model.TimeSpan{Name: "Afternoon", StartTime: baseTime.Add(3 * time.Hour), EndTime: baseTime.Add(5 * time.Hour), TagIds: []uuid.UUID{}},
			wantErr:  false,
		},
		{
			name:     "Test CreateTimeSpan with empty name",
			timeSpan: model.TimeSpan{Name: "", StartTime: baseTime, EndTime: baseTime.Add(1 * time.Hour)},
			want:     model.TimeSpan{},
			wantErr:  true,
		},
		{
			name:     "Test CreateTimeSpan with EndTime before StartTime",
			timeSpan: model.TimeSpan{Name: "InvalidTime", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime},
			want:     model.TimeSpan{},
			wantErr:  true,
		},
		{
			name:     "Test CreateTimeSpan with zero StartTime and EndTime",
			timeSpan: model.TimeSpan{Name: "ZeroTime", StartTime: time.Time{}, EndTime: time.Time{}},
			want:     model.TimeSpan{},
			wantErr:  true,
		},
		{
			name:     "Test CreateTimeSpan with identical StartTime and EndTime",
			timeSpan: model.TimeSpan{Name: "SameTime", StartTime: baseTime, EndTime: baseTime},
			want:     model.TimeSpan{},
			wantErr:  true,
		},
		{
			name:     "Test CreateTimeSpan with set ID (should be ignored)",
			timeSpan: model.TimeSpan{Id: uuid.New(), Name: "WithID", StartTime: baseTime, EndTime: baseTime.Add(1 * time.Hour)},
			want:     model.TimeSpan{Name: "WithID", StartTime: baseTime, EndTime: baseTime.Add(1 * time.Hour)},
			wantErr:  false,
		},
		{
			name:     "Test ensure tagIds slice is a copy",
			timeSpan: model.TimeSpan{Name: "TagCopy", StartTime: baseTime, EndTime: baseTime.Add(1 * time.Hour), TagIds: tagIds},
			want:     model.TimeSpan{Name: "TagCopy", StartTime: baseTime, EndTime: baseTime.Add(1 * time.Hour), TagIds: []uuid.UUID{tagIds[0], tagIds[1]}},
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := memory.NewTimeSpanStore()
			got, gotErr := ta.CreateTimeSpan(context.Background(), tt.timeSpan)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateTimeSpan() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateTimeSpan() succeeded unexpectedly")
			}
			if got.Name != tt.want.Name || got.StartTime != tt.want.StartTime || got.EndTime != tt.want.EndTime || got.Id == tt.want.Id {
				t.Errorf("CreateTimeSpan() = %v, want %v", got, tt.want)
			}
			if len(got.TagIds) != len(tt.want.TagIds) {
				t.Errorf("CreateTimeSpan() TagIds length = %v, want %v", len(got.TagIds), len(tt.want.TagIds))
				return
			}
			for i, tagId := range tt.want.TagIds {
				if got.TagIds[i] != tagId {
					t.Errorf("CreateTimeSpan() TagIds = %v, want %v", got.TagIds, tt.want.TagIds)
				}
			}
		})
	}
}

func TestTimeSpanStore_GetTimeSpan(t *testing.T) {
	baseTime := time.Now()

	tagIds := []uuid.UUID{uuid.New(), uuid.New()}

	tests := []struct {
		name           string
		createTimeSpan model.TimeSpan
		getId          func(createdTimeSpan *model.TimeSpan) uuid.UUID
		want           model.TimeSpan
		wantErr        bool
	}{
		{
			name:           "Test GetTimeSpan with existing ID",
			createTimeSpan: model.TimeSpan{Name: "TimeSpan1", StartTime: baseTime, EndTime: baseTime.Add(1 * time.Hour), TagIds: tagIds},
			getId: func(createdTimeSpan *model.TimeSpan) uuid.UUID {
				return createdTimeSpan.Id
			},
			want:    model.TimeSpan{Name: "TimeSpan1", StartTime: baseTime, EndTime: baseTime.Add(1 * time.Hour), TagIds: []uuid.UUID{tagIds[0], tagIds[1]}},
			wantErr: false,
		},
		{
			name:           "Test GetTimeSpan with non-existing ID",
			createTimeSpan: model.TimeSpan{Name: "TimeSpan2", StartTime: baseTime, EndTime: baseTime.Add(2 * time.Hour)},
			getId: func(createdTimeSpan *model.TimeSpan) uuid.UUID {
				return uuid.New()
			},
			want:    model.TimeSpan{},
			wantErr: true,
		},
		{
			name:           "Test GetTimeSpan with empty UUID",
			createTimeSpan: model.TimeSpan{Name: "TimeSpan3", StartTime: baseTime, EndTime: baseTime.Add(3 * time.Hour)},
			getId: func(createdTimeSpan *model.TimeSpan) uuid.UUID {
				return uuid.Nil
			},
			want:    model.TimeSpan{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := memory.NewTimeSpanStore()
			timeSpan, _ := ta.CreateTimeSpan(context.Background(), tt.createTimeSpan)
			getId := tt.getId(&timeSpan)
			tt.want.Id = getId

			got, gotErr := ta.GetTimeSpan(context.Background(), getId)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetTimeSpan() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetTimeSpan() succeeded unexpectedly")
			}
			if got.Name != tt.want.Name || !got.StartTime.Equal(tt.want.StartTime) || !got.EndTime.Equal(tt.want.EndTime) || got.Id != tt.want.Id || len(got.TagIds) != len(tt.want.TagIds) {
				t.Errorf("GetTimeSpan() = %v, want %v", got, tt.want)
				return
			}
			for i, tagId := range tt.want.TagIds {
				if got.TagIds[i] != tagId {
					t.Errorf("CreateTimeSpan() TagIds = %v, want %v", got.TagIds, tt.want.TagIds)
				}
			}
		})
	}
}

func TestTimeSpanStore_ListTimeSpans(t *testing.T) {
	baseTime := time.Now()

	tests := []struct {
		name            string // description of this test case
		insertTimeSpans []model.TimeSpan
		wantErr         bool
	}{
		{
			name:            "Test ListTimeSpans with multiple entries",
			insertTimeSpans: []model.TimeSpan{{Name: "TimeSpan1", StartTime: baseTime, EndTime: baseTime.Add(time.Hour), TagIds: []uuid.UUID{uuid.New()}}, {Name: "TimeSpan2", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(3 * time.Hour)}},
			wantErr:         false,
		},
		{
			name:            "Test ListTimeSpans with no entries",
			insertTimeSpans: []model.TimeSpan{},
			wantErr:         false,
		},
		{
			name:            "Test ListTimeSpans with one entry",
			insertTimeSpans: []model.TimeSpan{{Name: "OnlyTimeSpan", StartTime: baseTime, EndTime: baseTime.Add(30 * time.Minute)}},
			wantErr:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := memory.NewTimeSpanStore()

			for i, timeSpan := range tt.insertTimeSpans {
				createdTimeSpan, _ := ta.CreateTimeSpan(context.Background(), timeSpan)
				tt.insertTimeSpans[i].Id = createdTimeSpan.Id
			}

			got, gotErr := ta.ListTimeSpans(context.Background())
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ListTimeSpans() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ListTimeSpans() succeeded unexpectedly")
			}

			if len(got) != len(tt.insertTimeSpans) {
				t.Errorf("ListTimeSpans() = %v, want %v", got, tt.insertTimeSpans)
				return
			}

			for _, timeSpan := range tt.insertTimeSpans {
				found := false
				for _, gotTimeSpan := range got {
					if gotTimeSpan.Id == timeSpan.Id && gotTimeSpan.Name == timeSpan.Name && timeSpan.StartTime.Equal(gotTimeSpan.StartTime) && timeSpan.EndTime.Equal(gotTimeSpan.EndTime) && len(gotTimeSpan.TagIds) == len(timeSpan.TagIds) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("ListTimeSpans() missing expected timeSpan: %v", timeSpan)
				}
			}
		})
	}
}

func TestTimeSpanStore_UpdateTimeSpan(t *testing.T) {
	baseTime := time.Now()

	tagIds := []uuid.UUID{uuid.New(), uuid.New()}

	tests := []struct {
		name           string
		timeSpan       model.TimeSpan
		editTimeSpan   model.TimeSpan
		editTimeSpanId func(createdTimeSpan *model.TimeSpan) uuid.UUID
		want           model.TimeSpan
		wantErr        bool
	}{
		{
			name:         "Test UpdateTimeSpan with existing ID",
			timeSpan:     model.TimeSpan{Name: "OldName", StartTime: baseTime, EndTime: baseTime.Add(time.Hour), TagIds: []uuid.UUID{tagIds[0]}},
			editTimeSpan: model.TimeSpan{Name: "NewName", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(3 * time.Hour), TagIds: tagIds},
			editTimeSpanId: func(createdTimeSpan *model.TimeSpan) uuid.UUID {
				return createdTimeSpan.Id
			},
			want:    model.TimeSpan{Name: "NewName", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(3 * time.Hour), TagIds: []uuid.UUID{tagIds[0], tagIds[1]}},
			wantErr: false,
		},
		{
			name:         "Test UpdateTimeSpan with non-existing ID",
			timeSpan:     model.TimeSpan{Name: "SomeName", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			editTimeSpan: model.TimeSpan{Name: "AnotherName", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(3 * time.Hour)},
			editTimeSpanId: func(createdTimeSpan *model.TimeSpan) uuid.UUID {
				return uuid.New()
			},
			want:    model.TimeSpan{},
			wantErr: true,
		},
		{
			name:         "Test UpdateTimeSpan with empty UUID",
			timeSpan:     model.TimeSpan{Name: "Name1", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			editTimeSpan: model.TimeSpan{Name: "Name2", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(3 * time.Hour)},
			editTimeSpanId: func(createdTimeSpan *model.TimeSpan) uuid.UUID {
				return uuid.Nil
			},
			want:    model.TimeSpan{},
			wantErr: true,
		},
		{
			name:         "Test UpdateTimeSpan with empty name",
			timeSpan:     model.TimeSpan{Name: "ValidName", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			editTimeSpan: model.TimeSpan{Name: "", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(3 * time.Hour)},
			editTimeSpanId: func(createdTimeSpan *model.TimeSpan) uuid.UUID {
				return createdTimeSpan.Id
			},
			want:    model.TimeSpan{},
			wantErr: true,
		},
		{
			name:         "Test UpdateTimeSpan with EndTime before StartTime",
			timeSpan:     model.TimeSpan{Name: "ValidName2", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			editTimeSpan: model.TimeSpan{Name: "InvalidTime", StartTime: baseTime.Add(3 * time.Hour), EndTime: baseTime.Add(2 * time.Hour)},
			editTimeSpanId: func(createdTimeSpan *model.TimeSpan) uuid.UUID {
				return createdTimeSpan.Id
			},
			want:    model.TimeSpan{},
			wantErr: true,
		},
		{
			name:         "Test UpdateTimeSpan with identical StartTime and EndTime",
			timeSpan:     model.TimeSpan{Name: "ValidName3", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			editTimeSpan: model.TimeSpan{Name: "SameTime", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(2 * time.Hour)},
			editTimeSpanId: func(createdTimeSpan *model.TimeSpan) uuid.UUID {
				return createdTimeSpan.Id
			},
			want:    model.TimeSpan{},
			wantErr: true,
		},
		{
			name:         "Test UpdateTimeSpan with same name and times",
			timeSpan:     model.TimeSpan{Name: "NoChange", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			editTimeSpan: model.TimeSpan{Name: "NoChange", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			editTimeSpanId: func(createdTimeSpan *model.TimeSpan) uuid.UUID {
				return createdTimeSpan.Id
			},
			want:    model.TimeSpan{Name: "NoChange", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			wantErr: false,
		},
		{
			name:         "Test UpdateTimeSpan with nil TagIds (should become empty slice)",
			timeSpan:     model.TimeSpan{Name: "TagChange", StartTime: baseTime, EndTime: baseTime.Add(time.Hour), TagIds: tagIds},
			editTimeSpan: model.TimeSpan{Name: "TagChangeUpdated", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(3 * time.Hour), TagIds: nil},
			editTimeSpanId: func(createdTimeSpan *model.TimeSpan) uuid.UUID {
				return createdTimeSpan.Id
			},
			want:    model.TimeSpan{Name: "TagChangeUpdated", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(3 * time.Hour), TagIds: []uuid.UUID{}},
			wantErr: false,
		},
		{
			name:         "Test UpdateTimeSpan ensuring TagIds slice is a copy",
			timeSpan:     model.TimeSpan{Name: "TagCopy", StartTime: baseTime, EndTime: baseTime.Add(time.Hour), TagIds: tagIds},
			editTimeSpan: model.TimeSpan{Name: "TagCopyUpdated", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(3 * time.Hour), TagIds: []uuid.UUID{tagIds[0], tagIds[1]}},
			editTimeSpanId: func(createdTimeSpan *model.TimeSpan) uuid.UUID {
				createdTimeSpan.TagIds = append(createdTimeSpan.TagIds, uuid.New())
				return createdTimeSpan.Id
			},
			want:    model.TimeSpan{Name: "TagCopyUpdated", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(3 * time.Hour), TagIds: []uuid.UUID{tagIds[0], tagIds[1]}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := memory.NewTimeSpanStore()

			insertedTimeSpan, _ := ta.CreateTimeSpan(context.Background(), tt.timeSpan)
			editId := tt.editTimeSpanId(&insertedTimeSpan)

			tt.editTimeSpan.Id = editId
			tt.want.Id = editId

			got, gotErr := ta.UpdateTimeSpan(context.Background(), tt.editTimeSpan)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UpdateTimeSpan() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UpdateTimeSpan() succeeded unexpectedly")
			}
			if got.Name != tt.want.Name || !got.StartTime.Equal(tt.want.StartTime) || !got.EndTime.Equal(tt.want.EndTime) || got.Id != tt.want.Id || len(got.TagIds) != len(tt.want.TagIds) {
				t.Errorf("UpdateTimeSpan() = %v, want %v", got, tt.want)
				return
			}
			for i, tagId := range tt.want.TagIds {
				if got.TagIds[i] != tagId {
					t.Errorf("UpdateTimeSpan() TagIds = %v, want %v", got.TagIds, tt.want.TagIds)
					return
				}
			}
		})
	}
}

func TestTimeSpanStore_DeleteTimeSpan(t *testing.T) {
	baseTime := time.Now()

	tests := []struct {
		name           string
		insertTimeSpan model.TimeSpan
		deleteId       func(createdTimeSpan *model.TimeSpan) uuid.UUID
		wantErr        bool
	}{
		{
			name:           "Test DeleteTimeSpan with existing ID",
			insertTimeSpan: model.TimeSpan{Name: "TimeSpan1", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			deleteId: func(createdTimeSpan *model.TimeSpan) uuid.UUID {
				return createdTimeSpan.Id
			},
			wantErr: false,
		},
		{
			name:           "Test DeleteTimeSpan with non-existing ID",
			insertTimeSpan: model.TimeSpan{},
			deleteId: func(createdTimeSpan *model.TimeSpan) uuid.UUID {
				return uuid.New()
			},
			wantErr: true,
		},
		{
			name:           "Test DeleteTimeSpan with empty UUID",
			insertTimeSpan: model.TimeSpan{Name: "TimeSpan3", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			deleteId: func(createdTimeSpan *model.TimeSpan) uuid.UUID {
				return uuid.Nil
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := memory.NewTimeSpanStore()

			timeSpan, _ := ta.CreateTimeSpan(context.Background(), tt.insertTimeSpan)
			deleteId := tt.deleteId(&timeSpan)

			gotErr := ta.DeleteTimeSpan(context.Background(), deleteId)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("DeleteTimeSpan() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("DeleteTimeSpan() succeeded unexpectedly")
			}
			timeSpan, err := ta.GetTimeSpan(context.Background(), deleteId)
			if err == nil {
				t.Errorf("TimeSpan with ID %v was not deleted, still exists: %v", deleteId, timeSpan)
			}
		})
	}
}
