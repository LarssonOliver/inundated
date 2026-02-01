package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository/memory"
)

func ptrd(d time.Duration) *time.Duration {
	return &d
}

func TestProjectStore_CreateProject(t *testing.T) {
	tagIds := []uuid.UUID{uuid.New(), uuid.New()}

	tests := []struct {
		name     string
		project  model.Project
		want     model.Project
		wantErr  bool
		errType  error
		getTagFn func(ctx context.Context, id uuid.UUID) (model.Tag, error)
	}{
		{
			name:    "Test CreateProject with valid input",
			project: model.Project{Name: "Urgent", Color: "#FF0000", TimeBudget: ptrd(0), TagIds: tagIds},
			want:    model.Project{Name: "Urgent", Color: "#FF0000", TimeBudget: ptrd(0), TagIds: tagIds},
			wantErr: false,
			getTagFn: func(ctx context.Context, id uuid.UUID) (model.Tag, error) {
				return model.Tag{Id: id, Name: "Tag", Color: "#FFFFFF"}, nil
			},
		},
		{
			name:    "Test CreateProject with another valid input",
			project: model.Project{Name: "Optional", Color: "#00FF00", TimeBudget: ptrd(4 * time.Hour), TagIds: nil},
			want:    model.Project{Name: "Optional", Color: "#00FF00", TimeBudget: ptrd(4 * time.Hour), TagIds: []uuid.UUID{}},
			wantErr: false,
			getTagFn: func(ctx context.Context, id uuid.UUID) (model.Tag, error) {
				return model.Tag{Id: id, Name: "Tag", Color: "#FFFFFF"}, nil
			},
		},
		{
			name:    "Test CreateProject with empty name",
			project: model.Project{Name: "", Color: "#0000FF"},
			want:    model.Project{},
			wantErr: true,
			errType: model.ErrInvalidArgument,
			getTagFn: func(ctx context.Context, id uuid.UUID) (model.Tag, error) {
				return model.Tag{Id: id, Name: "Tag", Color: "#FFFFFF"}, nil
			},
		},
		{
			name:    "Test CreateProject with invalid color",
			project: model.Project{Name: "InvalidColor", Color: "NotAColor"},
			want:    model.Project{},
			wantErr: true,
			errType: model.ErrInvalidArgument,
			getTagFn: func(ctx context.Context, id uuid.UUID) (model.Tag, error) {
				return model.Tag{Id: id, Name: "Tag", Color: "#FFFFFF"}, nil
			},
		},
		{
			name:    "Test CreateProject with empty color",
			project: model.Project{Name: "NoColor", Color: ""},
			want:    model.Project{},
			wantErr: true,
			errType: model.ErrInvalidArgument,
			getTagFn: func(ctx context.Context, id uuid.UUID) (model.Tag, error) {
				return model.Tag{Id: id, Name: "Tag", Color: "#FFFFFF"}, nil
			},
		},
		{
			name:    "Test CreateProject with nil project",
			project: model.Project{},
			want:    model.Project{},
			wantErr: true,
			errType: model.ErrInvalidArgument,
			getTagFn: func(ctx context.Context, id uuid.UUID) (model.Tag, error) {
				return model.Tag{Id: id, Name: "Tag", Color: "#FFFFFF"}, nil
			},
		},
		{
			name:    "Test CreateProject with set ID (should be ignored)",
			project: model.Project{Id: uuid.New(), Name: "WithID", Color: "#123456"},
			want:    model.Project{Name: "WithID", Color: "#123456"},
			wantErr: false,
			getTagFn: func(ctx context.Context, id uuid.UUID) (model.Tag, error) {
				return model.Tag{Id: id, Name: "Tag", Color: "#FFFFFF"}, nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags := tagRepoMock{
				GetFn: tt.getTagFn,
			}
			ta := memory.NewProjectStore(&tags)
			got, gotErr := ta.CreateProject(context.Background(), tt.project)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateProject() failed: %v", gotErr)
				}
				if tt.errType != nil && gotErr != tt.errType {
					t.Errorf("CreateProject() error = %v, wantErrType %v", gotErr, tt.errType)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateProject() succeeded unexpectedly")
			}
			if got.Name != tt.want.Name || got.Color != tt.want.Color || got.Id == tt.want.Id || len(got.TagIds) != len(tt.want.TagIds) {
				t.Errorf("CreateProject() = %v, want %v", got, tt.want)
				return
			}
			if (tt.want.TimeBudget == nil && got.TimeBudget != nil) || (tt.want.TimeBudget != nil && got.TimeBudget == nil) {
				t.Errorf("CreateProject() TimeBudget = %v, want %v", got.TimeBudget, tt.want.TimeBudget)
				return
			}
			if tt.want.TimeBudget != nil && got.TimeBudget != nil && *got.TimeBudget != *tt.want.TimeBudget {
				t.Errorf("CreateProject() TimeBudget = %v, want %v", got.TimeBudget, tt.want.TimeBudget)
				return
			}
			for i, tagId := range tt.want.TagIds {
				if got.TagIds[i] != tagId {
					t.Errorf("CreateProject() TagIds = %v, want %v", got.TagIds, tt.want.TagIds)
				}
			}

		})
	}
}

func TestProjectStore_GetProject(t *testing.T) {
	tagIds := []uuid.UUID{uuid.New(), uuid.New()}
	tests := []struct {
		name          string
		createProject model.Project
		getId         func(createdProject *model.Project) uuid.UUID
		want          model.Project
		wantErr       bool
		errType       error
	}{
		{
			name:          "Test GetProject with existing ID",
			createProject: model.Project{Name: "Project1", Color: "#FF0000", TimeBudget: ptrd(2 * time.Hour), TagIds: tagIds},
			getId: func(createdProject *model.Project) uuid.UUID {
				return createdProject.Id
			},
			want:    model.Project{Name: "Project1", Color: "#FF0000", TimeBudget: ptrd(2 * time.Hour), TagIds: tagIds},
			wantErr: false,
		},
		{
			name:          "Test GetProject with non-existing ID",
			createProject: model.Project{Name: "Project2", Color: "#00FF00"},
			getId: func(createdProject *model.Project) uuid.UUID {
				return uuid.New()
			},
			want:    model.Project{},
			wantErr: true,
			errType: model.ErrNotFound,
		},
		{
			name:          "Test GetProject with empty UUID",
			createProject: model.Project{Name: "Project3", Color: "#0000FF"},
			getId: func(createdProject *model.Project) uuid.UUID {
				return uuid.Nil
			},
			want:    model.Project{},
			wantErr: true,
			errType: model.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := memory.NewProjectStore(&tagRepoMock{})
			project, _ := ta.CreateProject(context.Background(), tt.createProject)
			getId := tt.getId(&project)

			got, gotErr := ta.GetProject(context.Background(), getId)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetProject() failed: %v", gotErr)
				}
				if tt.errType != nil && gotErr != tt.errType {
					t.Errorf("CreateProject() error = %v, wantErrType %v", gotErr, tt.errType)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetProject() succeeded unexpectedly")
			}
			if project.Name != tt.want.Name || project.Color != tt.want.Color || project.Id != got.Id || len(got.TagIds) != len(tt.want.TagIds) {
				t.Errorf("GetProject() = %v, want %v", got, tt.want)
				return
			}
			if (tt.want.TimeBudget == nil && got.TimeBudget != nil) || (tt.want.TimeBudget != nil && got.TimeBudget == nil) {
				t.Errorf("CreateProject() TimeBudget = %v, want %v", got.TimeBudget, tt.want.TimeBudget)
				return
			}
			if tt.want.TimeBudget != nil && got.TimeBudget != nil && *got.TimeBudget != *tt.want.TimeBudget {
				t.Errorf("CreateProject() TimeBudget = %v, want %v", got.TimeBudget, tt.want.TimeBudget)
				return
			}
			for i, tagId := range tt.want.TagIds {
				if got.TagIds[i] != tagId {
					t.Errorf("GetProject() TagIds = %v, want %v", got.TagIds, tt.want.TagIds)
				}
			}
		})
	}
}

func TestProjectStore_ListProjects(t *testing.T) {
	tagIds := []uuid.UUID{uuid.New(), uuid.New()}

	tests := []struct {
		name           string // description of this test case
		insertProjects []model.Project
		wantErr        bool
		errType        error
	}{
		{
			name: "Test ListProjects with multiple projects",
			insertProjects: []model.Project{
				{Name: "Project1", Color: "#FF0000", TimeBudget: ptrd(3 * time.Hour), TagIds: tagIds},
				{Name: "Project2", Color: "#00FF00"},
				{Name: "Project3", Color: "#0000FF"},
			},
			wantErr: false,
		},
		{
			name:           "Test ListProjects with no projects",
			insertProjects: []model.Project{},
			wantErr:        false,
		},
		{
			name: "Test ListProjects with one project",
			insertProjects: []model.Project{
				{Name: "OnlyProject", Color: "#123456"},
			},
			wantErr: false,
		},
		{
			name:           "Test ListProjects with duplicate projects",
			insertProjects: []model.Project{{Name: "DupProject", Color: "#654321"}, {Name: "DupProject", Color: "#654321"}},
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := memory.NewProjectStore(&tagRepoMock{})

			for i, project := range tt.insertProjects {
				createdProject, _ := ta.CreateProject(context.Background(), project)
				tt.insertProjects[i].Id = createdProject.Id
			}

			got, gotErr := ta.ListProjects(context.Background())
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ListProjects() failed: %v", gotErr)
				}
				if tt.errType != nil && gotErr != tt.errType {
					t.Errorf("CreateProject() error = %v, wantErrType %v", gotErr, tt.errType)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ListProjects() succeeded unexpectedly")
			}

			if len(got) != len(tt.insertProjects) {
				t.Errorf("ListProjects() = %v, want %v", got, tt.insertProjects)
			}

			for _, project := range tt.insertProjects {
				found := false
			outer:
				for _, gotProject := range got {
					if gotProject.Id == project.Id && gotProject.Name == project.Name && gotProject.Color == project.Color && len(gotProject.TagIds) == len(project.TagIds) {
						for i, tagId := range project.TagIds {
							if gotProject.TagIds[i] != tagId {
								continue outer
							}
						}
						found = true
						break
					}
				}
				if !found {
					t.Errorf("ListProjects() missing expected project: %v", project)
				}
			}
		})
	}
}

func TestProjectStore_UpdateProject(t *testing.T) {
	tagIds := []uuid.UUID{uuid.New(), uuid.New()}

	tests := []struct {
		name          string
		project       model.Project
		editProject   model.Project
		editProjectId func(createdProject *model.Project) uuid.UUID
		want          model.Project
		wantErr       bool
		errType       error
		getTagFn      func(ctx context.Context, id uuid.UUID) (model.Tag, error)
	}{
		{
			name:        "Test UpdateProject with valid input",
			project:     model.Project{Name: "OldName", Color: "#FF0000", TimeBudget: ptrd(1 * time.Hour), TagIds: []uuid.UUID{tagIds[0]}},
			editProject: model.Project{Name: "NewName", Color: "#00FF00", TimeBudget: ptrd(2 * time.Hour), TagIds: tagIds},
			editProjectId: func(createdProject *model.Project) uuid.UUID {
				return createdProject.Id
			},
			want:    model.Project{Name: "NewName", Color: "#00FF00", TimeBudget: ptrd(2 * time.Hour), TagIds: tagIds},
			wantErr: false,
			getTagFn: func(ctx context.Context, id uuid.UUID) (model.Tag, error) {
				return model.Tag{Id: id, Name: "Tag", Color: "#FFFFFF"}, nil
			},
		},
		{
			name:        "Test UpdateProject with non-existing ID",
			project:     model.Project{Name: "Project1", Color: "#FF0000"},
			editProject: model.Project{Name: "ShouldFail", Color: "#0000FF"},
			editProjectId: func(createdProject *model.Project) uuid.UUID {
				return uuid.New()
			},
			want:    model.Project{},
			wantErr: true,
			errType: model.ErrNotFound,
			getTagFn: func(ctx context.Context, id uuid.UUID) (model.Tag, error) {
				return model.Tag{Id: id, Name: "Tag", Color: "#FFFFFF"}, nil
			},
		},
		{
			name:        "Test UpdateProject with empty name",
			project:     model.Project{Name: "Project2", Color: "#00FF00"},
			editProject: model.Project{Name: "", Color: "#0000FF"},
			editProjectId: func(createdProject *model.Project) uuid.UUID {
				return createdProject.Id
			},
			want:    model.Project{},
			wantErr: true,
			errType: model.ErrInvalidArgument,
			getTagFn: func(ctx context.Context, id uuid.UUID) (model.Tag, error) {
				return model.Tag{Id: id, Name: "Tag", Color: "#FFFFFF"}, nil
			},
		},
		{
			name:        "Test UpdateProject with invalid color",
			project:     model.Project{Name: "Project3", Color: "#0000FF"},
			editProject: model.Project{Name: "Project3", Color: "InvalidColor"},
			editProjectId: func(createdProject *model.Project) uuid.UUID {
				return createdProject.Id
			},
			want:    model.Project{},
			wantErr: true,
			errType: model.ErrInvalidArgument,
			getTagFn: func(ctx context.Context, id uuid.UUID) (model.Tag, error) {
				return model.Tag{Id: id, Name: "Tag", Color: "#FFFFFF"}, nil
			},
		},
		{
			name:        "Test UpdateProject with empty ID",
			project:     model.Project{Name: "Project4", Color: "#123456"},
			editProject: model.Project{Name: "Project4Updated", Color: "#654321"},
			editProjectId: func(createdProject *model.Project) uuid.UUID {
				return uuid.Nil
			},
			want:    model.Project{},
			wantErr: true,
			errType: model.ErrNotFound,
			getTagFn: func(ctx context.Context, id uuid.UUID) (model.Tag, error) {
				return model.Tag{Id: id, Name: "Tag", Color: "#FFFFFF"}, nil
			},
		},
		{
			name:        "Test UpdateProject with same name and color",
			project:     model.Project{Name: "Project5", Color: "#ABCDEF"},
			editProject: model.Project{Name: "Project5", Color: "#ABCDEF"},
			editProjectId: func(createdProject *model.Project) uuid.UUID {
				return createdProject.Id
			},
			want:    model.Project{Name: "Project5", Color: "#ABCDEF"},
			wantErr: false,
			getTagFn: func(ctx context.Context, id uuid.UUID) (model.Tag, error) {
				return model.Tag{Id: id, Name: "Tag", Color: "#FFFFFF"}, nil
			},
		},
		{
			name:        "Test UpdateProject with nil TagIds (should become empty slice)",
			project:     model.Project{Name: "Project6", Color: "#FEDCBA", TagIds: tagIds},
			editProject: model.Project{Name: "Project6Updated", Color: "#ABC123", TagIds: nil},
			editProjectId: func(createdProject *model.Project) uuid.UUID {
				return createdProject.Id
			},
			want:    model.Project{Name: "Project6Updated", Color: "#ABC123", TagIds: []uuid.UUID{}},
			wantErr: false,
			getTagFn: func(ctx context.Context, id uuid.UUID) (model.Tag, error) {
				return model.Tag{Id: id, Name: "Tag", Color: "#FFFFFF"}, nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := memory.NewProjectStore(&tagRepoMock{GetFn: tt.getTagFn})

			insertedProject, _ := ta.CreateProject(context.Background(), tt.project)
			editId := tt.editProjectId(&insertedProject)

			tt.editProject.Id = editId
			tt.want.Id = editId

			got, gotErr := ta.UpdateProject(context.Background(), tt.editProject)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UpdateProject() failed: %v", gotErr)
				}
				if tt.errType != nil && gotErr != tt.errType {
					t.Errorf("CreateProject() error = %v, wantErrType %v", gotErr, tt.errType)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UpdateProject() succeeded unexpectedly")
			}
			if tt.want.Name != got.Name || tt.want.Color != got.Color || tt.want.Id != got.Id || len(got.TagIds) != len(tt.want.TagIds) {
				t.Errorf("UpdateProject() = %v, want %v", got, tt.want)
				return
			}
			if (tt.want.TimeBudget == nil && got.TimeBudget != nil) || (tt.want.TimeBudget != nil && got.TimeBudget == nil) {
				t.Errorf("CreateProject() TimeBudget = %v, want %v", got.TimeBudget, tt.want.TimeBudget)
				return
			}
			if tt.want.TimeBudget != nil && got.TimeBudget != nil && *got.TimeBudget != *tt.want.TimeBudget {
				t.Errorf("CreateProject() TimeBudget = %v, want %v", got.TimeBudget, tt.want.TimeBudget)
				return
			}
			for i, tagId := range tt.want.TagIds {
				if got.TagIds[i] != tagId {
					t.Errorf("UpdateProject() TagIds = %v, want %v", got.TagIds, tt.want.TagIds)
				}
			}
		})
	}
}

func TestProjectStore_DeleteProject(t *testing.T) {
	tests := []struct {
		name          string
		insertProject model.Project
		deleteId      func(createdProject *model.Project) uuid.UUID
		wantErr       bool
		errType       error
	}{
		{
			name:          "Test DeleteProject with existing ID",
			insertProject: model.Project{Name: "Project1", Color: "#FF0000"},
			deleteId: func(createdProject *model.Project) uuid.UUID {
				return createdProject.Id
			},
			wantErr: false,
		},
		{
			name:          "Test DeleteProject with non-existing ID",
			insertProject: model.Project{},
			deleteId: func(createdProject *model.Project) uuid.UUID {
				return uuid.New()
			},
			wantErr: true,
			errType: model.ErrNotFound,
		},
		{
			name:          "Test DeleteProject with empty UUID",
			insertProject: model.Project{Name: "Project3", Color: "#0000FF"},
			deleteId: func(createdProject *model.Project) uuid.UUID {
				return uuid.Nil
			},
			wantErr: true,
			errType: model.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := memory.NewProjectStore(&tagRepoMock{})

			project, _ := ta.CreateProject(context.Background(), tt.insertProject)
			deleteId := tt.deleteId(&project)

			gotErr := ta.DeleteProject(context.Background(), deleteId)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("DeleteProject() failed: %v", gotErr)
				}
				if tt.errType != nil && gotErr != tt.errType {
					t.Errorf("CreateProject() error = %v, wantErrType %v", gotErr, tt.errType)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("DeleteProject() succeeded unexpectedly")
			}
			project, err := ta.GetProject(context.Background(), deleteId)
			if err == nil {
				t.Errorf("Project with ID %v was not deleted, still exists: %v", deleteId, project)
			}
		})
	}
}
