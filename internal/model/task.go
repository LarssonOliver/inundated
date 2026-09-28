package model

import (
	"time"

	"github.com/google/uuid"
)

// DefaultTaskTagColor is the color a task tag reports when its task carries
// no regular tag to borrow a color from (Nord's nord10).
const DefaultTaskTagColor = "#5e81ac"

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
}

func (t Task) Closed() bool {
	return t.CloseReason != nil
}

// TaskListParams filters ListTasks. Closed tasks are excluded unless
// IncludeClosed is set; PaginationParams.IncludeArchived is ignored.
type TaskListParams struct {
	PaginationParams
	IncludeClosed bool
	ParentId      *uuid.UUID
	TagId         *uuid.UUID
	DueFrom       *time.Time
	DueTo         *time.Time
}

// TagKind selects regular tags, task tags, or both in ListTags.
type TagKind string

const (
	TagKindLabel TagKind = "label"
	TagKindTask  TagKind = "task"
	TagKindAll   TagKind = "all"
)

// TagListParams extends PaginationParams for ListTags. An empty Kind means
// TagKindLabel, and an empty Query matches every name.
type TagListParams struct {
	PaginationParams
	Query string
	Kind  TagKind
}
