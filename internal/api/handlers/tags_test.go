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

func TestTagHandler_CreateTag(t *testing.T) {
	tests := []struct {
		name     string
		createFn func(ctx context.Context, tag model.Tag) (model.Tag, error)
		request  api.CreateTag
		want     api.Tag
		wantErr  bool
	}{
		{
			name: "successful create",
			createFn: func(ctx context.Context, tag model.Tag) (model.Tag, error) {
				tag.Id = uuid.New()
				return tag, nil
			},
			request: api.CreateTag{Name: "New Tag", Color: "#123456"},
			want:    api.Tag{Name: "New Tag", Color: "#123456"},
			wantErr: false,
		},
		{
			name: "service error",
			createFn: func(ctx context.Context, tag model.Tag) (model.Tag, error) {
				return model.Tag{}, errors.New("service error")
			},
			request: api.CreateTag{Name: "Error Tag", Color: "#654321"},
			want:    api.Tag{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &service.TagServiceMock{
				CreateFn: tt.createFn,
			}

			ta := handlers.NewTagHandler(svc)

			request := api.CreateTagRequestObject{
				Body: &tt.request,
			}

			raw, gotErr := ta.CreateTag(context.Background(), request)
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			got := raw.(api.CreateTag201JSONResponse)

			require.Equal(t, tt.request.Name, got.Name)
			require.Equal(t, tt.request.Color, got.Color)
		})
	}
}

func TestTagHandler_DeleteTag(t *testing.T) {
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
			svc := &service.TagServiceMock{
				DeleteFn: tt.deleteFn,
			}

			request := api.DeleteTagRequestObject{
				TagId: tt.request,
			}

			ta := handlers.NewTagHandler(svc)
			_, gotErr := ta.DeleteTag(context.Background(), request)
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}

func TestTagHandler_GetTag(t *testing.T) {
	duration := 2 * time.Hour
	ms := int(duration.Milliseconds())
	tests := []struct {
		name    string
		getFn   func(ctx context.Context, id uuid.UUID, i *service.TagServiceGetIncludes) (model.Tag, error)
		request uuid.UUID
		include []string
		want    api.Tag
		wantErr bool
	}{
		{
			name: "successful get",
			getFn: func(ctx context.Context, id uuid.UUID, i *service.TagServiceGetIncludes) (model.Tag, error) {
				return model.Tag{Id: id, Name: "Sample Tag", Color: "#abcdef"}, nil
			},
			request: uuid.New(),
			want:    api.Tag{Name: "Sample Tag", Color: "#abcdef"},
			wantErr: false,
		},
		{
			name: "service error",
			getFn: func(ctx context.Context, id uuid.UUID, i *service.TagServiceGetIncludes) (model.Tag, error) {
				return model.Tag{}, errors.New("service error")
			},
			request: uuid.New(),
			want:    api.Tag{},
			wantErr: true,
		},
		{
			name: "include totalTimeMs",
			getFn: func(ctx context.Context, id uuid.UUID, i *service.TagServiceGetIncludes) (model.Tag, error) {
				return model.Tag{Id: id, Name: "Sample Tag", Color: "#abcdef", TotalTime: &duration}, nil
			},
			request: uuid.New(),
			include: []string{"totalTimeMs"},
			want:    api.Tag{Name: "Sample Tag", Color: "#abcdef", TotalTimeMs: &ms},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &service.TagServiceMock{
				GetFn: tt.getFn,
			}

			ta := handlers.NewTagHandler(svc)

			request := api.GetTagRequestObject{
				TagId: tt.request,
			}
			if len(tt.include) > 0 {
				request.Params.Include = &tt.include
			}

			got, gotErr := ta.GetTag(context.Background(), request)
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			res := got.(api.GetTag200JSONResponse)
			require.Equal(t, tt.request, res.Id)
			require.Equal(t, tt.want.Name, res.Name)
			require.Equal(t, tt.want.Color, res.Color)

			if tt.want.TotalTimeMs != nil {
				require.NotNil(t, res.TotalTimeMs)
				require.Equal(t, *tt.want.TotalTimeMs, *res.TotalTimeMs)
			}
		})
	}
}

func TestTagHandler_ListTags(t *testing.T) {
	tag1 := model.Tag{Id: uuid.New(), Name: "backend", Color: "#ff0000"}
	tag2 := model.Tag{Id: uuid.New(), Name: "frontend", Color: "#00ff00"}

	tests := []struct {
		name       string
		listFn     func(ctx context.Context, params model.PaginationParams) (model.Page[model.Tag], error)
		initParams func() *api.ListTagsParams
		wantData   []api.Tag
		wantLimit  int
		wantTotal  int
		wantErr    bool
	}{
		{
			name: "success with multiple tags and no params",
			listFn: func(ctx context.Context, params model.PaginationParams) (model.Page[model.Tag], error) {
				return model.Page[model.Tag]{Data: []model.Tag{tag1, tag2}, TotalCount: 2, Limit: params.Limit, Offset: params.Offset}, nil
			},
			initParams: func() *api.ListTagsParams { return nil },
			wantData: []api.Tag{
				{Id: tag1.Id, Name: tag1.Name, Color: tag1.Color},
				{Id: tag2.Id, Name: tag2.Name, Color: tag2.Color},
			},
			wantLimit: 25,
			wantTotal: 2,
		},
		{
			name: "success with pagination params",
			listFn: func(ctx context.Context, params model.PaginationParams) (model.Page[model.Tag], error) {
				require.Equal(t, 10, params.Limit)
				require.Equal(t, 0, params.Offset)
				return model.Page[model.Tag]{Data: []model.Tag{tag1, tag2}, TotalCount: 2, Limit: params.Limit, Offset: params.Offset}, nil
			},
			initParams: func() *api.ListTagsParams {
				return &api.ListTagsParams{Limit: ptrLimit(10), Offset: ptrOffset(0)}
			},
			wantData: []api.Tag{
				{Id: tag1.Id, Name: tag1.Name, Color: tag1.Color},
				{Id: tag2.Id, Name: tag2.Name, Color: tag2.Color},
			},
			wantLimit: 10,
			wantTotal: 2,
		},
		{
			name: "success with offset",
			listFn: func(ctx context.Context, params model.PaginationParams) (model.Page[model.Tag], error) {
				require.Equal(t, 25, params.Limit)
				require.Equal(t, 1, params.Offset)
				return model.Page[model.Tag]{Data: []model.Tag{tag2}, TotalCount: 2, Limit: params.Limit, Offset: params.Offset}, nil
			},
			initParams: func() *api.ListTagsParams {
				return &api.ListTagsParams{Offset: ptrOffset(1)}
			},
			wantData:  []api.Tag{{Id: tag2.Id, Name: tag2.Name, Color: tag2.Color}},
			wantLimit: 25,
			wantTotal: 2,
		},
		{
			name: "success with empty list",
			listFn: func(ctx context.Context, params model.PaginationParams) (model.Page[model.Tag], error) {
				return model.Page[model.Tag]{Data: []model.Tag{}, TotalCount: 0, Limit: params.Limit, Offset: params.Offset}, nil
			},
			initParams: func() *api.ListTagsParams { return nil },
			wantData:   []api.Tag{},
			wantLimit:  25,
			wantTotal:  0,
		},
		{
			name: "limit too low returns 400",
			listFn: func(ctx context.Context, params model.PaginationParams) (model.Page[model.Tag], error) {
				t.Fatal("service should not be called")
				return model.Page[model.Tag]{}, nil
			},
			initParams: func() *api.ListTagsParams {
				return &api.ListTagsParams{Limit: ptrLimit(0)}
			},
		},
		{
			name: "limit too high returns 400",
			listFn: func(ctx context.Context, params model.PaginationParams) (model.Page[model.Tag], error) {
				t.Fatal("service should not be called")
				return model.Page[model.Tag]{}, nil
			},
			initParams: func() *api.ListTagsParams {
				return &api.ListTagsParams{Limit: ptrLimit(101)}
			},
		},
		{
			name: "negative offset returns 400",
			listFn: func(ctx context.Context, params model.PaginationParams) (model.Page[model.Tag], error) {
				t.Fatal("service should not be called")
				return model.Page[model.Tag]{}, nil
			},
			initParams: func() *api.ListTagsParams {
				return &api.ListTagsParams{Offset: ptrOffset(-1)}
			},
		},
		{
			name: "service returns error",
			listFn: func(ctx context.Context, params model.PaginationParams) (model.Page[model.Tag], error) {
				return model.Page[model.Tag]{}, errors.New("database unavailable")
			},
			initParams: func() *api.ListTagsParams { return nil },
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &service.TagServiceMock{ListFn: tt.listFn}
			ta := handlers.NewTagHandler(svc)
			params := tt.initParams()
			request := api.ListTagsRequestObject{}
			if params != nil {
				request.Params = *params
			}

			got, gotErr := ta.ListTags(context.Background(), request)
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)

			if _, ok := got.(api.ListTags400Response); ok {
				return
			}

			res := got.(api.ListTags200JSONResponse)
			require.Len(t, res.Data, len(tt.wantData))
			require.Equal(t, tt.wantLimit, res.Pagination.Limit)
			require.Equal(t, tt.wantTotal, res.Pagination.Total)
			for i, tag := range res.Data {
				require.Equal(t, tt.wantData[i].Id, tag.Id)
				require.Equal(t, tt.wantData[i].Name, tag.Name)
				require.Equal(t, tt.wantData[i].Color, tag.Color)
			}
		})
	}
}

func TestTagHandler_UpdateTag(t *testing.T) {
	existingID := uuid.New()
	name := "existing-tag"
	color := "#ffffff"
	newName := "updated-tag"

	tests := []struct {
		name      string
		getFn     func(ctx context.Context, id uuid.UUID, i *service.TagServiceGetIncludes) (model.Tag, error)
		updateFn  func(ctx context.Context, tag model.Tag) (model.Tag, error)
		requestId uuid.UUID
		request   api.UpdateTag
		want      api.Tag
		wantErr   bool
	}{
		{
			name:      "successfully updates tag",
			requestId: existingID,
			request: api.UpdateTag{
				Name:  &name,
				Color: &color,
			},
			getFn: func(ctx context.Context, id uuid.UUID, i *service.TagServiceGetIncludes) (model.Tag, error) {
				return model.Tag{Id: id, Name: "old-name", Color: "#000000"}, nil
			},
			updateFn: func(ctx context.Context, tag model.Tag) (model.Tag, error) {
				return tag, nil
			},
			want: api.Tag{
				Id:    existingID,
				Name:  name,
				Color: color,
			},
			wantErr: false,
		},
		{
			name:      "service returns generic error",
			requestId: existingID,
			request: api.UpdateTag{
				Name:  &name,
				Color: &color,
			},
			getFn: func(ctx context.Context, id uuid.UUID, i *service.TagServiceGetIncludes) (model.Tag, error) {
				return model.Tag{Id: id, Name: "old-name", Color: "#000000"}, nil
			},
			updateFn: func(ctx context.Context, tag model.Tag) (model.Tag, error) {
				return model.Tag{}, errors.New("database down")
			},
			wantErr: true,
		},
		{
			name:      "update only name",
			requestId: existingID,
			request: api.UpdateTag{
				Name: &newName,
			},
			getFn: func(ctx context.Context, id uuid.UUID, i *service.TagServiceGetIncludes) (model.Tag, error) {
				return model.Tag{Id: id, Name: "old-name", Color: "#000000"}, nil
			},
			updateFn: func(ctx context.Context, tag model.Tag) (model.Tag, error) {
				return tag, nil
			},
			want: api.Tag{
				Id:    existingID,
				Name:  newName,
				Color: "#000000",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &service.TagServiceMock{
				UpdateFn: tt.updateFn,
				GetFn:    tt.getFn,
			}
			ta := handlers.NewTagHandler(svc)
			request := api.UpdateTagRequestObject{
				TagId: tt.requestId,
				Body:  &tt.request,
			}

			got, gotErr := ta.UpdateTag(context.Background(), request)
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			res := got.(api.UpdateTag200JSONResponse)
			require.Equal(t, tt.requestId, res.Id)
			require.Equal(t, tt.want.Name, res.Name)
			require.Equal(t, tt.want.Color, res.Color)
		})
	}
}
