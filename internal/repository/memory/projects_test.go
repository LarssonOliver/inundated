package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository/memory"
	"github.com/stretchr/testify/require"
)

func ptrd(d time.Duration) *time.Duration {
	return &d
}

func TestMemoryStore_Project_ScopeIsolation(t *testing.T) {
	ctx := context.Background()
	store := memory.NewMemoryStore()
	a := model.UserScope(uuid.New())
	b := model.UserScope(uuid.New())

	proj, err := store.CreateProject(ctx, a, model.Project{Name: "x", Color: "#123456"})
	require.NoError(t, err)

	_, err = store.GetProject(ctx, b, proj.Id)
	require.ErrorIs(t, err, model.ErrNotFound)

	page, err := store.ListProjects(ctx, b, model.DefaultPaginationParams())
	require.NoError(t, err)
	require.Empty(t, page.Data)
	require.Equal(t, 0, page.TotalCount)

	proj.Name = "hijack"
	_, err = store.UpdateProject(ctx, b, proj)
	require.ErrorIs(t, err, model.ErrNotFound)

	err = store.DeleteProject(ctx, b, proj.Id)
	require.ErrorIs(t, err, model.ErrNotFound)

	got, err := store.GetProject(ctx, a, proj.Id)
	require.NoError(t, err)
	require.Equal(t, "x", got.Name)
}

func TestProjectStore_CreateProject(t *testing.T) {
	tagIds := []uuid.UUID{uuid.New(), uuid.New()}

	tests := []struct {
		name    string
		project model.Project
		want    model.Project
		wantErr bool
		errType error
	}{
		{
			name:    "Test CreateProject with valid input",
			project: model.Project{Name: "Urgent", Color: "#FF0000", TimeBudget: ptrd(0), TagIds: tagIds},
			want:    model.Project{Name: "Urgent", Color: "#FF0000", TimeBudget: ptrd(0), TagIds: tagIds},
			wantErr: false,
		},
		{
			name:    "Test CreateProject with another valid input",
			project: model.Project{Name: "Optional", Color: "#00FF00", TimeBudget: ptrd(4 * time.Hour), TagIds: nil},
			want:    model.Project{Name: "Optional", Color: "#00FF00", TimeBudget: ptrd(4 * time.Hour), TagIds: []uuid.UUID{}},
			wantErr: false,
		},
		{
			name:    "Test CreateProject with empty name",
			project: model.Project{Name: "", Color: "#0000FF"},
			want:    model.Project{},
			wantErr: true,
			errType: model.ErrInvalidArgument,
		},
		{
			name:    "Test CreateProject with invalid color",
			project: model.Project{Name: "InvalidColor", Color: "NotAColor"},
			want:    model.Project{},
			wantErr: true,
			errType: model.ErrInvalidArgument,
		},
		{
			name:    "Test CreateProject with empty color",
			project: model.Project{Name: "NoColor", Color: ""},
			want:    model.Project{},
			wantErr: true,
			errType: model.ErrInvalidArgument,
		},
		{
			name:    "Test CreateProject with nil project",
			project: model.Project{},
			want:    model.Project{},
			wantErr: true,
			errType: model.ErrInvalidArgument,
		},
		{
			name:    "Test CreateProject with set ID (should be ignored)",
			project: model.Project{Id: uuid.New(), Name: "WithID", Color: "#123456"},
			want:    model.Project{Name: "WithID", Color: "#123456"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := memory.NewMemoryStore()
			for _, tagId := range tt.project.TagIds {
				_, _ = ta.CreateTag(context.Background(), testScope, model.Tag{Id: tagId, Name: "Tag", Color: "#FFFFFF"})
			}

			got, gotErr := ta.CreateProject(context.Background(), testScope, tt.project)
			if tt.wantErr {
				require.Error(t, gotErr)
				if tt.errType != nil {
					require.ErrorIs(t, gotErr, tt.errType)
				}
				return
			}

			require.NoError(t, gotErr)
			require.NotEqual(t, uuid.Nil, got.Id)
			require.Equal(t, tt.want.Name, got.Name)
			require.Equal(t, tt.want.Color, got.Color)
			require.ElementsMatch(t, tt.want.TagIds, got.TagIds)

			if tt.want.TimeBudget == nil {
				require.Nil(t, got.TimeBudget)
			} else {
				require.NotNil(t, got.TimeBudget)
				require.Equal(t, *tt.want.TimeBudget, *got.TimeBudget)
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
			ta := memory.NewMemoryStore()
			for _, tagId := range tt.createProject.TagIds {
				_, _ = ta.CreateTag(context.Background(), testScope, model.Tag{Id: tagId, Name: "Tag", Color: "#FFFFFF"})
			}
			project, _ := ta.CreateProject(context.Background(), testScope, tt.createProject)
			getId := tt.getId(&project)

			got, gotErr := ta.GetProject(context.Background(), testScope, getId)
			if tt.wantErr {
				require.Error(t, gotErr)
				if tt.errType != nil {
					require.ErrorIs(t, gotErr, tt.errType)
				}
				return
			}

			require.NoError(t, gotErr)
			require.NotEqual(t, uuid.Nil, got.Id)
			require.Equal(t, tt.want.Name, got.Name)
			require.Equal(t, tt.want.Color, got.Color)
			require.ElementsMatch(t, tt.want.TagIds, got.TagIds)

			if tt.want.TimeBudget == nil {
				require.Nil(t, got.TimeBudget)
			} else {
				require.NotNil(t, got.TimeBudget)
				require.Equal(t, *tt.want.TimeBudget, *got.TimeBudget)
			}
		})
	}
}

func TestProjectStore_ListProjects(t *testing.T) {
	tagIds := []uuid.UUID{uuid.New(), uuid.New()}

	tests := []struct {
		name           string
		insertProjects []model.Project
		params         model.PaginationParams
		wantLen        int
		wantTotal      int
		wantErr        bool
		errType        error
	}{
		{
			name: "Test ListProjects with multiple projects default pagination",
			insertProjects: []model.Project{
				{Name: "Project1", Color: "#FF0000", TimeBudget: ptrd(3 * time.Hour), TagIds: tagIds},
				{Name: "Project2", Color: "#00FF00"},
				{Name: "Project3", Color: "#0000FF"},
			},
			params:    model.DefaultPaginationParams(),
			wantLen:   3,
			wantTotal: 3,
		},
		{
			name:           "Test ListProjects with no projects",
			insertProjects: []model.Project{},
			params:         model.DefaultPaginationParams(),
			wantLen:        0,
			wantTotal:      0,
		},
		{
			name:           "Test ListProjects with one project",
			insertProjects: []model.Project{{Name: "OnlyProject", Color: "#123456"}},
			params:         model.DefaultPaginationParams(),
			wantLen:        1,
			wantTotal:      1,
		},
		{
			name:           "Test ListProjects with duplicate projects",
			insertProjects: []model.Project{{Name: "DupProject", Color: "#654321"}, {Name: "DupProject", Color: "#654321"}},
			params:         model.DefaultPaginationParams(),
			wantLen:        2,
			wantTotal:      2,
		},
		{
			name: "Test ListProjects with limit",
			insertProjects: []model.Project{
				{Name: "Project1", Color: "#FF0000"},
				{Name: "Project2", Color: "#00FF00"},
				{Name: "Project3", Color: "#0000FF"},
			},
			params:    model.PaginationParams{Limit: 2, Offset: 0},
			wantLen:   2,
			wantTotal: 3,
		},
		{
			name: "Test ListProjects with offset",
			insertProjects: []model.Project{
				{Name: "Project1", Color: "#FF0000"},
				{Name: "Project2", Color: "#00FF00"},
				{Name: "Project3", Color: "#0000FF"},
			},
			params:    model.PaginationParams{Limit: 10, Offset: 2},
			wantLen:   1,
			wantTotal: 3,
		},
		{
			name:           "Test ListProjects with offset beyond end",
			insertProjects: []model.Project{{Name: "Project1", Color: "#FF0000"}},
			params:         model.PaginationParams{Limit: 10, Offset: 100},
			wantLen:        0,
			wantTotal:      1,
		},
		{
			name: "Test ListProjects total count unaffected by limit",
			insertProjects: []model.Project{
				{Name: "Project1", Color: "#FF0000"},
				{Name: "Project2", Color: "#00FF00"},
				{Name: "Project3", Color: "#0000FF"},
			},
			params:    model.PaginationParams{Limit: 1, Offset: 0},
			wantLen:   1,
			wantTotal: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := memory.NewMemoryStore()

			for _, tagId := range tagIds {
				_, _ = ta.CreateTag(context.Background(), testScope, model.Tag{Id: tagId, Name: "Tag", Color: "#FFFFFF"})
			}

			insertedIds := make(map[uuid.UUID]bool)
			for i, project := range tt.insertProjects {
				createdProject, _ := ta.CreateProject(context.Background(), testScope, project)
				tt.insertProjects[i].Id = createdProject.Id
				insertedIds[createdProject.Id] = true
			}

			page, gotErr := ta.ListProjects(context.Background(), testScope, tt.params)
			if tt.wantErr {
				require.Error(t, gotErr)
				if tt.errType != nil {
					require.ErrorIs(t, gotErr, tt.errType)
				}
				return
			}

			require.NoError(t, gotErr)
			require.Len(t, page.Data, tt.wantLen)
			require.Equal(t, tt.wantTotal, page.TotalCount)
			require.Equal(t, tt.params.Limit, page.Limit)
			require.Equal(t, tt.params.Offset, page.Offset)

			for _, project := range page.Data {
				require.True(t, insertedIds[project.Id], "returned project not in inserted set")
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := memory.NewMemoryStore()
			for _, tagId := range tagIds {
				_, _ = ta.CreateTag(context.Background(), testScope, model.Tag{Id: tagId, Name: "Tag", Color: "#FFFFFF"})
			}
			for i, tagId := range tt.project.TagIds {
				tag, _ := ta.CreateTag(context.Background(), testScope, model.Tag{Id: tagId, Name: "Tag", Color: "#FFFFFF"})
				tagIds[i] = tag.Id
			}

			insertedProject, _ := ta.CreateProject(context.Background(), testScope, tt.project)
			editId := tt.editProjectId(&insertedProject)

			tt.editProject.Id = editId
			tt.want.Id = editId

			got, gotErr := ta.UpdateProject(context.Background(), testScope, tt.editProject)
			if tt.wantErr {
				require.Error(t, gotErr)
				if tt.errType != nil {
					require.ErrorIs(t, gotErr, tt.errType)
				}
				return
			}

			require.NoError(t, gotErr)
			require.NotEqual(t, uuid.Nil, got.Id)
			require.Equal(t, tt.want.Name, got.Name)
			require.Equal(t, tt.want.Color, got.Color)
			require.ElementsMatch(t, tt.want.TagIds, got.TagIds)

			if tt.want.TimeBudget == nil {
				require.Nil(t, got.TimeBudget)
			} else {
				require.NotNil(t, got.TimeBudget)
				require.Equal(t, *tt.want.TimeBudget, *got.TimeBudget)
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
			ta := memory.NewMemoryStore()

			project, _ := ta.CreateProject(context.Background(), testScope, tt.insertProject)
			deleteId := tt.deleteId(&project)

			gotErr := ta.DeleteProject(context.Background(), testScope, deleteId)
			if tt.wantErr {
				require.Error(t, gotErr)
				if tt.errType != nil {
					require.ErrorIs(t, gotErr, tt.errType)
				}
				return
			}

			require.NoError(t, gotErr)
			project, err := ta.GetProject(context.Background(), testScope, deleteId)
			require.ErrorIs(t, err, model.ErrNotFound)
		})
	}
}
