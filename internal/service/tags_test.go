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

func durPtr(d time.Duration) *time.Duration {
	return &d
}

func TestTagService_GetTag(t *testing.T) {
	testId := uuid.New()

	tests := []struct {
		name        string
		id          uuid.UUID
		getFn       func(ctx context.Context, id uuid.UUID) (model.Tag, error)
		includes    *service.TagServiceGetIncludes
		totalTimeFn func(ctx context.Context, ids []uuid.UUID) (time.Duration, error)
		want        model.Tag
		wantErr     bool
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
		{
			name: "Include total time",
			id:   testId,
			getFn: func(ctx context.Context, id uuid.UUID) (model.Tag, error) {
				return model.Tag{Id: id, Name: "Test Tag", Color: "#abcdef"}, nil
			},
			includes: &service.TagServiceGetIncludes{TotalTime: true},
			totalTimeFn: func(ctx context.Context, ids []uuid.UUID) (time.Duration, error) {
				return 2 * time.Hour, nil
			},
			want: model.Tag{Id: testId, Name: "Test Tag", Color: "#abcdef", TotalTime: durPtr(2 * time.Hour)},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repository.RepoMock{
				GetTagFn: tt.getFn,
			}
			if tt.totalTimeFn != nil {
				repo.GetTotalDurationByTagsFn = tt.totalTimeFn
			}

			s := service.NewService(repo)
			got, gotErr := s.GetTag(context.Background(), tt.id, tt.includes)
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			require.NotEqual(t, uuid.Nil, got.Id)
			require.Equal(t, tt.want.Name, got.Name)
			require.Equal(t, tt.want.Color, got.Color)

			if tt.includes != nil && tt.includes.TotalTime {
				require.NotNil(t, got.TotalTime)
				require.Equal(t, *tt.want.TotalTime, *got.TotalTime)
			} else {
				require.Nil(t, got.TotalTime)
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
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			require.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestTagService_CreateTag(t *testing.T) {
	id := uuid.New()
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
			name: "ensure new ID is generated",
			tag:  model.Tag{Id: id, Name: "New Tag", Color: "#123456"},
			createFn: func(ctx context.Context, tag model.Tag) (model.Tag, error) {
				return tag, nil
			},
			want:    model.Tag{Id: id, Name: "New Tag", Color: "#123456"},
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
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			require.NotEqual(t, uuid.Nil, got.Id)
			require.NotEqual(t, tt.want.Id, got.Id)
			require.Equal(t, tt.want.Name, got.Name)
			require.Equal(t, tt.want.Color, got.Color)
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
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			require.NotEqual(t, uuid.Nil, got.Id)
			require.Equal(t, tt.want.Id, got.Id)
			require.Equal(t, tt.want.Name, got.Name)
			require.Equal(t, tt.want.Color, got.Color)
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
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}
