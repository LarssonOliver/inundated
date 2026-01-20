package memory

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/helpers"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
)

type ProjectStore struct {
	mu   sync.RWMutex
	data map[uuid.UUID]model.Project
}

var _ repository.ProjectRepository = (*ProjectStore)(nil)

func NewProjectStore() *ProjectStore {
	return &ProjectStore{
		mu:   sync.RWMutex{},
		data: make(map[uuid.UUID]model.Project),
	}
}

// CreateProject implements [repository.ProjectRepository].
func (t *ProjectStore) CreateProject(ctx context.Context, project model.Project) (model.Project, error) {
	if project.Name == "" || project.Color == "" || !helpers.IsValidColor(project.Color) {
		return model.Project{}, model.ErrInvalidArgument
	}

	var tagIds []uuid.UUID

	if project.TagIds != nil {
		tagIds = make([]uuid.UUID, len(project.TagIds))
		copy(tagIds, project.TagIds)
	} else {
		tagIds = []uuid.UUID{}
	}

	newId := uuid.New()
	newProject := model.Project{
		Id:         newId,
		Name:       project.Name,
		Color:      project.Color,
		TimeBudget: project.TimeBudget,
		TagIds:     tagIds,
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.data[newId] = newProject
	return newProject, nil
}

// GetProject implements [repository.ProjectRepository].
func (t *ProjectStore) GetProject(ctx context.Context, id uuid.UUID) (model.Project, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	project, exists := t.data[id]
	if !exists {
		return model.Project{}, model.ErrNotFound
	}

	return project, nil
}

// ListProjects implements [repository.ProjectRepository].
func (t *ProjectStore) ListProjects(ctx context.Context) ([]model.Project, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	projects := make([]model.Project, 0, len(t.data))

	for _, project := range t.data {
		projects = append(projects, project)
	}

	return projects, nil
}

// UpdateProject implements [repository.ProjectRepository].
func (t *ProjectStore) UpdateProject(ctx context.Context, project model.Project) (model.Project, error) {
	if project.Name == "" || project.Color == "" || !helpers.IsValidColor(project.Color) {
		return model.Project{}, model.ErrInvalidArgument
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	_, exists := t.data[project.Id]
	if !exists {
		return model.Project{}, model.ErrNotFound
	}

	t.data[project.Id] = project
	return project, nil
}

// DeleteProject implements [repository.ProjectRepository].
func (t *ProjectStore) DeleteProject(ctx context.Context, id uuid.UUID) error {
	// Skip locking for write if the project does not exist
	t.mu.RLock()
	_, exists := t.data[id]
	t.mu.RUnlock()

	if !exists {
		return model.ErrNotFound
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// delete is a noop if the key does not exist
	// thus, it does not matter if it has been deleted by another thread before this line
	delete(t.data, id)
	return nil
}
