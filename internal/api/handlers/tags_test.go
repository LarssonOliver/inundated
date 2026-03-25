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

type mockTagService struct {
	CreateFn func(ctx context.Context, tag model.Tag) (model.Tag, error)
	DeleteFn func(ctx context.Context, id uuid.UUID) error
	GetFn    func(ctx context.Context, id uuid.UUID, includes *service.TagServiceGetIncludes) (model.Tag, error)
	ListFn   func(ctx context.Context) ([]model.Tag, error)
	UpdateFn func(ctx context.Context, tag model.Tag) (model.Tag, error)
}

var _ service.TagService = (*mockTagService)(nil)

// CreateTag implements [service.TagService].
func (m *mockTagService) CreateTag(ctx context.Context, tag model.Tag) (model.Tag, error) {
	return m.CreateFn(ctx, tag)
}

// DeleteTag implements [service.TagService].
func (m *mockTagService) DeleteTag(ctx context.Context, id uuid.UUID) error {
	return m.DeleteFn(ctx, id)
}

// GetTag implements [service.TagService].
func (m *mockTagService) GetTag(ctx context.Context, id uuid.UUID, includes *service.TagServiceGetIncludes) (model.Tag, error) {
	return m.GetFn(ctx, id, includes)
}

// ListTags implements [service.TagService].
func (m *mockTagService) ListTags(ctx context.Context) ([]model.Tag, error) {
	return m.ListFn(ctx)
}

// UpdateTag implements [service.TagService].
func (m *mockTagService) UpdateTag(ctx context.Context, tag model.Tag) (model.Tag, error) {
	return m.UpdateFn(ctx, tag)
}

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
			svc := &mockTagService{
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
			svc := &mockTagService{
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
			svc := &mockTagService{
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
	tag1 := model.Tag{
		Id:    uuid.New(),
		Name:  "backend",
		Color: "#ff0000",
	}
	tag2 := model.Tag{
		Id:    uuid.New(),
		Name:  "frontend",
		Color: "#00ff00",
	}

	tests := []struct {
		name    string
		listFn  func(ctx context.Context) ([]model.Tag, error)
		want    []api.Tag
		wantErr bool
	}{
		{
			name: "success with multiple tags",
			listFn: func(ctx context.Context) ([]model.Tag, error) {
				return []model.Tag{tag1, tag2}, nil
			},
			want: []api.Tag{
				{
					Id:    tag1.Id,
					Name:  tag1.Name,
					Color: tag1.Color,
				},
				{
					Id:    tag2.Id,
					Name:  tag2.Name,
					Color: tag2.Color,
				},
			},
			wantErr: false,
		},
		{
			name: "success with empty list",
			listFn: func(ctx context.Context) ([]model.Tag, error) {
				return []model.Tag{}, nil
			},
			want:    []api.Tag{},
			wantErr: false,
		},
		{
			name: "service returns error",
			listFn: func(ctx context.Context) ([]model.Tag, error) {
				return nil, errors.New("database unavailable")
			},
			want:    nil,
			wantErr: true,
		},
		// {
		// 	name: "context cancelled",
		// 	listFn: func(ctx context.Context) ([]model.Tag, error) {
		// 		<-ctx.Done()
		// 		return nil, ctx.Err()
		// 	},
		// 	want:    nil,
		// 	wantErr: true,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTagService{
				ListFn: tt.listFn,
			}
			ta := handlers.NewTagHandler(svc)
			request := api.ListTagsRequestObject{}
			got, gotErr := ta.ListTags(context.Background(), request)
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			res := got.(api.ListTags200JSONResponse)
			require.ElementsMatch(t, tt.want, res)
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
			svc := &mockTagService{
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
