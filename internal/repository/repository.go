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
	ProjectStatsRepository
}

type TagRepository interface {
	GetTag(ctx context.Context, id uuid.UUID) (model.Tag, error)
	ListTags(ctx context.Context, params model.PaginationParams) (model.Page[model.Tag], error)
	CreateTag(ctx context.Context, tag model.Tag) (model.Tag, error)
	UpdateTag(ctx context.Context, tag model.Tag) (model.Tag, error)
	DeleteTag(ctx context.Context, id uuid.UUID) error
}

type UserRepository interface {
	GetUser(ctx context.Context, id uuid.UUID) (model.User, error)
	GetUserBySub(ctx context.Context, sub string) (model.User, error)
	CreateUser(ctx context.Context, user model.User) (model.User, error)
	UpdateUser(ctx context.Context, user model.User) (model.User, error)

	// CreateUserAdoptingOrphans creates a user, like CreateUser. Additionally,
	// when the created user is the first user in the system, every project, tag
	// and timespan with a NULL user_id is assigned to that user in the same
	// atomic operation. The returned OrphanAdoption reports how many rows were
	// claimed; it is the zero value for every user after the first.
	CreateUserAdoptingOrphans(ctx context.Context, user model.User) (model.User, model.OrphanAdoption, error)
}

type ProjectRepository interface {
	GetProject(ctx context.Context, id uuid.UUID) (model.Project, error)
	ListProjects(ctx context.Context, params model.PaginationParams) (model.Page[model.Project], error)
	CreateProject(ctx context.Context, project model.Project) (model.Project, error)
	UpdateProject(ctx context.Context, project model.Project) (model.Project, error)
	DeleteProject(ctx context.Context, id uuid.UUID) error
}

type TimespanRepository interface {
	GetTimespan(ctx context.Context, id uuid.UUID) (model.Timespan, error)
	ListTimespans(ctx context.Context, params model.PaginationParams) (model.Page[model.Timespan], error)
	CreateTimespan(ctx context.Context, timespan model.Timespan) (model.Timespan, error)
	UpdateTimespan(ctx context.Context, timespan model.Timespan) (model.Timespan, error)
	DeleteTimespan(ctx context.Context, id uuid.UUID) error
	GetTotalDurationByTags(ctx context.Context, tagIds []uuid.UUID) (time.Duration, error)
}

type ProjectStatsRepository interface {
	AggregateTimeSpentByTagsAndBuckets(ctx context.Context, tagIds []uuid.UUID, buckets []model.BucketRange) ([]model.BucketValue, error)
}

// This interface is deliberately not included in the Repository interface.
// This allows for the session repository to be implemented in a different way
// than the other repositories, e.g. using valkey or similar.
type SessionRepository interface {
	GetSession(ctx context.Context, id uuid.UUID) (model.Session, error)
	CreateSession(ctx context.Context, session model.Session) (model.Session, error)
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
