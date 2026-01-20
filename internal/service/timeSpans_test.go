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

var _ repository.TimeSpanRepository = (*timeSpanRepoMock)(nil)

type timeSpanRepoMock struct {
	CreateFn func(ctx context.Context, timeSpan model.TimeSpan) (model.TimeSpan, error)
	DeleteFn func(ctx context.Context, id uuid.UUID) error
	GetFn    func(ctx context.Context, id uuid.UUID) (model.TimeSpan, error)
	ListFn   func(ctx context.Context) ([]model.TimeSpan, error)
	UpdateFn func(ctx context.Context, timeSpan model.TimeSpan) (model.TimeSpan, error)
}

// CreateTimeSpan implements repository.TimeSpanRepository.
func (t *timeSpanRepoMock) CreateTimeSpan(ctx context.Context, timeSpan model.TimeSpan) (model.TimeSpan, error) {
	return t.CreateFn(ctx, timeSpan)
}

// DeleteTimeSpan implements repository.TimeSpanRepository.
func (t *timeSpanRepoMock) DeleteTimeSpan(ctx context.Context, id uuid.UUID) error {
	return t.DeleteFn(ctx, id)
}

// GetTimeSpan implements repository.TimeSpanRepository.
func (t *timeSpanRepoMock) GetTimeSpan(ctx context.Context, id uuid.UUID) (model.TimeSpan, error) {
	return t.GetFn(ctx, id)
}

// ListTimeSpans implements repository.TimeSpanRepository.
func (t *timeSpanRepoMock) ListTimeSpans(ctx context.Context) ([]model.TimeSpan, error) {
	return t.ListFn(ctx)
}

// UpdateTimeSpan implements repository.TimeSpanRepository.
func (t *timeSpanRepoMock) UpdateTimeSpan(ctx context.Context, timeSpan model.TimeSpan) (model.TimeSpan, error) {
	return t.UpdateFn(ctx, timeSpan)
}

func TestTimeSpanService_GetTimeSpan(t *testing.T) {
	testId := uuid.New()
	baseTime := time.Now()

	tests := []struct {
		name    string
		id      uuid.UUID
		getFn   func(ctx context.Context, id uuid.UUID) (model.TimeSpan, error)
		want    model.TimeSpan
		wantErr bool
	}{
		{
			name: "successful get",
			id:   testId,
			getFn: func(ctx context.Context, id uuid.UUID) (model.TimeSpan, error) {
				return model.TimeSpan{Id: id, Name: "Test TimeSpan", StartTime: baseTime, EndTime: baseTime.Add(time.Second)}, nil
			},
			want:    model.TimeSpan{Id: testId, Name: "Test TimeSpan", StartTime: baseTime, EndTime: baseTime.Add(time.Second)},
			wantErr: false,
		},
		{
			name: "repository error",
			id:   testId,
			getFn: func(ctx context.Context, id uuid.UUID) (model.TimeSpan, error) {
				return model.TimeSpan{}, errors.New("not found")
			},
			want:    model.TimeSpan{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &timeSpanRepoMock{
				GetFn: tt.getFn,
			}

			s := service.NewTimeSpanService(repo)
			got, gotErr := s.GetTimeSpan(context.Background(), tt.id)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetTimeSpan() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetTimeSpan() succeeded unexpectedly")
			}
			if got.Id != tt.want.Id || got.Name != tt.want.Name || !got.StartTime.Equal(tt.want.StartTime) || !got.EndTime.Equal(tt.want.EndTime) {
				t.Errorf("GetTimeSpan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimeSpanService_ListTimeSpans(t *testing.T) {
	baseTime := time.Now()

	timeSpans := []model.TimeSpan{
		{Id: uuid.New(), Name: "TimeSpan1", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
		{Id: uuid.New(), Name: "TimeSpan2", StartTime: baseTime.Add(2 * time.Hour), EndTime: baseTime.Add(3 * time.Hour)},
	}

	tests := []struct {
		name    string
		listFn  func(ctx context.Context) ([]model.TimeSpan, error)
		want    []model.TimeSpan
		wantErr bool
	}{
		{
			name: "successful list",
			listFn: func(ctx context.Context) ([]model.TimeSpan, error) {
				return timeSpans, nil
			},
			want:    timeSpans,
			wantErr: false,
		},
		{
			name: "repository error",
			listFn: func(ctx context.Context) ([]model.TimeSpan, error) {
				return nil, errors.New("database error")
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &timeSpanRepoMock{
				ListFn: tt.listFn,
			}

			s := service.NewTimeSpanService(repo)
			got, gotErr := s.ListTimeSpans(context.Background())
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ListTimeSpans() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ListTimeSpans() succeeded unexpectedly")
			}
			if len(got) != len(tt.want) {
				t.Errorf("ListTimeSpans() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i].Id != tt.want[i].Id || got[i].Name != tt.want[i].Name || !got[i].StartTime.Equal(tt.want[i].StartTime) || !got[i].EndTime.Equal(tt.want[i].EndTime) {
					t.Errorf("ListTimeSpans()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestTimeSpanService_CreateTimeSpan(t *testing.T) {
	baseTime := time.Now()

	tests := []struct {
		name     string
		timeSpan model.TimeSpan
		createFn func(ctx context.Context, timeSpan model.TimeSpan) (model.TimeSpan, error)
		want     model.TimeSpan
		wantErr  bool
	}{
		{
			name:     "successful create",
			timeSpan: model.TimeSpan{Name: "New TimeSpan", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			createFn: func(ctx context.Context, timeSpan model.TimeSpan) (model.TimeSpan, error) {
				timeSpan.Id = uuid.New()
				return timeSpan, nil
			},
			want:    model.TimeSpan{Name: "New TimeSpan", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			wantErr: false,
		},
		{
			name:     "repository error",
			timeSpan: model.TimeSpan{Name: "New TimeSpan", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			createFn: func(ctx context.Context, timeSpan model.TimeSpan) (model.TimeSpan, error) {
				return model.TimeSpan{}, errors.New("database error")
			},
			want:    model.TimeSpan{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &timeSpanRepoMock{
				CreateFn: tt.createFn,
			}
			s := service.NewTimeSpanService(repo)
			got, gotErr := s.CreateTimeSpan(context.Background(), tt.timeSpan)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateTimeSpan() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateTimeSpan() succeeded unexpectedly")
			}
			if got.Name != tt.want.Name || got.Id == tt.timeSpan.Id || !got.StartTime.Equal(tt.want.StartTime) || !got.EndTime.Equal(tt.want.EndTime) {
				t.Errorf("CreateTimeSpan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimeSpanService_UpdateTimeSpan(t *testing.T) {
	baseTime := time.Now()
	timeSpanId := uuid.New()

	tests := []struct {
		name     string
		timeSpan model.TimeSpan
		updateFn func(ctx context.Context, timeSpan model.TimeSpan) (model.TimeSpan, error)
		want     model.TimeSpan
		wantErr  bool
	}{
		{
			name:     "successful update",
			timeSpan: model.TimeSpan{Id: timeSpanId, Name: "Updated TimeSpan", StartTime: baseTime, EndTime: baseTime.Add(2 * time.Hour)},
			updateFn: func(ctx context.Context, timeSpan model.TimeSpan) (model.TimeSpan, error) {
				return timeSpan, nil
			},
			want:    model.TimeSpan{Id: timeSpanId, Name: "Updated TimeSpan", StartTime: baseTime, EndTime: baseTime.Add(2 * time.Hour)},
			wantErr: false,
		},
		{
			name:     "repository error",
			timeSpan: model.TimeSpan{Id: timeSpanId, Name: "Updated TimeSpan", StartTime: baseTime, EndTime: baseTime.Add(2 * time.Hour)},
			updateFn: func(ctx context.Context, timeSpan model.TimeSpan) (model.TimeSpan, error) {
				return model.TimeSpan{}, errors.New("database error")
			},
			want:    model.TimeSpan{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &timeSpanRepoMock{
				UpdateFn: tt.updateFn,
			}
			s := service.NewTimeSpanService(repo)
			got, gotErr := s.UpdateTimeSpan(context.Background(), tt.timeSpan)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UpdateTimeSpan() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UpdateTimeSpan() succeeded unexpectedly")
			}
			if got.Id != tt.want.Id || got.Name != tt.want.Name || !got.StartTime.Equal(tt.want.StartTime) || !got.EndTime.Equal(tt.want.EndTime) {
				t.Errorf("UpdateTimeSpan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimeSpanService_DeleteTimeSpan(t *testing.T) {
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
			repo := &timeSpanRepoMock{
				DeleteFn: tt.deleteFn,
			}
			s := service.NewTimeSpanService(repo)
			gotErr := s.DeleteTimeSpan(context.Background(), uuid.New())
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("DeleteTimeSpan() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("DeleteTimeSpan() succeeded unexpectedly")
			}
		})
	}
}
