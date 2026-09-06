package memory

import (
	"context"
	"slices"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/utils"
)

// CreateProject implements [repository.ProjectRepository].
func (t *MemoryStore) CreateProject(ctx context.Context, scope model.OwnerScope, project model.Project) (model.Project, error) {
	if project.Name == "" || project.Color == "" || !utils.IsValidColor(project.Color) {
		return model.Project{}, model.ErrInvalidArgument
	}

	var tagIds []uuid.UUID

	if project.TagIds != nil {
		if !t.tagsExist(ctx, scope, project.TagIds) {
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
		UserId:     scope.UserID(),
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.projects = append(t.projects, newProject)
	return newProject, nil
}

// GetProject implements [repository.ProjectRepository].
func (t *MemoryStore) GetProject(ctx context.Context, scope model.OwnerScope, id uuid.UUID) (model.Project, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	idx := slices.IndexFunc(t.projects, func(p model.Project) bool { return p.Id == id })
	if idx == -1 || !matchesScope(t.projects[idx].UserId, scope) {
		return model.Project{}, model.ErrNotFound
	}

	return t.projects[idx], nil
}

// ListProjects implements [repository.ProjectRepository].
func (t *MemoryStore) ListProjects(ctx context.Context, scope model.OwnerScope, params model.PaginationParams) (model.Page[model.Project], error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	all := make([]model.Project, 0, len(t.projects))
	for _, p := range t.projects {
		if matchesScope(p.UserId, scope) {
			all = append(all, p)
		}
	}

	total := len(all)
	start := min(params.Offset, total)
	end := min(start+params.Limit, total)

	return model.Page[model.Project]{
		Data:       all[start:end],
		TotalCount: total,
		Limit:      params.Limit,
		Offset:     params.Offset,
	}, nil
}

// UpdateProject implements [repository.ProjectRepository].
func (t *MemoryStore) UpdateProject(ctx context.Context, scope model.OwnerScope, project model.Project) (model.Project, error) {
	if project.Name == "" || project.Color == "" || !utils.IsValidColor(project.Color) {
		return model.Project{}, model.ErrInvalidArgument
	}

	if project.TagIds != nil && !t.tagsExist(ctx, scope, project.TagIds) {
		return model.Project{}, model.ErrInvalidReference
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	idx := slices.IndexFunc(t.projects, func(p model.Project) bool {
		return p.Id == project.Id && matchesScope(p.UserId, scope)
	})
	if idx == -1 {
		return model.Project{}, model.ErrNotFound
	}

	project.UserId = t.projects[idx].UserId
	t.projects[idx] = project
	return project, nil
}

// DeleteProject implements [repository.ProjectRepository].
func (t *MemoryStore) DeleteProject(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	idx := slices.IndexFunc(t.projects, func(p model.Project) bool {
		return p.Id == id && matchesScope(p.UserId, scope)
	})
	if idx == -1 {
		return model.ErrNotFound
	}

	t.projects = slices.Delete(t.projects, idx, idx+1)
	return nil
}
