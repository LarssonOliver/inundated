package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

type Repository interface {
	UserRepository
	TagRepository
	ProjectRepository
	TimespanRepository
	TaskRepository
	ProjectStatsRepository
	SettingsRepository
}

type TagRepository interface {
	GetTag(ctx context.Context, scope model.OwnerScope, id uuid.UUID) (model.Tag, error)
	// GetTag and ListTags report a task tag with its TaskId set, and with
	// its color and archived state derived from its task: archived while
	// the task is closed, and colored like the task's first regular tag by
	// name (or model.DefaultTaskTagColor). ListTags puts regular tags
	// first, then orders by name case-insensitively, then by name bytes,
	// then by id.
	ListTags(ctx context.Context, scope model.OwnerScope, params model.TagListParams) (model.Page[model.Tag], error)
	CreateTag(ctx context.Context, scope model.OwnerScope, tag model.Tag) (model.Tag, error)
	// UpdateTag replaces the tag's mutable fields (including Archived)
	// wholesale with those on tag - it does not merge with the stored tag.
	// Callers must fetch the current tag first and copy forward any field
	// they don't intend to change, or that field resets to its zero value
	// (e.g. an omitted/false Archived unarchives the tag).
	//
	// UpdateTag and DeleteTag refuse task tags, which follow their task,
	// with model.ErrInvalidArgument. Deleting a tag drops it from the
	// TagIds of every task, project and timespan.
	UpdateTag(ctx context.Context, scope model.OwnerScope, tag model.Tag) (model.Tag, error)
	DeleteTag(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error
}

type TaskRepository interface {
	GetTask(ctx context.Context, scope model.OwnerScope, id uuid.UUID) (model.Task, error)
	// ListTasks orders tasks by parent, then by rank among siblings.
	ListTasks(ctx context.Context, scope model.OwnerScope, params model.TaskListParams) (model.Page[model.Task], error)
	// CreateTask creates the task together with its task tag and places it
	// last among its siblings. The parent must be an open task in scope,
	// and TagIds must name regular tags only. Id, TagId, Rank and the
	// closed fields on task are ignored.
	CreateTask(ctx context.Context, scope model.OwnerScope, task model.Task) (model.Task, error)
	// UpdateTask applies patch to the stored task within the same write,
	// so concurrent updates to different fields don't undo each other;
	// parent and rank only change through MoveTask. Closing an open task
	// also closes its open descendants with the same reason, and reopening
	// a task also reopens its closed ancestors. The task tag's name follows
	// the task's.
	UpdateTask(ctx context.Context, scope model.OwnerScope, id uuid.UUID, patch model.TaskPatch) (model.Task, error)
	// MoveTask places the task under parentId (nil for the top level),
	// directly after the sibling afterId (nil for first). A parent that is
	// the task itself or one of its descendants, a closed parent for an
	// open task, or an afterId that isn't a sibling under the new parent is
	// model.ErrInvalidArgument.
	MoveTask(ctx context.Context, scope model.OwnerScope, id uuid.UUID, parentId *uuid.UUID, afterId *uuid.UUID) (model.Task, error)
	// DeleteTask deletes the task, its descendants and their task tags. It
	// fails with model.ErrConflict when any of them has logged time.
	DeleteTask(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error
	// ListTaskProjectIds maps each of taskIds to the projects in scope
	// whose tags meet the task's effective tags: its task tag, its regular
	// tags, and those of its ancestors. Tasks with no projects are left
	// out.
	ListTaskProjectIds(ctx context.Context, scope model.OwnerScope, taskIds []uuid.UUID) (map[uuid.UUID][]uuid.UUID, error)
}

type UserRepository interface {
	GetUser(ctx context.Context, id uuid.UUID) (model.User, error)
	GetUserBySub(ctx context.Context, sub string) (model.User, error)
	HasUsers(ctx context.Context) (bool, error)
	CreateUser(ctx context.Context, user model.User) (model.User, error)
	UpdateUser(ctx context.Context, user model.User) (model.User, error)
	CreateUserAdoptingOrphans(ctx context.Context, user model.User) (model.User, model.OrphanAdoption, error)
}

type ProjectRepository interface {
	GetProject(ctx context.Context, scope model.OwnerScope, id uuid.UUID) (model.Project, error)
	ListProjects(ctx context.Context, scope model.OwnerScope, params model.PaginationParams) (model.Page[model.Project], error)
	CreateProject(ctx context.Context, scope model.OwnerScope, project model.Project) (model.Project, error)
	// UpdateProject replaces the project's mutable fields (including
	// Archived) wholesale with those on project - it does not merge with
	// the stored project. Callers must fetch the current project first and
	// copy forward any field they don't intend to change, or that field
	// resets to its zero value (e.g. an omitted/false Archived unarchives
	// the project).
	UpdateProject(ctx context.Context, scope model.OwnerScope, project model.Project) (model.Project, error)
	DeleteProject(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error
}

type TimespanRepository interface {
	GetTimespan(ctx context.Context, scope model.OwnerScope, id uuid.UUID) (model.Timespan, error)
	ListTimespans(ctx context.Context, scope model.OwnerScope, params model.TimespanListParams) (model.Page[model.Timespan], error)
	CreateTimespan(ctx context.Context, scope model.OwnerScope, timespan model.Timespan) (model.Timespan, error)
	UpdateTimespan(ctx context.Context, scope model.OwnerScope, timespan model.Timespan) (model.Timespan, error)
	DeleteTimespan(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error
	GetTotalDurationByTags(ctx context.Context, scope model.OwnerScope, tagIds []uuid.UUID) (time.Duration, error)
	// GetTotalDurationPerTag totals each of tagIds on its own, as
	// GetTotalDurationByTags would for that one tag, in a single pass.
	// Tags with no time are left out of the map.
	GetTotalDurationPerTag(ctx context.Context, scope model.OwnerScope, tagIds []uuid.UUID) (map[uuid.UUID]time.Duration, error)
	// GetTaskDurationByTags totals the timespans logged on any task whose
	// effective tags include one of tagIds, open or closed, counting each
	// timespan once. For a project's tags, that is the part of the
	// project's total logged on its tasks.
	GetTaskDurationByTags(ctx context.Context, scope model.OwnerScope, tagIds []uuid.UUID) (time.Duration, error)
}

type ProjectStatsRepository interface {
	AggregateTimeSpentByTagsAndBuckets(ctx context.Context, scope model.OwnerScope, tagIds []uuid.UUID, buckets []model.BucketRange) ([]model.BucketValue, error)
}

// SettingsRepository manages the singleton Settings row for a scope. Unlike
// the other user-scoped resources there is no id/list/delete: a scope always
// has at most one row, lazily created (see GetSettings's callers) rather than
// user-initiated.
type SettingsRepository interface {
	GetSettings(ctx context.Context, scope model.OwnerScope) (model.Settings, error)
	CreateSettings(ctx context.Context, scope model.OwnerScope, settings model.Settings) (model.Settings, error)
	// UpdateSettings replaces the settings row's fields wholesale with those
	// on settings - it does not merge with the stored row. Callers must fetch
	// the current settings first (e.g. via GetSettings) and copy forward any
	// field they don't intend to change.
	UpdateSettings(ctx context.Context, scope model.OwnerScope, settings model.Settings) (model.Settings, error)
}

// This interface is deliberately not included in the Repository interface.
// This allows for the session repository to be implemented in a different way
// than the other repositories, e.g. using valkey or similar.
type SessionRepository interface {
	// GetSessionByToken resolves a session from the opaque token the client
	// presented in its cookie. The store keeps only a hash of the token, so
	// it is matched by hash and never handed back.
	GetSessionByToken(ctx context.Context, token string) (model.Session, error)
	// CreateSession persists session, keyed by the hash of the given raw
	// token. The token is not stored in the clear and is not part of the
	// returned session.
	CreateSession(ctx context.Context, session model.Session, token string) (model.Session, error)
	TouchSession(ctx context.Context, id uuid.UUID, expiresAt time.Time) (model.Session, error)
	DeleteSession(ctx context.Context, id uuid.UUID) error
	DeleteAllExpiredSessions(ctx context.Context) error
}

// This interface is deliberately not included in the Repository interface.
// This allows for the session repository to be implemented in a different way
// than the other repositories, e.g. using valkey or similar.
type LoginStateRepository interface {
	GetLoginState(ctx context.Context, id uuid.UUID) (model.LoginState, error)
	CreateLoginState(ctx context.Context, loginState model.LoginState) (model.LoginState, error)
	DeleteLoginState(ctx context.Context, id uuid.UUID) error
	DeleteAllExpiredLoginStates(ctx context.Context) error
}
