package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/service"
	"github.com/stretchr/testify/require"
)

func TestProjectService_GetProject(t *testing.T) {
	testId := uuid.New()

	tests := []struct {
		name    string
		id      uuid.UUID
		getFn   func(ctx context.Context, id uuid.UUID) (model.Project, error)
		want    model.Project
		wantErr bool
	}{
		{
			name: "successful get",
			id:   testId,
			getFn: func(ctx context.Context, id uuid.UUID) (model.Project, error) {
				return model.Project{Id: id, Name: "Test Project", Color: "#abcdef"}, nil
			},
			want:    model.Project{Id: testId, Name: "Test Project", Color: "#abcdef"},
			wantErr: false,
		},
		{
			name: "repository error",
			id:   testId,
			getFn: func(ctx context.Context, id uuid.UUID) (model.Project, error) {
				return model.Project{}, errors.New("not found")
			},
			want:    model.Project{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repository.RepoMock{
				GetProjectFn: tt.getFn,
			}

			s := service.NewService(repo)
			got, gotErr := s.GetProject(context.Background(), tt.id, nil)
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

func TestProjectService_ListProjects(t *testing.T) {
	projects := []model.Project{
		{Id: uuid.New(), Name: "Project1", Color: "#ff0000"},
		{Id: uuid.New(), Name: "Project2", Color: "#00ff00"},
	}

	tests := []struct {
		name    string
		listFn  func(ctx context.Context) ([]model.Project, error)
		want    []model.Project
		wantErr bool
	}{
		{
			name: "successful list",
			listFn: func(ctx context.Context) ([]model.Project, error) {
				return projects, nil
			},
			want:    projects,
			wantErr: false,
		},
		{
			name: "repository error",
			listFn: func(ctx context.Context) ([]model.Project, error) {
				return nil, errors.New("database error")
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repository.RepoMock{
				ListProjectFn: tt.listFn,
			}

			s := service.NewService(repo)
			got, gotErr := s.ListProjects(context.Background())
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			require.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestProjectService_CreateProject(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		name     string
		project  model.Project
		createFn func(ctx context.Context, project model.Project) (model.Project, error)
		want     model.Project
		wantErr  bool
	}{
		{
			name:    "successful create",
			project: model.Project{Name: "New Project", Color: "#123456"},
			createFn: func(ctx context.Context, project model.Project) (model.Project, error) {
				project.Id = uuid.New()
				return project, nil
			},
			want:    model.Project{Name: "New Project", Color: "#123456"},
			wantErr: false,
		},
		{
			name:    "ensure create generates Id",
			project: model.Project{Id: id, Name: "New Project", Color: "#123456"},
			createFn: func(ctx context.Context, project model.Project) (model.Project, error) {
				return project, nil
			},
			want:    model.Project{Id: id, Name: "New Project", Color: "#123456"},
			wantErr: false,
		},
		{
			name:    "repository error",
			project: model.Project{Name: "New Project", Color: "#123456"},
			createFn: func(ctx context.Context, project model.Project) (model.Project, error) {
				return model.Project{}, errors.New("database error")
			},
			want:    model.Project{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repository.RepoMock{
				CreateProjectFn: tt.createFn,
			}
			s := service.NewService(repo)
			got, gotErr := s.CreateProject(context.Background(), tt.project)
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

func TestProjectService_UpdateProject(t *testing.T) {
	projectId := uuid.New()
	tests := []struct {
		name     string
		project  model.Project
		updateFn func(ctx context.Context, project model.Project) (model.Project, error)
		want     model.Project
		wantErr  bool
	}{
		{
			name:    "successful update",
			project: model.Project{Id: projectId, Name: "Updated Project", Color: "#654321"},
			updateFn: func(ctx context.Context, project model.Project) (model.Project, error) {
				return project, nil
			},
			want:    model.Project{Id: projectId, Name: "Updated Project", Color: "#654321"},
			wantErr: false,
		},
		{
			name:    "repository error",
			project: model.Project{Id: projectId, Name: "Updated Project", Color: "#654321"},
			updateFn: func(ctx context.Context, project model.Project) (model.Project, error) {
				return model.Project{}, errors.New("database error")
			},
			want:    model.Project{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repository.RepoMock{
				UpdateProjectFn: tt.updateFn,
			}
			s := service.NewService(repo)
			got, gotErr := s.UpdateProject(context.Background(), tt.project)
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

func TestProjectService_DeleteProject(t *testing.T) {
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
				DeleteProjectFn: tt.deleteFn,
			}
			s := service.NewService(repo)
			gotErr := s.DeleteProject(context.Background(), uuid.New())
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}
