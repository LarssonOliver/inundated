package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/service"
)

func TestTimespanService_GetTimespan(t *testing.T) {
	testId := uuid.New()
	baseTime := time.Now()

	tests := []struct {
		name    string
		id      uuid.UUID
		getFn   func(ctx context.Context, id uuid.UUID) (model.Timespan, error)
		want    model.Timespan
		wantErr bool
	}{
		{
			name: "successful get",
			id:   testId,
			getFn: func(ctx context.Context, id uuid.UUID) (model.Timespan, error) {
				return model.Timespan{Id: id, Name: "Test Timespan", StartTime: baseTime, EndTime: baseTime.Add(time.Second)}, nil
			},
			want:    model.Timespan{Id: testId, Name: "Test Timespan", StartTime: baseTime, EndTime: baseTime.Add(time.Second)},
			wantErr: false,
		},
		{
			name: "repository error",
			id:   testId,
			getFn: func(ctx context.Context, id uuid.UUID) (model.Timespan, error) {
				return model.Timespan{}, errors.New("not found")
			},
			want:    model.Timespan{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repository.RepoMock{
				GetTimespanFn: tt.getFn,
			}

			s := service.NewService(repo)
			got, gotErr := s.GetTimespan(context.Background(), tt.id)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetTimespan() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetTimespan() succeeded unexpectedly")
			}
			if got.Id != tt.want.Id || got.Name != tt.want.Name || !got.StartTime.Equal(tt.want.StartTime) || !got.EndTime.Equal(tt.want.EndTime) {
				t.Errorf("GetTimespan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimespanService_ListTimespans(t *testing.T) {
	baseTime := time.Now()

	timespans := []model.Timespan{
		{Id: uuid.New(), Name: "Timespan1", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
		{Id: uuid.New(), Name: "Timespan2", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(3 * time.Hour)},
	}

	tests := []struct {
		name    string
		listFn  func(ctx context.Context) ([]model.Timespan, error)
		want    []model.Timespan
		wantErr bool
	}{
		{
			name: "successful list",
			listFn: func(ctx context.Context) ([]model.Timespan, error) {
				return timespans, nil
			},
			want:    timespans,
			wantErr: false,
		},
		{
			name: "repository error",
			listFn: func(ctx context.Context) ([]model.Timespan, error) {
				return nil, errors.New("database error")
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repository.RepoMock{
				ListTimespanFn: tt.listFn,
			}

			s := service.NewService(repo)
			got, gotErr := s.ListTimespans(context.Background())
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ListTimespans() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ListTimespans() succeeded unexpectedly")
			}
			if len(got) != len(tt.want) {
				t.Errorf("ListTimespans() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i].Id != tt.want[i].Id || got[i].Name != tt.want[i].Name || !got[i].StartTime.Equal(tt.want[i].StartTime) || !got[i].EndTime.Equal(tt.want[i].EndTime) {
					t.Errorf("ListTimespans()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestTimespanService_CreateTimespan(t *testing.T) {
	baseTime := time.Now()

	tests := []struct {
		name     string
		timespan model.Timespan
		createFn func(ctx context.Context, timespan model.Timespan) (model.Timespan, error)
		want     model.Timespan
		wantErr  bool
	}{
		{
			name:     "successful create",
			timespan: model.Timespan{Name: "New Timespan", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			createFn: func(ctx context.Context, timespan model.Timespan) (model.Timespan, error) {
				timespan.Id = uuid.New()
				return timespan, nil
			},
			want:    model.Timespan{Name: "New Timespan", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			wantErr: false,
		},
		{
			name:     "repository error",
			timespan: model.Timespan{Name: "New Timespan", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			createFn: func(ctx context.Context, timespan model.Timespan) (model.Timespan, error) {
				return model.Timespan{}, errors.New("database error")
			},
			want:    model.Timespan{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repository.RepoMock{
				CreateTimespanFn: tt.createFn,
			}
			s := service.NewService(repo)
			got, gotErr := s.CreateTimespan(context.Background(), tt.timespan)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateTimespan() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateTimespan() succeeded unexpectedly")
			}
			if got.Name != tt.want.Name || got.Id == tt.timespan.Id || !got.StartTime.Equal(tt.want.StartTime) || !got.EndTime.Equal(tt.want.EndTime) {
				t.Errorf("CreateTimespan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimespanService_UpdateTimespan(t *testing.T) {
	baseTime := time.Now()
	timespanId := uuid.New()

	tests := []struct {
		name     string
		timespan model.Timespan
		updateFn func(ctx context.Context, timespan model.Timespan) (model.Timespan, error)
		want     model.Timespan
		wantErr  bool
	}{
		{
			name:     "successful update",
			timespan: model.Timespan{Id: timespanId, Name: "Updated Timespan", StartTime: baseTime, EndTime: baseTime.Add(2 * time.Hour)},
			updateFn: func(ctx context.Context, timespan model.Timespan) (model.Timespan, error) {
				return timespan, nil
			},
			want:    model.Timespan{Id: timespanId, Name: "Updated Timespan", StartTime: baseTime, EndTime: baseTime.Add(2 * time.Hour)},
			wantErr: false,
		},
		{
			name:     "repository error",
			timespan: model.Timespan{Id: timespanId, Name: "Updated Timespan", StartTime: baseTime, EndTime: baseTime.Add(2 * time.Hour)},
			updateFn: func(ctx context.Context, timespan model.Timespan) (model.Timespan, error) {
				return model.Timespan{}, errors.New("database error")
			},
			want:    model.Timespan{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repository.RepoMock{
				UpdateTimespanFn: tt.updateFn,
			}
			s := service.NewService(repo)
			got, gotErr := s.UpdateTimespan(context.Background(), tt.timespan)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UpdateTimespan() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UpdateTimespan() succeeded unexpectedly")
			}
			if got.Id != tt.want.Id || got.Name != tt.want.Name || !got.StartTime.Equal(tt.want.StartTime) || !got.EndTime.Equal(tt.want.EndTime) {
				t.Errorf("UpdateTimespan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimespanService_DeleteTimespan(t *testing.T) {
	tests := []struct {
		name     string
		deleteFn func(ctx context.Context, id uuid.UUID) error
		wantErr  bool
	}{
		{
			name: "successful delete",
			deleteFn: func(ctx context.Context, id uuid.UUID) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "repository error",
			deleteFn: func(ctx context.Context, id uuid.UUID) error {
				return errors.New("database error")
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repository.RepoMock{
				DeleteTimespanFn: tt.deleteFn,
			}
			s := service.NewService(repo)
			gotErr := s.DeleteTimespan(context.Background(), uuid.New())
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("DeleteTimespan() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("DeleteTimespan() succeeded unexpectedly")
			}
		})
	}
}
