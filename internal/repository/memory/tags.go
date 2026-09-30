package memory

import (
	"cmp"
	"context"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/utils"
)

// tagsUsable reports whether every id refers to a live tag owned by scope
// that holder may carry (see model.TagHolder.MayCarry).
// An archived tag is only acceptable if it's already in alreadyAssociated
// (i.e. it was attached before this call) - that keeps existing
// associations with a since-archived tag intact across unrelated edits,
// while still blocking a fresh attachment of an archived tag that a picker
// would never surface. Callers must hold t.mu for writing, so the check and
// the write that relies on it see the same state.
func (t *MemoryStore) tagsUsable(scope model.OwnerScope, holder model.TagHolder, tagIds []uuid.UUID, alreadyAssociated []uuid.UUID) bool {
	if len(tagIds) == 0 {
		return true
	}
	allowedArchived := make(map[uuid.UUID]bool, len(alreadyAssociated))
	for _, id := range alreadyAssociated {
		allowedArchived[id] = true
	}

	lookup := t.newTagLookup()
	for _, tagId := range tagIds {
		stored, ok := lookup.tags[tagId]
		if !ok || !matchesScope(stored.UserId, scope) || !holder.MayCarry(stored) {
			return false
		}
		if lookup.view(stored).Archived && !allowedArchived[tagId] {
			return false
		}
	}
	return true
}

// CreateTag implements [repository.TagRepository].
func (t *MemoryStore) CreateTag(ctx context.Context, scope model.OwnerScope, tag model.Tag) (model.Tag, error) {
	if tag.Name == "" || tag.Color == "" || !utils.IsValidColor(tag.Color) {
		return model.Tag{}, model.ErrInvalidArgument
	}

	if tag.Id == uuid.Nil {
		tag.Id = uuid.New()
	}

	newTag := model.Tag{
		Id:     tag.Id,
		Name:   tag.Name,
		Color:  tag.Color,
		UserId: scope.UserID(),
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.tags = append(t.tags, newTag)
	return newTag, nil
}

// GetTag implements [repository.TagRepository].
func (t *MemoryStore) GetTag(ctx context.Context, scope model.OwnerScope, id uuid.UUID) (model.Tag, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	idx := slices.IndexFunc(t.tags, func(tag model.Tag) bool { return tag.Id == id })
	if idx == -1 || !matchesScope(t.tags[idx].UserId, scope) {
		return model.Tag{}, model.ErrNotFound
	}

	return t.newTagLookup().view(t.tags[idx]), nil
}

// tagLookup indexes tags and their owners by id, so viewing many tags
// doesn't rescan every slice for each one.
type tagLookup struct {
	tags     map[uuid.UUID]model.Tag
	tasks    map[uuid.UUID]model.Task
	projects map[uuid.UUID]model.Project
}

// newTagLookup indexes t's tags and their owners. Callers must hold t.mu.
func (t *MemoryStore) newTagLookup() tagLookup {
	l := tagLookup{
		tags:     make(map[uuid.UUID]model.Tag, len(t.tags)),
		tasks:    make(map[uuid.UUID]model.Task, len(t.tasks)),
		projects: make(map[uuid.UUID]model.Project, len(t.projects)),
	}
	for _, tag := range t.tags {
		l.tags[tag.Id] = tag
	}
	for _, task := range t.tasks {
		l.tasks[task.Id] = task
	}
	for _, project := range t.projects {
		l.projects[project.Id] = project
	}
	return l
}

// view returns tag as readers see it: an owned tag takes its archived
// state from its owner (a task tag is archived while its task is closed).
func (l tagLookup) view(tag model.Tag) model.Tag {
	if tag.Owner == nil {
		return tag
	}
	switch tag.Owner.Kind {
	case model.TagOwnerTask:
		if task, ok := l.tasks[tag.Owner.Id]; ok {
			tag.Archived = task.Closed()
		}
	case model.TagOwnerProject:
		if project, ok := l.projects[tag.Owner.Id]; ok {
			tag.Archived = project.Archived
		}
	}
	return tag
}

// ListDerivedTagSources implements [repository.TagRepository].
func (t *MemoryStore) ListDerivedTagSources(ctx context.Context, scope model.OwnerScope, tagIds []uuid.UUID) (map[uuid.UUID][]model.Tag, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	lookup := t.newTagLookup()
	out := map[uuid.UUID][]model.Tag{}
	for _, id := range tagIds {
		derived, ok := lookup.tags[id]
		if !ok || !matchesScope(derived.UserId, scope) || !derived.IsDerived() {
			continue
		}
		// A task tag's sources are its task's regular tags, not the project
		// tags it also carries.
		task, ok := lookup.tasks[derived.Owner.Id]
		if !ok || derived.Owner.Kind != model.TagOwnerTask {
			continue
		}
		for _, sourceId := range task.TagIds {
			if source, ok := lookup.tags[sourceId]; ok && source.Owner == nil {
				out[id] = append(out[id], source)
			}
		}
	}
	return out, nil
}

// ListTags implements [repository.TagRepository].
func (t *MemoryStore) ListTags(ctx context.Context, scope model.OwnerScope, params model.TagListParams) (model.Page[model.Tag], error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	query := strings.ToLower(params.Query)
	lookup := t.newTagLookup()
	all := make([]model.Tag, 0, len(t.tags))
	for _, stored := range t.tags {
		tag := lookup.view(stored)
		if !matchesScope(tag.UserId, scope) || (!params.IncludeArchived && tag.Archived) {
			continue
		}
		if !params.SelectsKind(tag.Owner) {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(tag.Name), query) {
			continue
		}
		if params.Ids != nil && !slices.Contains(params.Ids, tag.Id) {
			continue
		}
		all = append(all, tag)
	}

	// Kinds take turns (see model.TagListParams), as the Postgres store
	// orders them: number each tag within its kind by name, then order by
	// that number, regular tags first within each turn.
	kindOf := func(tag model.Tag) string {
		if tag.Owner == nil {
			return ""
		}
		return string(tag.Owner.Kind)
	}
	slices.SortStableFunc(all, model.CompareTagNames)
	turn := make(map[uuid.UUID]int, len(all))
	perKind := map[string]int{}
	for _, tag := range all {
		turn[tag.Id] = perKind[kindOf(tag)]
		perKind[kindOf(tag)]++
	}
	slices.SortStableFunc(all, func(a, b model.Tag) int {
		return cmp.Or(cmp.Compare(turn[a.Id], turn[b.Id]), cmp.Compare(kindOf(a), kindOf(b)))
	})

	total := len(all)
	start := min(params.Offset, total)
	end := min(start+params.Limit, total)

	return model.Page[model.Tag]{
		Data:       all[start:end],
		TotalCount: total,
		Limit:      params.Limit,
		Offset:     params.Offset,
	}, nil
}

// UpdateTag implements [repository.TagRepository].
func (t *MemoryStore) UpdateTag(ctx context.Context, scope model.OwnerScope, tag model.Tag) (model.Tag, error) {
	if tag.Name == "" || tag.Color == "" || !utils.IsValidColor(tag.Color) {
		return model.Tag{}, model.ErrInvalidArgument
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	idx := slices.IndexFunc(t.tags, func(existing model.Tag) bool {
		return existing.Id == tag.Id && matchesScope(existing.UserId, scope)
	})
	if idx == -1 {
		return model.Tag{}, model.ErrNotFound
	}
	if t.tags[idx].Owner != nil {
		return model.Tag{}, model.ErrInvalidArgument
	}

	tag.UserId = t.tags[idx].UserId
	tag.Owner = nil
	t.tags[idx] = tag
	return tag, nil
}

// DeleteTag implements [repository.TagRepository].
func (t *MemoryStore) DeleteTag(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	idx := slices.IndexFunc(t.tags, func(existing model.Tag) bool {
		return existing.Id == id && matchesScope(existing.UserId, scope)
	})
	if idx == -1 {
		return model.ErrNotFound
	}
	if t.tags[idx].Owner != nil {
		return model.ErrInvalidArgument
	}

	t.deleteTags(func(tagId uuid.UUID) bool { return tagId == id })
	return nil
}

// deleteTags deletes the tags isDeleted picks and drops them from the
// TagIds of every task, project and timespan. Callers must hold t.mu for
// writing.
func (t *MemoryStore) deleteTags(isDeleted func(tagId uuid.UUID) bool) {
	t.tags = slices.DeleteFunc(t.tags, func(tag model.Tag) bool { return isDeleted(tag.Id) })
	for i := range t.tasks {
		t.tasks[i].TagIds = slices.DeleteFunc(t.tasks[i].TagIds, isDeleted)
	}
	for i := range t.projects {
		t.projects[i].TagIds = slices.DeleteFunc(t.projects[i].TagIds, isDeleted)
	}
	for i := range t.timespans {
		t.timespans[i].TagIds = slices.DeleteFunc(t.timespans[i].TagIds, isDeleted)
	}
}
