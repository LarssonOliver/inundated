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

type mockTimeSpanService struct {
	CreateFn func(ctx context.Context, timespan model.TimeSpan) (model.TimeSpan, error)
	DeleteFn func(ctx context.Context, id uuid.UUID) error
	GetFn    func(ctx context.Context, id uuid.UUID) (model.TimeSpan, error)
	ListFn   func(ctx context.Context) ([]model.TimeSpan, error)
	UpdateFn func(ctx context.Context, timespan model.TimeSpan) (model.TimeSpan, error)
}

var _ service.TimeSpanService = (*mockTimeSpanService)(nil)

// CreateTimeSpan implements [service.TimeSpanService].
func (m *mockTimeSpanService) CreateTimeSpan(ctx context.Context, timespan model.TimeSpan) (model.TimeSpan, error) {
	return m.CreateFn(ctx, timespan)
}

// DeleteTimeSpan implements [service.TimeSpanService].
func (m *mockTimeSpanService) DeleteTimeSpan(ctx context.Context, id uuid.UUID) error {
	return m.DeleteFn(ctx, id)
}

// GetTimeSpan implements [service.TimeSpanService].
func (m *mockTimeSpanService) GetTimeSpan(ctx context.Context, id uuid.UUID) (model.TimeSpan, error) {
	return m.GetFn(ctx, id)
}

// ListTimeSpans implements [service.TimeSpanService].
func (m *mockTimeSpanService) ListTimeSpans(ctx context.Context) ([]model.TimeSpan, error) {
	return m.ListFn(ctx)
}

// UpdateTimeSpan implements [service.TimeSpanService].
func (m *mockTimeSpanService) UpdateTimeSpan(ctx context.Context, timespan model.TimeSpan) (model.TimeSpan, error) {
	return m.UpdateFn(ctx, timespan)
}

func TestTimeSpanHandler_CreateTimeSpan(t *testing.T) {
	baseTime := time.Now()

	tests := []struct {
		name     string
		createFn func(ctx context.Context, timespan model.TimeSpan) (model.TimeSpan, error)
		request  api.CreateTimeSpan
		want     api.TimeSpan
		wantErr  bool
	}{
		{
			name: "successful create",
			createFn: func(ctx context.Context, timespan model.TimeSpan) (model.TimeSpan, error) {
				timespan.Id = uuid.New()
				return timespan, nil
			},
			request: api.CreateTimeSpan{Name: "New TimeSpan", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			want:    api.TimeSpan{Name: "New TimeSpan", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			wantErr: false,
		},
		{
			name: "service error",
			createFn: func(ctx context.Context, timespan model.TimeSpan) (model.TimeSpan, error) {
				return model.TimeSpan{}, errors.New("service error")
			},
			request: api.CreateTimeSpan{Name: "Error TimeSpan", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			want:    api.TimeSpan{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTimeSpanService{
				CreateFn: tt.createFn,
			}

			ta := handlers.NewTimeSpanHandler(svc)

			request := api.CreateTimeSpanRequestObject{
				Body: &tt.request,
			}

			raw, gotErr := ta.CreateTimeSpan(context.Background(), request)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateTimeSpan() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateTimeSpan() succeeded unexpectedly")
			}

			got := raw.(api.CreateTimeSpan201JSONResponse)

			if tt.want.Name != got.Name || tt.want.StartTime != got.StartTime || tt.want.EndTime != got.EndTime {
				t.Errorf("CreateTimeSpan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimeSpanHandler_DeleteTimeSpan(t *testing.T) {
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
			svc := &mockTimeSpanService{
				DeleteFn: tt.deleteFn,
			}

			request := api.DeleteTimeSpanRequestObject{
				TimeSpanId: tt.request,
			}

			ta := handlers.NewTimeSpanHandler(svc)
			_, gotErr := ta.DeleteTimeSpan(context.Background(), request)

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

func TestTimeSpanHandler_GetTimeSpan(t *testing.T) {
	baseTime := time.Now()

	tests := []struct {
		name    string
		getFn   func(ctx context.Context, id uuid.UUID) (model.TimeSpan, error)
		request uuid.UUID
		want    api.TimeSpan
		wantErr bool
	}{
		{
			name: "successful get",
			getFn: func(ctx context.Context, id uuid.UUID) (model.TimeSpan, error) {
				return model.TimeSpan{Id: id, Name: "Sample TimeSpan", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)}, nil
			},
			request: uuid.New(),
			want:    api.TimeSpan{Name: "Sample TimeSpan", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)},
			wantErr: false,
		},
		{
			name: "service error",
			getFn: func(ctx context.Context, id uuid.UUID) (model.TimeSpan, error) {
				return model.TimeSpan{}, errors.New("service error")
			},
			request: uuid.New(),
			want:    api.TimeSpan{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTimeSpanService{
				GetFn: tt.getFn,
			}

			ta := handlers.NewTimeSpanHandler(svc)

			request := api.GetTimeSpanRequestObject{
				TimeSpanId: tt.request,
			}

			got, gotErr := ta.GetTimeSpan(context.Background(), request)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetTimeSpan() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetTimeSpan() succeeded unexpectedly")
			}

			res := got.(api.GetTimeSpan200JSONResponse)
			if res.Id == uuid.Nil || res.Name != tt.want.Name || res.StartTime != tt.want.StartTime || res.EndTime != tt.want.EndTime {
				t.Errorf("GetTimeSpan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimeSpanHandler_ListTimeSpans(t *testing.T) {
	baseTime := time.Now()

	timespan1 := model.TimeSpan{
		Id:        uuid.New(),
		Name:      "backend",
		StartTime: baseTime, EndTime: baseTime.Add(time.Hour),
	}
	timespan2 := model.TimeSpan{
		Id:        uuid.New(),
		Name:      "frontend",
		StartTime: baseTime, EndTime: baseTime.Add(time.Hour),
	}

	tests := []struct {
		name    string
		listFn  func(ctx context.Context) ([]model.TimeSpan, error)
		want    []api.TimeSpan
		wantErr bool
	}{
		{
			name: "success with multiple timespans",
			listFn: func(ctx context.Context) ([]model.TimeSpan, error) {
				return []model.TimeSpan{timespan1, timespan2}, nil
			},
			want: []api.TimeSpan{
				{
					Id:    timespan1.Id,
					Name:  timespan1.Name,
				},
				{
					Id:    timespan2.Id,
					Name:  timespan2.Name,
				},
			},
			wantErr: false,
		},
		{
			name: "success with empty list",
			listFn: func(ctx context.Context) ([]model.TimeSpan, error) {
				return []model.TimeSpan{}, nil
			},
			want:    []api.TimeSpan{},
			wantErr: false,
		},
		{
			name: "service returns error",
			listFn: func(ctx context.Context) ([]model.TimeSpan, error) {
				return nil, errors.New("database unavailable")
			},
			want:    nil,
			wantErr: true,
		},
		// {
		// 	name: "context cancelled",
		// 	listFn: func(ctx context.Context) ([]model.TimeSpan, error) {
		// 		<-ctx.Done()
		// 		return nil, ctx.Err()
		// 	},
		// 	want:    nil,
		// 	wantErr: true,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTimeSpanService{
				ListFn: tt.listFn,
			}
			ta := handlers.NewTimeSpanHandler(svc)
			request := api.ListTimeSpansRequestObject{}
			got, gotErr := ta.ListTimeSpans(context.Background(), request)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ListTimeSpans() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ListTimeSpans() succeeded unexpectedly")
			}
			res := got.(api.ListTimeSpans200JSONResponse)
			if len(res) != len(tt.want) {
				t.Errorf("ListTimeSpans() = %v, want %v", got, tt.want)
				return
			}
			for i, timespan := range res {
				if timespan.Id == uuid.Nil || timespan.Name != tt.want[i].Name {
					t.Errorf("ListTimeSpans() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestTimeSpanHandler_UpdateTimeSpan(t *testing.T) {
	baseTime := time.Now()

	existingID := uuid.New()
	name := "existing-timespan"
	newName := "updated-timespan"

	tests := []struct {
		name      string
		getFn     func(ctx context.Context, id uuid.UUID) (model.TimeSpan, error)
		updateFn  func(ctx context.Context, timespan model.TimeSpan) (model.TimeSpan, error)
		requestId uuid.UUID
		request   api.UpdateTimeSpan
		want      api.TimeSpan
		wantErr   bool
	}{
		{
			name:      "successfully updates timespan",
			requestId: existingID,
			request: api.UpdateTimeSpan{
				Name: &name,
			},
			getFn: func(ctx context.Context, id uuid.UUID) (model.TimeSpan, error) {
				return model.TimeSpan{Id: id, Name: "old-name"}, nil
			},
			updateFn: func(ctx context.Context, timespan model.TimeSpan) (model.TimeSpan, error) {
				return timespan, nil
			},
			want: api.TimeSpan{
				Id:    existingID,
				Name:  name,
			},
			wantErr: false,
		},
		{
			name:      "service returns generic error",
			requestId: existingID,
			request: api.UpdateTimeSpan{
				Name:  &name,
			},
			getFn: func(ctx context.Context, id uuid.UUID) (model.TimeSpan, error) {
				return model.TimeSpan{Id: id, Name: "old-name", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)}, nil
			},
			updateFn: func(ctx context.Context, timespan model.TimeSpan) (model.TimeSpan, error) {
				return model.TimeSpan{}, errors.New("database down")
			},
			wantErr: true,
		},
		{
			name:      "update only name",
			requestId: existingID,
			request: api.UpdateTimeSpan{
				Name: &newName,
			},
			getFn: func(ctx context.Context, id uuid.UUID) (model.TimeSpan, error) {
				return model.TimeSpan{Id: id, Name: "old-name", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)}, nil
			},
			updateFn: func(ctx context.Context, timespan model.TimeSpan) (model.TimeSpan, error) {
				return timespan, nil
			},
			want: api.TimeSpan{
				Id:        existingID,
				Name:      newName,
				StartTime: baseTime, EndTime: baseTime.Add(time.Hour),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTimeSpanService{
				UpdateFn: tt.updateFn,
				GetFn:    tt.getFn,
			}
			ta := handlers.NewTimeSpanHandler(svc)
			request := api.UpdateTimeSpanRequestObject{
				TimeSpanId: tt.requestId,
				Body:       &tt.request,
			}

			got, gotErr := ta.UpdateTimeSpan(context.Background(), request)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UpdateTimeSpan() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UpdateTimeSpan() succeeded unexpectedly")
			}
			res := got.(api.UpdateTimeSpan200JSONResponse)
			if res.Id == uuid.Nil || res.Name != tt.want.Name {
				t.Errorf("UpdateTimeSpan() = %v, want %v", got, tt.want)
			}
		})
	}
}
