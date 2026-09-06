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
	"github.com/stretchr/testify/require"
)

func TestTimespanService_GetTimespan(t *testing.T) {
	testId := uuid.New()
	baseTime := time.Now()

	tests := []struct {
		name    string
		id      uuid.UUID
		getFn   func(ctx context.Context, scope model.OwnerScope, id uuid.UUID) (model.Timespan, error)
		want    model.Timespan
		wantErr bool
	}{
		{
			name: "successful get",
			id:   testId,
			getFn: func(ctx context.Context, scope model.OwnerScope, id uuid.UUID) (model.Timespan, error) {
				return model.Timespan{Id: id, Name: "Test Timespan", StartTime: baseTime, EndTime: baseTime.Add(time.Second)}, nil
			},
			want:    model.Timespan{Id: testId, Name: "Test Timespan", StartTime: baseTime, EndTime: baseTime.Add(time.Second)},
			wantErr: false,
		},
		{
			name: "repository error",
			id:   testId,
			getFn: func(ctx context.Context, scope model.OwnerScope, id uuid.UUID) (model.Timespan, error) {
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
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			require.NotEqual(t, uuid.Nil, got.Id)
			require.Equal(t, tt.want.Name, got.Name)
			require.True(t, got.StartTime.Equal(tt.want.StartTime))
			require.True(t, got.EndTime.Equal(tt.want.EndTime))
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
		params  model.PaginationParams
		listFn  func(ctx context.Context, scope model.OwnerScope, params model.PaginationParams) (model.Page[model.Timespan], error)
		want    model.Page[model.Timespan]
		wantErr bool
	}{
		{
			name:   "successful list",
			params: model.DefaultPaginationParams(),
			listFn: func(ctx context.Context, scope model.OwnerScope, params model.PaginationParams) (model.Page[model.Timespan], error) {
				return model.Page[model.Timespan]{Data: timespans, TotalCount: 2}, nil
			},
			want:    model.Page[model.Timespan]{Data: timespans, TotalCount: 2},
			wantErr: false,
		},
		{
			name:   "repository error",
			params: model.DefaultPaginationParams(),
			listFn: func(ctx context.Context, scope model.OwnerScope, params model.PaginationParams) (model.Page[model.Timespan], error) {
				return model.Page[model.Timespan]{}, errors.New("database error")
			},
			wantErr: true,
		},
		{
			name:   "pagination params are forwarded",
			params: model.PaginationParams{Limit: 1, Offset: 1},
			listFn: func(ctx context.Context, scope model.OwnerScope, params model.PaginationParams) (model.Page[model.Timespan], error) {
				require.Equal(t, 1, params.Limit)
				require.Equal(t, 1, params.Offset)
				return model.Page[model.Timespan]{Data: timespans[1:], TotalCount: 2}, nil
			},
			want:    model.Page[model.Timespan]{Data: timespans[1:], TotalCount: 2},
			wantErr: false,
		},
		{
			name:   "empty page",
			params: model.PaginationParams{Limit: 10, Offset: 100},
			listFn: func(ctx context.Context, scope model.OwnerScope, params model.PaginationParams) (model.Page[model.Timespan], error) {
				return model.Page[model.Timespan]{Data: []model.Timespan{}, TotalCount: 2}, nil
			},
			want:    model.Page[model.Timespan]{Data: []model.Timespan{}, TotalCount: 2},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repository.RepoMock{
				ListTimespanFn: tt.listFn,
			}

			s := service.NewService(repo)
			got, gotErr := s.ListTimespans(context.Background(), tt.params)
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			require.Equal(t, tt.want.TotalCount, got.TotalCount)
			require.ElementsMatch(t, tt.want.Data, got.Data)
		})
	}
}

func TestTimespanService_CreateTimespan(t *testing.T) {
	baseTime := time.Now()
	id := uuid.New()

	tests := []struct {
		name     string
		timespan model.Timespan
		createFn func(ctx context.Context, scope model.OwnerScope, timespan model.Timespan) (model.Timespan, error)
		want     model.Timespan
		wantErr  bool
	}{
		{
			name:     "successful create",
			timespan: model.Timespan{Name: "New Timespan", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			createFn: func(ctx context.Context, scope model.OwnerScope, timespan model.Timespan) (model.Timespan, error) {
				timespan.Id = uuid.New()
				return timespan, nil
			},
			want:    model.Timespan{Name: "New Timespan", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			wantErr: false,
		},
		{
			name:     "ensure create generates new ID",
			timespan: model.Timespan{Id: id, Name: "New Timespan", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			createFn: func(ctx context.Context, scope model.OwnerScope, timespan model.Timespan) (model.Timespan, error) {
				return timespan, nil
			},
			want:    model.Timespan{Id: id, Name: "New Timespan", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			wantErr: false,
		},
		{
			name:     "repository error",
			timespan: model.Timespan{Name: "New Timespan", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			createFn: func(ctx context.Context, scope model.OwnerScope, timespan model.Timespan) (model.Timespan, error) {
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
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			require.NotEqual(t, uuid.Nil, got.Id)
			require.NotEqual(t, tt.want.Id, got.Id)
			require.Equal(t, tt.want.Name, got.Name)
			require.True(t, got.StartTime.Equal(tt.want.StartTime))
			require.True(t, got.EndTime.Equal(tt.want.EndTime))
		})
	}
}

func TestTimespanService_UpdateTimespan(t *testing.T) {
	baseTime := time.Now()
	timespanId := uuid.New()

	tests := []struct {
		name     string
		timespan model.Timespan
		updateFn func(ctx context.Context, scope model.OwnerScope, timespan model.Timespan) (model.Timespan, error)
		want     model.Timespan
		wantErr  bool
	}{
		{
			name:     "successful update",
			timespan: model.Timespan{Id: timespanId, Name: "Updated Timespan", StartTime: baseTime, EndTime: baseTime.Add(2 * time.Hour)},
			updateFn: func(ctx context.Context, scope model.OwnerScope, timespan model.Timespan) (model.Timespan, error) {
				return timespan, nil
			},
			want:    model.Timespan{Id: timespanId, Name: "Updated Timespan", StartTime: baseTime, EndTime: baseTime.Add(2 * time.Hour)},
			wantErr: false,
		},
		{
			name:     "repository error",
			timespan: model.Timespan{Id: timespanId, Name: "Updated Timespan", StartTime: baseTime, EndTime: baseTime.Add(2 * time.Hour)},
			updateFn: func(ctx context.Context, scope model.OwnerScope, timespan model.Timespan) (model.Timespan, error) {
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
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			require.NotEqual(t, uuid.Nil, got.Id)
			require.Equal(t, tt.want.Id, got.Id)
			require.Equal(t, tt.want.Name, got.Name)
			require.True(t, got.StartTime.Equal(tt.want.StartTime))
			require.True(t, got.EndTime.Equal(tt.want.EndTime))
		})
	}
}

func TestTimespanService_DeleteTimespan(t *testing.T) {
	tests := []struct {
		name     string
		deleteFn func(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error
		wantErr  bool
	}{
		{
			name: "successful delete",
			deleteFn: func(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "repository error",
			deleteFn: func(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error {
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
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}
