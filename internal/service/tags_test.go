package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/service"
)

func TestTagService_GetTag(t *testing.T) {
	testId := uuid.New()

	tests := []struct {
		name    string
		id      uuid.UUID
		getFn   func(ctx context.Context, id uuid.UUID) (model.Tag, error)
		want    model.Tag
		wantErr bool
	}{
		{
			name: "successful get",
			id:   testId,
			getFn: func(ctx context.Context, id uuid.UUID) (model.Tag, error) {
				return model.Tag{Id: id, Name: "Test Tag", Color: "#abcdef"}, nil
			},
			want:    model.Tag{Id: testId, Name: "Test Tag", Color: "#abcdef"},
			wantErr: false,
		},
		{
			name: "repository error",
			id:   testId,
			getFn: func(ctx context.Context, id uuid.UUID) (model.Tag, error) {
				return model.Tag{}, errors.New("not found")
			},
			want:    model.Tag{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repository.RepoMock{
				GetTagFn: tt.getFn,
			}

			s := service.NewService(repo)
			got, gotErr := s.GetTag(context.Background(), tt.id)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetTag() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetTag() succeeded unexpectedly")
			}
			if got.Id != tt.want.Id || got.Name != tt.want.Name || got.Color != tt.want.Color {
				t.Errorf("GetTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTagService_ListTags(t *testing.T) {
	tags := []model.Tag{
		{Id: uuid.New(), Name: "Tag1", Color: "#ff0000"},
		{Id: uuid.New(), Name: "Tag2", Color: "#00ff00"},
	}

	tests := []struct {
		name    string
		listFn  func(ctx context.Context) ([]model.Tag, error)
		want    []model.Tag
		wantErr bool
	}{
		{
			name: "successful list",
			listFn: func(ctx context.Context) ([]model.Tag, error) {
				return tags, nil
			},
			want:    tags,
			wantErr: false,
		},
		{
			name: "repository error",
			listFn: func(ctx context.Context) ([]model.Tag, error) {
				return nil, errors.New("database error")
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repository.RepoMock{
				ListTagFn: tt.listFn,
			}

			s := service.NewService(repo)
			got, gotErr := s.ListTags(context.Background())
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ListTags() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ListTags() succeeded unexpectedly")
			}
			if len(got) != len(tt.want) {
				t.Errorf("ListTags() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i].Id != tt.want[i].Id || got[i].Name != tt.want[i].Name || got[i].Color != tt.want[i].Color {
					t.Errorf("ListTags()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestTagService_CreateTag(t *testing.T) {
	tests := []struct {
		name     string
		tag      model.Tag
		createFn func(ctx context.Context, tag model.Tag) (model.Tag, error)
		want     model.Tag
		wantErr  bool
	}{
		{
			name: "successful create",
			tag:  model.Tag{Name: "New Tag", Color: "#123456"},
			createFn: func(ctx context.Context, tag model.Tag) (model.Tag, error) {
				tag.Id = uuid.New()
				return tag, nil
			},
			want:    model.Tag{Name: "New Tag", Color: "#123456"},
			wantErr: false,
		},
		{
			name: "repository error",
			tag:  model.Tag{Name: "New Tag", Color: "#123456"},
			createFn: func(ctx context.Context, tag model.Tag) (model.Tag, error) {
				return model.Tag{}, errors.New("database error")
			},
			want:    model.Tag{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repository.RepoMock{
				CreateTagFn: tt.createFn,
			}
			s := service.NewService(repo)
			got, gotErr := s.CreateTag(context.Background(), tt.tag)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateTag() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateTag() succeeded unexpectedly")
			}
			if got.Name != tt.want.Name || got.Color != tt.want.Color || got.Id == tt.tag.Id {
				t.Errorf("CreateTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTagService_UpdateTag(t *testing.T) {
	tagId := uuid.New()
	tests := []struct {
		name     string
		tag      model.Tag
		updateFn func(ctx context.Context, tag model.Tag) (model.Tag, error)
		want     model.Tag
		wantErr  bool
	}{
		{
			name: "successful update",
			tag:  model.Tag{Id: tagId, Name: "Updated Tag", Color: "#654321"},
			updateFn: func(ctx context.Context, tag model.Tag) (model.Tag, error) {
				return tag, nil
			},
			want:    model.Tag{Id: tagId, Name: "Updated Tag", Color: "#654321"},
			wantErr: false,
		},
		{
			name: "repository error",
			tag:  model.Tag{Id: tagId, Name: "Updated Tag", Color: "#654321"},
			updateFn: func(ctx context.Context, tag model.Tag) (model.Tag, error) {
				return model.Tag{}, errors.New("database error")
			},
			want:    model.Tag{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repository.RepoMock{
				UpdateTagFn: tt.updateFn,
			}
			s := service.NewService(repo)
			got, gotErr := s.UpdateTag(context.Background(), tt.tag)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UpdateTag() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UpdateTag() succeeded unexpectedly")
			}
			if got.Id != tt.want.Id || got.Name != tt.want.Name || got.Color != tt.want.Color {
				t.Errorf("UpdateTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTagService_DeleteTag(t *testing.T) {
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
				DeleteTagFn: tt.deleteFn,
			}
			s := service.NewService(repo)
			gotErr := s.DeleteTag(context.Background(), uuid.New())
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
