package model

import (
	"time"

	"github.com/google/uuid"
)

type CloseReason string

const (
	CloseReasonDone    CloseReason = "done"
	CloseReasonIgnored CloseReason = "ignored"
)

func (r CloseReason) Valid() bool {
	return r == CloseReasonDone || r == CloseReasonIgnored
}

// Task is a to-do item. Every task owns a task tag (TagId): a timespan
// carrying that tag is time logged on the task, and a project carrying it
// has the task assigned to it. TagIds holds the task's own regular tags and
// never contains task tags.
type Task struct {
	Id       uuid.UUID
	Name     string
	TagId    uuid.UUID
	ParentId *uuid.UUID
	TagIds   []uuid.UUID
	// DueDate is a calendar day, stored as midnight UTC.
	DueDate  *time.Time
	Estimate *time.Duration
	// Rank orders the task among its siblings; see [utils.RankBetween].
	Rank string
	// CloseReason is nil while the task is open.
	CloseReason *CloseReason
	ClosedAt    *time.Time
	UserId      *uuid.UUID

	// TotalTime and ProjectIds are only filled in when requested.
	// TotalTime is the time logged on the task and its subtasks.
	// ProjectIds are the projects whose tags meet the task's effective
	// tags, whether it was assigned directly or shares a regular tag.
	TotalTime  *time.Duration
	ProjectIds []uuid.UUID
}

func (t Task) Closed() bool {
	return t.CloseReason != nil
}

// TaskPatch is a partial update to a task. Nil fields keep the stored
// value. Applying it to the stored task inside the write, rather than
// writing back a whole task read earlier, keeps concurrent updates from
// undoing each other's changes.
type TaskPatch struct {
	Name          *string
	TagIds        *[]uuid.UUID
	DueDate       *time.Time
	ClearDueDate  bool
	Estimate      *time.Duration
	ClearEstimate bool
	// Closed opens or closes the task. Closing an open task without a
	// CloseReason closes it as done; closing a closed one keeps its reason.
	Closed *bool
	// CloseReason closes the task, or changes why it's closed.
	CloseReason *CloseReason
}

// Apply returns task with p's changes made.
func (p TaskPatch) Apply(task Task) Task {
	if p.Name != nil {
		task.Name = *p.Name
	}
	if p.TagIds != nil {
		task.TagIds = *p.TagIds
	}
	if p.DueDate != nil {
		task.DueDate = p.DueDate
	}
	if p.ClearDueDate {
		task.DueDate = nil
	}
	if p.Estimate != nil {
		task.Estimate = p.Estimate
	}
	if p.ClearEstimate {
		task.Estimate = nil
	}
	switch {
	case p.CloseReason != nil:
		reason := *p.CloseReason
		task.CloseReason = &reason
	case p.Closed != nil && *p.Closed && task.CloseReason == nil:
		reason := CloseReasonDone
		task.CloseReason = &reason
	case p.Closed != nil && !*p.Closed:
		task.CloseReason = nil
	}
	return task
}

// TaskListParams filters ListTasks. Closed tasks are excluded unless
// IncludeClosed is set; PaginationParams.IncludeArchived is ignored.
type TaskListParams struct {
	PaginationParams
	IncludeClosed bool
	ParentId      *uuid.UUID
	TagId         *uuid.UUID
	// ProjectId keeps the tasks that belong to the project; see
	// Task.ProjectIds.
	ProjectId *uuid.UUID
	DueFrom   *time.Time
	DueTo     *time.Time
}

// TagKind selects the tags ListTags returns: regular tags (TagKindLabel),
// every tag (TagKindAll), or the tags owned by one kind of owner, whose
// TagKind is the TagOwnerKind itself (TagKindTask, TagKindProject).
type TagKind string

const (
	TagKindLabel   TagKind = "label"
	TagKindTask            = TagKind(TagOwnerTask)
	TagKindProject         = TagKind(TagOwnerProject)
	TagKindAll     TagKind = "all"
)

// OwnerKind returns the owner kind k selects, if it selects one.
func (k TagKind) OwnerKind() (TagOwnerKind, bool) {
	owner := TagOwnerKind(k)
	return owner, owner.Valid()
}

// Valid reports whether k is a kind ListTags knows.
func (k TagKind) Valid() bool {
	_, owned := k.OwnerKind()
	return owned || k == TagKindLabel || k == TagKindAll
}

// TagListParams extends PaginationParams for ListTags. Empty Kinds means
// TagKindLabel alone, and an empty Query matches every name.
//
// With several kinds, they take turns filling each page: the first tag of
// each kind by name, then the second of each, and so on, regular tags
// first and owned ones by owner kind within each turn. So one kind's
// matches can't crowd another's out of a page. With one kind, that's name
// order.
type TagListParams struct {
	PaginationParams
	Query string
	Kinds []TagKind
	// Ids, when set, keeps only these tags.
	Ids []uuid.UUID
}

// SelectsKind reports whether a tag owned by owner (nil for a regular tag)
// is of one of p's kinds.
func (p TagListParams) SelectsKind(owner *TagOwner) bool {
	kinds := p.Kinds
	if len(kinds) == 0 {
		kinds = []TagKind{TagKindLabel}
	}
	for _, k := range kinds {
		switch {
		case k == TagKindAll:
			return true
		case k == TagKindLabel && owner == nil:
			return true
		case owner != nil && k == TagKind(owner.Kind):
			return true
		}
	}
	return false
}
