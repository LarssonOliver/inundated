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

	t.mu.Lock()
	defer t.mu.Unlock()

	var tagIds []uuid.UUID

	if project.TagIds != nil {
		if !t.tagsUsable(scope, model.TagHolderProject, project.TagIds, nil) {
			return model.Project{}, model.ErrInvalidReference
		}

		deduped := utils.DedupeUUIDs(project.TagIds)
		tagIds = make([]uuid.UUID, len(deduped))
		copy(tagIds, deduped)
	} else {
		tagIds = []uuid.UUID{}
	}

	newId := uuid.New()
	// The project tag's name and color follow the project's.
	tag := model.Tag{
		Id:     uuid.New(),
		Name:   project.Name,
		Color:  project.Color,
		UserId: scope.UserID(),
		Owner:  &model.TagOwner{Kind: model.TagOwnerProject, Id: newId},
	}
	newProject := model.Project{
		Id:         newId,
		Name:       project.Name,
		Color:      project.Color,
		TimeBudget: project.TimeBudget,
		TagId:      tag.Id,
		TagIds:     tagIds,
		UserId:     scope.UserID(),
	}

	t.tags = append(t.tags, tag)
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
		if matchesScope(p.UserId, scope) && (params.IncludeArchived || !p.Archived) {
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

	t.mu.Lock()
	defer t.mu.Unlock()

	idx := slices.IndexFunc(t.projects, func(p model.Project) bool {
		return p.Id == project.Id && matchesScope(p.UserId, scope)
	})
	if idx == -1 {
		return model.Project{}, model.ErrNotFound
	}
	if project.TagIds != nil {
		if !t.tagsUsable(scope, model.TagHolderProject, project.TagIds, t.projects[idx].TagIds) {
			return model.Project{}, model.ErrInvalidReference
		}
		project.TagIds = utils.DedupeUUIDs(project.TagIds)
	}

	project.UserId = t.projects[idx].UserId
	project.TagId = t.projects[idx].TagId
	t.projects[idx] = project
	if tagIdx := slices.IndexFunc(t.tags, func(tag model.Tag) bool { return tag.Id == project.TagId }); tagIdx != -1 {
		t.tags[tagIdx].Name = project.Name
		t.tags[tagIdx].Color = project.Color
	}
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

	// A project with time attributed to its project tag can only be
	// archived. Time reaching it only through its linked tags stays with
	// those tags.
	tagId := t.projects[idx].TagId
	projectTag := map[uuid.UUID]struct{}{tagId: {}}
	for _, span := range t.timespans {
		if matchesScope(span.UserId, scope) && t.timespanHasAnyTag(span, projectTag) {
			return model.ErrConflict
		}
	}

	// The project tag goes with its project.
	t.projects = slices.Delete(t.projects, idx, idx+1)
	t.deleteTags(func(id uuid.UUID) bool { return id == tagId })
	return nil
}
