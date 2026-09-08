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
	GetTag(ctx context.Context, scope model.OwnerScope, id uuid.UUID) (model.Tag, error)
	ListTags(ctx context.Context, scope model.OwnerScope, params model.PaginationParams) (model.Page[model.Tag], error)
	CreateTag(ctx context.Context, scope model.OwnerScope, tag model.Tag) (model.Tag, error)
	UpdateTag(ctx context.Context, scope model.OwnerScope, tag model.Tag) (model.Tag, error)
	DeleteTag(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error
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
	UpdateProject(ctx context.Context, scope model.OwnerScope, project model.Project) (model.Project, error)
	DeleteProject(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error
}

type TimespanRepository interface {
	GetTimespan(ctx context.Context, scope model.OwnerScope, id uuid.UUID) (model.Timespan, error)
	ListTimespans(ctx context.Context, scope model.OwnerScope, params model.PaginationParams) (model.Page[model.Timespan], error)
	CreateTimespan(ctx context.Context, scope model.OwnerScope, timespan model.Timespan) (model.Timespan, error)
	UpdateTimespan(ctx context.Context, scope model.OwnerScope, timespan model.Timespan) (model.Timespan, error)
	DeleteTimespan(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error
	GetTotalDurationByTags(ctx context.Context, scope model.OwnerScope, tagIds []uuid.UUID) (time.Duration, error)
}

type ProjectStatsRepository interface {
	AggregateTimeSpentByTagsAndBuckets(ctx context.Context, scope model.OwnerScope, tagIds []uuid.UUID, buckets []model.BucketRange) ([]model.BucketValue, error)
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
