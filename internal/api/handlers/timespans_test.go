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
	"github.com/stretchr/testify/require"
)

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
			svc := &service.TimespanServiceMock{
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
			svc := &service.TimespanServiceMock{
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
			svc := &service.TimespanServiceMock{
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

	timespan1 := model.Timespan{Id: uuid.New(), Name: "backend", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)}
	timespan2 := model.Timespan{Id: uuid.New(), Name: "frontend", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)}

	tests := []struct {
		name       string
		listFn     func(ctx context.Context, params model.PaginationParams) (model.Page[model.Timespan], error)
		initParams func() *api.ListTimespansParams
		wantData   []api.Timespan
		wantLimit  int
		wantTotal  int
		wantErr    bool
	}{
		{
			name: "success with multiple timespans and no params",
			listFn: func(ctx context.Context, params model.PaginationParams) (model.Page[model.Timespan], error) {
				return model.Page[model.Timespan]{Data: []model.Timespan{timespan1, timespan2}, TotalCount: 2, Limit: params.Limit, Offset: params.Offset}, nil
			},
			initParams: func() *api.ListTimespansParams { return nil },
			wantData: []api.Timespan{
				{Id: timespan1.Id, Name: &timespan1.Name},
				{Id: timespan2.Id, Name: &timespan2.Name},
			},
			wantLimit: 25,
			wantTotal: 2,
		},
		{
			name: "success with pagination params",
			listFn: func(ctx context.Context, params model.PaginationParams) (model.Page[model.Timespan], error) {
				require.Equal(t, 10, params.Limit)
				require.Equal(t, 0, params.Offset)
				return model.Page[model.Timespan]{Data: []model.Timespan{timespan1, timespan2}, TotalCount: 2, Limit: params.Limit, Offset: params.Offset}, nil
			},
			initParams: func() *api.ListTimespansParams {
				return &api.ListTimespansParams{Limit: ptrLimit(10), Offset: ptrOffset(0)}
			},
			wantData: []api.Timespan{
				{Id: timespan1.Id, Name: &timespan1.Name},
				{Id: timespan2.Id, Name: &timespan2.Name},
			},
			wantLimit: 10,
			wantTotal: 2,
		},
		{
			name: "success with offset",
			listFn: func(ctx context.Context, params model.PaginationParams) (model.Page[model.Timespan], error) {
				require.Equal(t, 25, params.Limit)
				require.Equal(t, 1, params.Offset)
				return model.Page[model.Timespan]{Data: []model.Timespan{timespan2}, TotalCount: 2, Limit: params.Limit, Offset: params.Offset}, nil
			},
			initParams: func() *api.ListTimespansParams {
				return &api.ListTimespansParams{Offset: ptrOffset(1)}
			},
			wantData:  []api.Timespan{{Id: timespan2.Id, Name: &timespan2.Name}},
			wantLimit: 25,
			wantTotal: 2,
		},
		{
			name: "success with empty list",
			listFn: func(ctx context.Context, params model.PaginationParams) (model.Page[model.Timespan], error) {
				return model.Page[model.Timespan]{Data: []model.Timespan{}, TotalCount: 0, Limit: params.Limit, Offset: params.Offset}, nil
			},
			initParams: func() *api.ListTimespansParams { return nil },
			wantData:   []api.Timespan{},
			wantLimit:  25,
			wantTotal:  0,
		},
		{
			name: "limit too low returns 400",
			listFn: func(ctx context.Context, params model.PaginationParams) (model.Page[model.Timespan], error) {
				t.Fatal("service should not be called")
				return model.Page[model.Timespan]{}, nil
			},
			initParams: func() *api.ListTimespansParams {
				return &api.ListTimespansParams{Limit: ptrLimit(0)}
			},
		},
		{
			name: "limit too high returns 400",
			listFn: func(ctx context.Context, params model.PaginationParams) (model.Page[model.Timespan], error) {
				t.Fatal("service should not be called")
				return model.Page[model.Timespan]{}, nil
			},
			initParams: func() *api.ListTimespansParams {
				return &api.ListTimespansParams{Limit: ptrLimit(101)}
			},
		},
		{
			name: "negative offset returns 400",
			listFn: func(ctx context.Context, params model.PaginationParams) (model.Page[model.Timespan], error) {
				t.Fatal("service should not be called")
				return model.Page[model.Timespan]{}, nil
			},
			initParams: func() *api.ListTimespansParams {
				return &api.ListTimespansParams{Offset: ptrOffset(-1)}
			},
		},
		{
			name: "service returns error",
			listFn: func(ctx context.Context, params model.PaginationParams) (model.Page[model.Timespan], error) {
				return model.Page[model.Timespan]{}, errors.New("database unavailable")
			},
			initParams: func() *api.ListTimespansParams { return nil },
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &service.TimespanServiceMock{ListFn: tt.listFn}
			ta := handlers.NewTimespanHandler(svc)
			params := tt.initParams()
			request := api.ListTimespansRequestObject{}
			if params != nil {
				request.Params = *params
			}

			got, gotErr := ta.ListTimespans(context.Background(), request)
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)

			if _, ok := got.(api.ListTimespans400Response); ok {
				return
			}

			res := got.(api.ListTimespans200JSONResponse)
			require.Len(t, res.Data, len(tt.wantData))
			require.Equal(t, tt.wantLimit, res.Pagination.Limit)
			require.Equal(t, tt.wantTotal, res.Pagination.Total)
			for i, ts := range res.Data {
				require.NotEqual(t, uuid.Nil, ts.Id)
				require.Equal(t, *tt.wantData[i].Name, *ts.Name)
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
			svc := &service.TimespanServiceMock{
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
