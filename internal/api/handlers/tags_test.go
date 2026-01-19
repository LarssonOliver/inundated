package handlers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/api"
	"github.com/larssonoliver/inundated/internal/api/handlers"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/service"
)

type mockTagService struct {
	CreateFn func(ctx context.Context, tag model.Tag) (model.Tag, error)
	DeleteFn func(ctx context.Context, id uuid.UUID) error
	GetFn    func(ctx context.Context, id uuid.UUID) (model.Tag, error)
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
func (m *mockTagService) GetTag(ctx context.Context, id uuid.UUID) (model.Tag, error) {
	return m.GetFn(ctx, id)
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

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateTag() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateTag() succeeded unexpectedly")
			}

			got := raw.(api.CreateTag201JSONResponse)

			if tt.want.Name != got.Name || tt.want.Color != got.Color {
				t.Errorf("CreateTag() = %v, want %v", got, tt.want)
			}
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

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("DeleteTag() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("DeleteTag() succeeded unexpectedly")
			}
		})
	}
}

func TestTagHandler_GetTag(t *testing.T) {
	tests := []struct {
		name    string
		getFn   func(ctx context.Context, id uuid.UUID) (model.Tag, error)
		request uuid.UUID
		want    api.Tag
		wantErr bool
	}{
		{
			name: "successful get",
			getFn: func(ctx context.Context, id uuid.UUID) (model.Tag, error) {
				return model.Tag{Id: id, Name: "Sample Tag", Color: "#abcdef"}, nil
			},
			request: uuid.New(),
			want:    api.Tag{Name: "Sample Tag", Color: "#abcdef"},
			wantErr: false,
		},
		{
			name: "service error",
			getFn: func(ctx context.Context, id uuid.UUID) (model.Tag, error) {
				return model.Tag{}, errors.New("service error")
			},
			request: uuid.New(),
			want:    api.Tag{},
			wantErr: true,
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

			got, gotErr := ta.GetTag(context.Background(), request)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetTag() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetTag() succeeded unexpectedly")
			}

			res := got.(api.GetTag200JSONResponse)
			if res.Id == uuid.Nil || res.Name != tt.want.Name || res.Color != tt.want.Color {
				t.Errorf("GetTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTagHandler_ListTags(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		svc service.TagService
		// Named input parameters for target function.
		request api.ListTagsRequestObject
		want    api.ListTagsResponseObject
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := handlers.NewTagHandler(tt.svc)
			got, gotErr := ta.ListTags(context.Background(), tt.request)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ListTags() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ListTags() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("ListTags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTagHandler_UpdateTag(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		svc service.TagService
		// Named input parameters for target function.
		request api.UpdateTagRequestObject
		want    api.UpdateTagResponseObject
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := handlers.NewTagHandler(tt.svc)
			got, gotErr := ta.UpdateTag(context.Background(), tt.request)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UpdateTag() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UpdateTag() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("UpdateTag() = %v, want %v", got, tt.want)
			}
		})
	}
}
