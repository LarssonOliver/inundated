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

type mockProjectService struct {
	CreateFn func(ctx context.Context, project model.Project) (model.Project, error)
	DeleteFn func(ctx context.Context, id uuid.UUID) error
	GetFn    func(ctx context.Context, id uuid.UUID) (model.Project, error)
	ListFn   func(ctx context.Context) ([]model.Project, error)
	UpdateFn func(ctx context.Context, project model.Project) (model.Project, error)
}

var _ service.ProjectService = (*mockProjectService)(nil)

// CreateProject implements [service.ProjectService].
func (m *mockProjectService) CreateProject(ctx context.Context, project model.Project) (model.Project, error) {
	return m.CreateFn(ctx, project)
}

// DeleteProject implements [service.ProjectService].
func (m *mockProjectService) DeleteProject(ctx context.Context, id uuid.UUID) error {
	return m.DeleteFn(ctx, id)
}

// GetProject implements [service.ProjectService].
func (m *mockProjectService) GetProject(ctx context.Context, id uuid.UUID) (model.Project, error) {
	return m.GetFn(ctx, id)
}

// ListProjects implements [service.ProjectService].
func (m *mockProjectService) ListProjects(ctx context.Context) ([]model.Project, error) {
	return m.ListFn(ctx)
}

// UpdateProject implements [service.ProjectService].
func (m *mockProjectService) UpdateProject(ctx context.Context, project model.Project) (model.Project, error) {
	return m.UpdateFn(ctx, project)
}

func TestProjectHandler_CreateProject(t *testing.T) {
	tests := []struct {
		name     string
		createFn func(ctx context.Context, project model.Project) (model.Project, error)
		request  api.CreateProject
		want     api.Project
		wantErr  bool
	}{
		{
			name: "successful create",
			createFn: func(ctx context.Context, project model.Project) (model.Project, error) {
				project.Id = uuid.New()
				return project, nil
			},
			request: api.CreateProject{Name: "New Project", Color: "#123456"},
			want:    api.Project{Name: "New Project", Color: "#123456"},
			wantErr: false,
		},
		{
			name: "service error",
			createFn: func(ctx context.Context, project model.Project) (model.Project, error) {
				return model.Project{}, errors.New("service error")
			},
			request: api.CreateProject{Name: "Error Project", Color: "#654321"},
			want:    api.Project{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockProjectService{
				CreateFn: tt.createFn,
			}

			ta := handlers.NewProjectHandler(svc)

			request := api.CreateProjectRequestObject{
				Body: &tt.request,
			}

			raw, gotErr := ta.CreateProject(context.Background(), request)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateProject() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateProject() succeeded unexpectedly")
			}

			got := raw.(api.CreateProject201JSONResponse)

			if tt.want.Name != got.Name || tt.want.Color != got.Color {
				t.Errorf("CreateProject() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProjectHandler_DeleteProject(t *testing.T) {
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
			svc := &mockProjectService{
				DeleteFn: tt.deleteFn,
			}

			request := api.DeleteProjectRequestObject{
				ProjectId: tt.request,
			}

			ta := handlers.NewProjectHandler(svc)
			_, gotErr := ta.DeleteProject(context.Background(), request)

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

func TestProjectHandler_GetProject(t *testing.T) {
	tests := []struct {
		name    string
		getFn   func(ctx context.Context, id uuid.UUID) (model.Project, error)
		request uuid.UUID
		want    api.Project
		wantErr bool
	}{
		{
			name: "successful get",
			getFn: func(ctx context.Context, id uuid.UUID) (model.Project, error) {
				return model.Project{Id: id, Name: "Sample Project", Color: "#abcdef"}, nil
			},
			request: uuid.New(),
			want:    api.Project{Name: "Sample Project", Color: "#abcdef"},
			wantErr: false,
		},
		{
			name: "service error",
			getFn: func(ctx context.Context, id uuid.UUID) (model.Project, error) {
				return model.Project{}, errors.New("service error")
			},
			request: uuid.New(),
			want:    api.Project{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockProjectService{
				GetFn: tt.getFn,
			}

			ta := handlers.NewProjectHandler(svc)

			request := api.GetProjectRequestObject{
				ProjectId: tt.request,
			}

			got, gotErr := ta.GetProject(context.Background(), request)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetProject() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetProject() succeeded unexpectedly")
			}

			res := got.(api.GetProject200JSONResponse)
			if res.Id == uuid.Nil || res.Name != tt.want.Name || res.Color != tt.want.Color {
				t.Errorf("GetProject() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProjectHandler_ListProjects(t *testing.T) {
	project1 := model.Project{
		Id:    uuid.New(),
		Name:  "backend",
		Color: "#ff0000",
	}
	project2 := model.Project{
		Id:    uuid.New(),
		Name:  "frontend",
		Color: "#00ff00",
	}

	tests := []struct {
		name    string
		listFn  func(ctx context.Context) ([]model.Project, error)
		want    []api.Project
		wantErr bool
	}{
		{
			name: "success with multiple projects",
			listFn: func(ctx context.Context) ([]model.Project, error) {
				return []model.Project{project1, project2}, nil
			},
			want: []api.Project{
				{
					Id:    project1.Id,
					Name:  project1.Name,
					Color: project1.Color,
				},
				{
					Id:    project2.Id,
					Name:  project2.Name,
					Color: project2.Color,
				},
			},
			wantErr: false,
		},
		{
			name: "success with empty list",
			listFn: func(ctx context.Context) ([]model.Project, error) {
				return []model.Project{}, nil
			},
			want:    []api.Project{},
			wantErr: false,
		},
		{
			name: "service returns error",
			listFn: func(ctx context.Context) ([]model.Project, error) {
				return nil, errors.New("database unavailable")
			},
			want:    nil,
			wantErr: true,
		},
		// {
		// 	name: "context cancelled",
		// 	listFn: func(ctx context.Context) ([]model.Project, error) {
		// 		<-ctx.Done()
		// 		return nil, ctx.Err()
		// 	},
		// 	want:    nil,
		// 	wantErr: true,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockProjectService{
				ListFn: tt.listFn,
			}
			ta := handlers.NewProjectHandler(svc)
			request := api.ListProjectsRequestObject{}
			got, gotErr := ta.ListProjects(context.Background(), request)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ListProjects() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ListProjects() succeeded unexpectedly")
			}
			res := got.(api.ListProjects200JSONResponse)
			if len(res) != len(tt.want) {
				t.Errorf("ListProjects() = %v, want %v", got, tt.want)
				return
			}
			for i, project := range res {
				if project.Id == uuid.Nil || project.Name != tt.want[i].Name || project.Color != tt.want[i].Color {
					t.Errorf("ListProjects() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestProjectHandler_UpdateProject(t *testing.T) {
	existingID := uuid.New()
	name := "existing-project"
	color := "#ffffff"
	newName := "updated-project"

	tests := []struct {
		name      string
		getFn     func(ctx context.Context, id uuid.UUID) (model.Project, error)
		updateFn  func(ctx context.Context, project model.Project) (model.Project, error)
		requestId uuid.UUID
		request   api.UpdateProject
		want      api.Project
		wantErr   bool
	}{
		{
			name:      "successfully updates project",
			requestId: existingID,
			request: api.UpdateProject{
				Name:  &name,
				Color: &color,
			},
			getFn: func(ctx context.Context, id uuid.UUID) (model.Project, error) {
				return model.Project{Id: id, Name: "old-name", Color: "#000000"}, nil
			},
			updateFn: func(ctx context.Context, project model.Project) (model.Project, error) {
				return project, nil
			},
			want: api.Project{
				Id:    existingID,
				Name:  name,
				Color: color,
			},
			wantErr: false,
		},
		{
			name:      "service returns generic error",
			requestId: existingID,
			request: api.UpdateProject{
				Name:  &name,
				Color: &color,
			},
			getFn: func(ctx context.Context, id uuid.UUID) (model.Project, error) {
				return model.Project{Id: id, Name: "old-name", Color: "#000000"}, nil
			},
			updateFn: func(ctx context.Context, project model.Project) (model.Project, error) {
				return model.Project{}, errors.New("database down")
			},
			wantErr: true,
		},
		{
			name:      "update only name",
			requestId: existingID,
			request: api.UpdateProject{
				Name: &newName,
			},
			getFn: func(ctx context.Context, id uuid.UUID) (model.Project, error) {
				return model.Project{Id: id, Name: "old-name", Color: "#000000"}, nil
			},
			updateFn: func(ctx context.Context, project model.Project) (model.Project, error) {
				return project, nil
			},
			want: api.Project{
				Id:    existingID,
				Name:  newName,
				Color: "#000000",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockProjectService{
				UpdateFn: tt.updateFn,
				GetFn:    tt.getFn,
			}
			ta := handlers.NewProjectHandler(svc)
			request := api.UpdateProjectRequestObject{
				ProjectId: tt.requestId,
				Body:  &tt.request,
			}

			got, gotErr := ta.UpdateProject(context.Background(), request)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UpdateProject() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UpdateProject() succeeded unexpectedly")
			}
			res := got.(api.UpdateProject200JSONResponse)
			if res.Id == uuid.Nil || res.Name != tt.want.Name || res.Color != tt.want.Color {
				t.Errorf("UpdateProject() = %v, want %v", got, tt.want)
			}
		})
	}
}
