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
			got, gotErr := s.GetProject(context.Background(), tt.id)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetProject() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetProject() succeeded unexpectedly")
			}
			if got.Id != tt.want.Id || got.Name != tt.want.Name || got.Color != tt.want.Color {
				t.Errorf("GetProject() = %v, want %v", got, tt.want)
			}
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
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ListProjects() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ListProjects() succeeded unexpectedly")
			}
			if len(got) != len(tt.want) {
				t.Errorf("ListProjects() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i].Id != tt.want[i].Id || got[i].Name != tt.want[i].Name || got[i].Color != tt.want[i].Color {
					t.Errorf("ListProjects()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestProjectService_CreateProject(t *testing.T) {
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
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateProject() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateProject() succeeded unexpectedly")
			}
			if got.Name != tt.want.Name || got.Color != tt.want.Color || got.Id == tt.project.Id {
				t.Errorf("CreateProject() = %v, want %v", got, tt.want)
			}
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
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UpdateProject() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UpdateProject() succeeded unexpectedly")
			}
			if got.Id != tt.want.Id || got.Name != tt.want.Name || got.Color != tt.want.Color {
				t.Errorf("UpdateProject() = %v, want %v", got, tt.want)
			}
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
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("DeleteProject() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("DeleteProject() succeeded unexpectedly")
			}
		})
	}
}
