package memory

import (
	"context"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/utils"
)

// CreateProject implements [repository.ProjectRepository].
func (t *MemoryStore) CreateProject(ctx context.Context, project model.Project) (model.Project, error) {
	if project.Name == "" || project.Color == "" || !utils.IsValidColor(project.Color) {
		return model.Project{}, model.ErrInvalidArgument
	}

	var tagIds []uuid.UUID

	if project.TagIds != nil {
		if !t.tagsExist(ctx, project.TagIds) {
			return model.Project{}, model.ErrInvalidReference
		}

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

	t.projects[newId] = newProject
	return newProject, nil
}

// GetProject implements [repository.ProjectRepository].
func (t *MemoryStore) GetProject(ctx context.Context, id uuid.UUID) (model.Project, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	project, exists := t.projects[id]
	if !exists {
		return model.Project{}, model.ErrNotFound
	}

	return project, nil
}

// ListProjects implements [repository.ProjectRepository].
func (t *MemoryStore) ListProjects(ctx context.Context, params model.PaginationParams) (model.Page[model.Project], error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	all := make([]model.Project, 0, len(t.projects))
	for _, project := range t.projects {
		all = append(all, project)
	}

	total := len(all)
	start := params.Offset
	if start > total {
		start = total
	}
	end := start + params.Limit
	if end > total {
		end = total
	}

	return model.Page[model.Project]{
		Data:       all[start:end],
		TotalCount: total,
		Limit:      params.Limit,
		Offset:     params.Offset,
	}, nil
}

// UpdateProject implements [repository.ProjectRepository].
func (t *MemoryStore) UpdateProject(ctx context.Context, project model.Project) (model.Project, error) {
	if project.Name == "" || project.Color == "" || !utils.IsValidColor(project.Color) {
		return model.Project{}, model.ErrInvalidArgument
	}

	if project.TagIds != nil && !t.tagsExist(ctx, project.TagIds) {
		return model.Project{}, model.ErrInvalidReference
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	_, exists := t.projects[project.Id]
	if !exists {
		return model.Project{}, model.ErrNotFound
	}

	t.projects[project.Id] = project
	return project, nil
}

// DeleteProject implements [repository.ProjectRepository].
func (t *MemoryStore) DeleteProject(ctx context.Context, id uuid.UUID) error {
	// Skip locking for write if the project does not exist
	t.mu.RLock()
	_, exists := t.projects[id]
	t.mu.RUnlock()

	if !exists {
		return model.ErrNotFound
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// delete is a noop if the key does not exist
	// thus, it does not matter if it has been deleted by another thread before this line
	delete(t.projects, id)
	return nil
}
