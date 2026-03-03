package handlers_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/api"
	"github.com/larssonoliver/inundated/internal/api/handlers"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/service"
)

type mockTimespanService struct {
	CreateFn func(ctx context.Context, timespan model.Timespan) (model.Timespan, error)
	DeleteFn func(ctx context.Context, id uuid.UUID) error
	GetFn    func(ctx context.Context, id uuid.UUID) (model.Timespan, error)
	ListFn   func(ctx context.Context) ([]model.Timespan, error)
	UpdateFn func(ctx context.Context, timespan model.Timespan) (model.Timespan, error)
}

var _ service.TimespanService = (*mockTimespanService)(nil)

// CreateTimespan implements [service.TimespanService].
func (m *mockTimespanService) CreateTimespan(ctx context.Context, timespan model.Timespan) (model.Timespan, error) {
	return m.CreateFn(ctx, timespan)
}

// DeleteTimespan implements [service.TimespanService].
func (m *mockTimespanService) DeleteTimespan(ctx context.Context, id uuid.UUID) error {
	return m.DeleteFn(ctx, id)
}

// GetTimespan implements [service.TimespanService].
func (m *mockTimespanService) GetTimespan(ctx context.Context, id uuid.UUID) (model.Timespan, error) {
	return m.GetFn(ctx, id)
}

// ListTimespans implements [service.TimespanService].
func (m *mockTimespanService) ListTimespans(ctx context.Context) ([]model.Timespan, error) {
	return m.ListFn(ctx)
}

// UpdateTimespan implements [service.TimespanService].
func (m *mockTimespanService) UpdateTimespan(ctx context.Context, timespan model.Timespan) (model.Timespan, error) {
	return m.UpdateFn(ctx, timespan)
}

func TestTimespanHandler_CreateTimespan(t *testing.T) {
	baseTime := time.Now()

	name := "New Timespan"

	tests := []struct {
		name     string
		createFn func(ctx context.Context, timespan model.Timespan) (model.Timespan, error)
		request  api.CreateTimespan
		want     api.Timespan
		wantErr  bool
	}{
		{
			name: "successful create",
			createFn: func(ctx context.Context, timespan model.Timespan) (model.Timespan, error) {
				timespan.Id = uuid.New()
				return timespan, nil
			},
			request: api.CreateTimespan{Name: &name, StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			want:    api.Timespan{Name: &name, StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			wantErr: false,
		},
		{
			name: "service error",
			createFn: func(ctx context.Context, timespan model.Timespan) (model.Timespan, error) {
				return model.Timespan{}, errors.New("service error")
			},
			request: api.CreateTimespan{Name: &name, StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			want:    api.Timespan{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTimespanService{
				CreateFn: tt.createFn,
			}

			ta := handlers.NewTimespanHandler(svc)

			request := api.CreateTimespanRequestObject{
				Body: &tt.request,
			}

			raw, gotErr := ta.CreateTimespan(context.Background(), request)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateTimespan() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateTimespan() succeeded unexpectedly")
			}

			got := raw.(api.CreateTimespan201JSONResponse)

			if *tt.want.Name != *got.Name || tt.want.StartTime != got.StartTime || tt.want.EndTime != got.EndTime {
				t.Errorf("CreateTimespan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimespanHandler_DeleteTimespan(t *testing.T) {
	tests := []struct {
		name     string
		deleteFn func(ctx context.Context, id uuid.UUID) error
		request  uuid.UUID
		wantErr  bool
	}{
		{
			name: "successful delete",
			deleteFn: func(ctx context.Context, id uuid.UUID) error {
				return nil
			},
			request: uuid.New(),
			wantErr: false,
		},
		{
			name: "service error",
			deleteFn: func(ctx context.Context, id uuid.UUID) error {
				return errors.New("service error")
			},
			request: uuid.New(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTimespanService{
				DeleteFn: tt.deleteFn,
			}

			request := api.DeleteTimespanRequestObject{
				TimespanId: tt.request,
			}

			ta := handlers.NewTimespanHandler(svc)
			_, gotErr := ta.DeleteTimespan(context.Background(), request)

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

func TestTimespanHandler_GetTimespan(t *testing.T) {
	baseTime := time.Now()

	name := "Sample Timespan"

	tests := []struct {
		name    string
		getFn   func(ctx context.Context, id uuid.UUID) (model.Timespan, error)
		request uuid.UUID
		want    api.Timespan
		wantErr bool
	}{
		{
			name: "successful get",
			getFn: func(ctx context.Context, id uuid.UUID) (model.Timespan, error) {
				return model.Timespan{Id: id, Name: "Sample Timespan", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)}, nil
			},
			request: uuid.New(),
			want:    api.Timespan{Name: &name, StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			wantErr: false,
		},
		{
			name: "service error",
			getFn: func(ctx context.Context, id uuid.UUID) (model.Timespan, error) {
				return model.Timespan{}, errors.New("service error")
			},
			request: uuid.New(),
			want:    api.Timespan{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTimespanService{
				GetFn: tt.getFn,
			}

			ta := handlers.NewTimespanHandler(svc)

			request := api.GetTimespanRequestObject{
				TimespanId: tt.request,
			}

			got, gotErr := ta.GetTimespan(context.Background(), request)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetTimespan() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetTimespan() succeeded unexpectedly")
			}

			res := got.(api.GetTimespan200JSONResponse)
			if res.Id == uuid.Nil || *res.Name != *tt.want.Name || res.StartTime != tt.want.StartTime || res.EndTime != tt.want.EndTime {
				t.Errorf("GetTimespan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimespanHandler_ListTimespans(t *testing.T) {
	baseTime := time.Now()

	timespan1 := model.Timespan{
		Id:        uuid.New(),
		Name:      "backend",
		StartTime: baseTime, EndTime: baseTime.Add(time.Hour),
	}
	timespan2 := model.Timespan{
		Id:        uuid.New(),
		Name:      "frontend",
		StartTime: baseTime, EndTime: baseTime.Add(time.Hour),
	}

	tests := []struct {
		name    string
		listFn  func(ctx context.Context) ([]model.Timespan, error)
		want    []api.Timespan
		wantErr bool
	}{
		{
			name: "success with multiple timespans",
			listFn: func(ctx context.Context) ([]model.Timespan, error) {
				return []model.Timespan{timespan1, timespan2}, nil
			},
			want: []api.Timespan{
				{
					Id:   timespan1.Id,
					Name: &timespan1.Name,
				},
				{
					Id:   timespan2.Id,
					Name: &timespan2.Name,
				},
			},
			wantErr: false,
		},
		{
			name: "success with empty list",
			listFn: func(ctx context.Context) ([]model.Timespan, error) {
				return []model.Timespan{}, nil
			},
			want:    []api.Timespan{},
			wantErr: false,
		},
		{
			name: "service returns error",
			listFn: func(ctx context.Context) ([]model.Timespan, error) {
				return nil, errors.New("database unavailable")
			},
			want:    nil,
			wantErr: true,
		},
		// {
		// 	name: "context cancelled",
		// 	listFn: func(ctx context.Context) ([]model.Timespan, error) {
		// 		<-ctx.Done()
		// 		return nil, ctx.Err()
		// 	},
		// 	want:    nil,
		// 	wantErr: true,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTimespanService{
				ListFn: tt.listFn,
			}
			ta := handlers.NewTimespanHandler(svc)
			request := api.ListTimespansRequestObject{}
			got, gotErr := ta.ListTimespans(context.Background(), request)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ListTimespans() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ListTimespans() succeeded unexpectedly")
			}
			res := got.(api.ListTimespans200JSONResponse)
			if len(res) != len(tt.want) {
				t.Errorf("ListTimespans() = %v, want %v", got, tt.want)
				return
			}
			for i, timespan := range res {
				if timespan.Id == uuid.Nil || *timespan.Name != *tt.want[i].Name {
					t.Errorf("ListTimespans() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestTimespanHandler_UpdateTimespan(t *testing.T) {
	baseTime := time.Now()

	existingID := uuid.New()
	name := "existing-timespan"
	newName := "updated-timespan"
	tagId := uuid.New()

	tests := []struct {
		name      string
		getFn     func(ctx context.Context, id uuid.UUID) (model.Timespan, error)
		updateFn  func(ctx context.Context, timespan model.Timespan) (model.Timespan, error)
		requestId uuid.UUID
		request   api.UpdateTimespan
		want      api.Timespan
		wantErr   bool
	}{
		{
			name:      "successfully updates timespan",
			requestId: existingID,
			request: api.UpdateTimespan{
				Name:   &name,
				TagIds: &[]uuid.UUID{},
			},
			getFn: func(ctx context.Context, id uuid.UUID) (model.Timespan, error) {
				return model.Timespan{Id: id, Name: "old-name", TagIds: []uuid.UUID{tagId}}, nil
			},
			updateFn: func(ctx context.Context, timespan model.Timespan) (model.Timespan, error) {
				return timespan, nil
			},
			want: api.Timespan{
				Id:     existingID,
				Name:   &name,
				TagIds: nil,
			},
			wantErr: false,
		},
		{
			name:      "service returns generic error",
			requestId: existingID,
			request: api.UpdateTimespan{
				Name: &name,
			},
			getFn: func(ctx context.Context, id uuid.UUID) (model.Timespan, error) {
				return model.Timespan{Id: id, Name: "old-name", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)}, nil
			},
			updateFn: func(ctx context.Context, timespan model.Timespan) (model.Timespan, error) {
				return model.Timespan{}, errors.New("database down")
			},
			wantErr: true,
		},
		{
			name:      "update only name",
			requestId: existingID,
			request: api.UpdateTimespan{
				Name: &newName,
			},
			getFn: func(ctx context.Context, id uuid.UUID) (model.Timespan, error) {
				return model.Timespan{Id: id, Name: "old-name", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)}, nil
			},
			updateFn: func(ctx context.Context, timespan model.Timespan) (model.Timespan, error) {
				return timespan, nil
			},
			want: api.Timespan{
				Id:        existingID,
				Name:      &newName,
				StartTime: baseTime, EndTime: baseTime.Add(time.Hour),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTimespanService{
				UpdateFn: tt.updateFn,
				GetFn:    tt.getFn,
			}
			ta := handlers.NewTimespanHandler(svc)
			request := api.UpdateTimespanRequestObject{
				TimespanId: tt.requestId,
				Body:       &tt.request,
			}

			got, gotErr := ta.UpdateTimespan(context.Background(), request)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UpdateTimespan() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UpdateTimespan() succeeded unexpectedly")
			}
			res := got.(api.UpdateTimespan200JSONResponse)
			if res.Id == uuid.Nil || *res.Name != *tt.want.Name {
				t.Errorf("UpdateTimespan() = %v, want %v", got, tt.want)
			}
			if tt.want.TagIds == nil && res.TagIds != nil && len(*res.TagIds) != 0 {
				t.Errorf("UpdateProject() TagIds = %v, want %v", res.TagIds, tt.want.TagIds)
			}
			if tt.want.TagIds != nil {
				if res.TagIds == nil || len(*res.TagIds) != len(*tt.want.TagIds) {
					t.Errorf("UpdateProject() TagIds = %v, want %v", res.TagIds, tt.want.TagIds)
				}
			}
		})
	}
}
