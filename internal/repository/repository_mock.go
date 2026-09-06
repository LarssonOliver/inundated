package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

var _ Repository = (*RepoMock)(nil)

type RepoMock struct {
	CreateProjectFn func(ctx context.Context, scope model.OwnerScope, project model.Project) (model.Project, error)
	DeleteProjectFn func(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error
	GetProjectFn    func(ctx context.Context, scope model.OwnerScope, id uuid.UUID) (model.Project, error)
	ListProjectFn   func(ctx context.Context, scope model.OwnerScope, params model.PaginationParams) (model.Page[model.Project], error)
	UpdateProjectFn func(ctx context.Context, scope model.OwnerScope, project model.Project) (model.Project, error)

	CreateTagFn func(ctx context.Context, scope model.OwnerScope, tag model.Tag) (model.Tag, error)
	DeleteTagFn func(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error
	GetTagFn    func(ctx context.Context, scope model.OwnerScope, id uuid.UUID) (model.Tag, error)
	ListTagFn   func(ctx context.Context, scope model.OwnerScope, params model.PaginationParams) (model.Page[model.Tag], error)
	UpdateTagFn func(ctx context.Context, scope model.OwnerScope, tag model.Tag) (model.Tag, error)

	GetUserFn                   func(ctx context.Context, id uuid.UUID) (model.User, error)
	GetUserBySubFn              func(ctx context.Context, sub string) (model.User, error)
	CreateUserFn                func(ctx context.Context, user model.User) (model.User, error)
	CreateUserAdoptingOrphansFn func(ctx context.Context, user model.User) (model.User, model.OrphanAdoption, error)
	UpdateUserFn                func(ctx context.Context, user model.User) (model.User, error)

	CreateTimespanFn         func(ctx context.Context, scope model.OwnerScope, timespan model.Timespan) (model.Timespan, error)
	DeleteTimespanFn         func(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error
	GetTimespanFn            func(ctx context.Context, scope model.OwnerScope, id uuid.UUID) (model.Timespan, error)
	ListTimespanFn           func(ctx context.Context, scope model.OwnerScope, params model.PaginationParams) (model.Page[model.Timespan], error)
	UpdateTimespanFn         func(ctx context.Context, scope model.OwnerScope, timespan model.Timespan) (model.Timespan, error)
	GetTotalDurationByTagsFn func(ctx context.Context, scope model.OwnerScope, tagIds []uuid.UUID) (time.Duration, error)

	AggregateTimeSpentByTagsAndBucketsFn func(ctx context.Context, scope model.OwnerScope, tagIds []uuid.UUID, buckets []model.BucketRange) ([]model.BucketValue, error)
}

var _ SessionRepository = (*SessionRepoMock)(nil)

type SessionRepoMock struct {
	GetSessionFn               func(ctx context.Context, id uuid.UUID) (model.Session, error)
	CreateSessionFn            func(ctx context.Context, session model.Session) (model.Session, error)
	TouchSessionFn             func(ctx context.Context, id uuid.UUID, expiresAt time.Time) (model.Session, error)
	DeleteSessionFn            func(ctx context.Context, id uuid.UUID) error
	DeleteAllExpiredSessionsFn func(ctx context.Context) error
}

var _ LoginStateRepository = (*LoginStateRepoMock)(nil)

type LoginStateRepoMock struct {
	GetLoginStateFn               func(ctx context.Context, id uuid.UUID) (model.LoginState, error)
	CreateLoginStateFn            func(ctx context.Context, state model.LoginState) (model.LoginState, error)
	DeleteLoginStateFn            func(ctx context.Context, id uuid.UUID) error
	DeleteAllExpiredLoginStatesFn func(ctx context.Context) error
}

// CreateProject implements repository.ProjectRepository.
func (t *RepoMock) CreateProject(ctx context.Context, scope model.OwnerScope, project model.Project) (model.Project, error) {
	return t.CreateProjectFn(ctx, scope, project)
}

// DeleteProject implements repository.ProjectRepository.
func (t *RepoMock) DeleteProject(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error {
	return t.DeleteProjectFn(ctx, scope, id)
}

// GetProject implements repository.ProjectRepository.
func (t *RepoMock) GetProject(ctx context.Context, scope model.OwnerScope, id uuid.UUID) (model.Project, error) {
	return t.GetProjectFn(ctx, scope, id)
}

// ListProjects implements repository.ProjectRepository.
func (t *RepoMock) ListProjects(ctx context.Context, scope model.OwnerScope, params model.PaginationParams) (model.Page[model.Project], error) {
	return t.ListProjectFn(ctx, scope, params)
}

// UpdateProject implements repository.ProjectRepository.
func (t *RepoMock) UpdateProject(ctx context.Context, scope model.OwnerScope, project model.Project) (model.Project, error) {
	return t.UpdateProjectFn(ctx, scope, project)
}

// CreateTag implements repository.TagRepository.
func (t *RepoMock) CreateTag(ctx context.Context, scope model.OwnerScope, tag model.Tag) (model.Tag, error) {
	return t.CreateTagFn(ctx, scope, tag)
}

// DeleteTag implements repository.TagRepository.
func (t *RepoMock) DeleteTag(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error {
	return t.DeleteTagFn(ctx, scope, id)
}

// GetTag implements repository.TagRepository.
func (t *RepoMock) GetTag(ctx context.Context, scope model.OwnerScope, id uuid.UUID) (model.Tag, error) {
	return t.GetTagFn(ctx, scope, id)
}

// ListTags implements repository.TagRepository.
func (t *RepoMock) ListTags(ctx context.Context, scope model.OwnerScope, params model.PaginationParams) (model.Page[model.Tag], error) {
	return t.ListTagFn(ctx, scope, params)
}

// UpdateTag implements repository.TagRepository.
func (t *RepoMock) UpdateTag(ctx context.Context, scope model.OwnerScope, tag model.Tag) (model.Tag, error) {
	return t.UpdateTagFn(ctx, scope, tag)
}

// GetUser implements repository.UserRepository.
func (t *RepoMock) GetUser(ctx context.Context, id uuid.UUID) (model.User, error) {
	return t.GetUserFn(ctx, id)
}

// GetUserBySub implements repository.UserRepository.
func (t *RepoMock) GetUserBySub(ctx context.Context, sub string) (model.User, error) {
	return t.GetUserBySubFn(ctx, sub)
}

// CreateUser implements repository.UserRepository.
func (t *RepoMock) CreateUser(ctx context.Context, user model.User) (model.User, error) {
	return t.CreateUserFn(ctx, user)
}

// CreateUserAdoptingOrphans implements repository.UserRepository.
func (t *RepoMock) CreateUserAdoptingOrphans(ctx context.Context, user model.User) (model.User, model.OrphanAdoption, error) {
	return t.CreateUserAdoptingOrphansFn(ctx, user)
}

// UpdateUser implements repository.UserRepository.
func (t *RepoMock) UpdateUser(ctx context.Context, user model.User) (model.User, error) {
	return t.UpdateUserFn(ctx, user)
}

// CreateTimespan implements repository.TimespanRepository.
func (t *RepoMock) CreateTimespan(ctx context.Context, scope model.OwnerScope, timespan model.Timespan) (model.Timespan, error) {
	return t.CreateTimespanFn(ctx, scope, timespan)
}

// DeleteTimespan implements repository.TimespanRepository.
func (t *RepoMock) DeleteTimespan(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error {
	return t.DeleteTimespanFn(ctx, scope, id)
}

// GetTimespan implements repository.TimespanRepository.
func (t *RepoMock) GetTimespan(ctx context.Context, scope model.OwnerScope, id uuid.UUID) (model.Timespan, error) {
	return t.GetTimespanFn(ctx, scope, id)
}

// ListTimespans implements repository.TimespanRepository.
func (t *RepoMock) ListTimespans(ctx context.Context, scope model.OwnerScope, params model.PaginationParams) (model.Page[model.Timespan], error) {
	return t.ListTimespanFn(ctx, scope, params)
}

// UpdateTimespan implements repository.TimespanRepository.
func (t *RepoMock) UpdateTimespan(ctx context.Context, scope model.OwnerScope, timespan model.Timespan) (model.Timespan, error) {
	return t.UpdateTimespanFn(ctx, scope, timespan)
}

// GetTotalDurationByTags implements repository.TimespanRepository.
func (t *RepoMock) GetTotalDurationByTags(ctx context.Context, scope model.OwnerScope, tagIds []uuid.UUID) (time.Duration, error) {
	return t.GetTotalDurationByTagsFn(ctx, scope, tagIds)
}

// AggregateTimeSpentByTagsAndBuckets implements repository.ProjectStatsRepository.
func (t *RepoMock) AggregateTimeSpentByTagsAndBuckets(ctx context.Context, scope model.OwnerScope, tagIds []uuid.UUID, buckets []model.BucketRange) ([]model.BucketValue, error) {
	return t.AggregateTimeSpentByTagsAndBucketsFn(ctx, scope, tagIds, buckets)
}

// CreateSession implements [SessionRepository].
func (s *SessionRepoMock) CreateSession(ctx context.Context, session model.Session) (model.Session, error) {
	return s.CreateSessionFn(ctx, session)
}

// DeleteSession implements [SessionRepository].
func (s *SessionRepoMock) DeleteSession(ctx context.Context, id uuid.UUID) error {
	return s.DeleteSessionFn(ctx, id)
}

// GetSession implements [SessionRepository].
func (s *SessionRepoMock) GetSession(ctx context.Context, id uuid.UUID) (model.Session, error) {
	return s.GetSessionFn(ctx, id)
}

// TouchSession implements [SessionRepository].
func (s *SessionRepoMock) TouchSession(ctx context.Context, id uuid.UUID, expiresAt time.Time) (model.Session, error) {
	return s.TouchSessionFn(ctx, id, expiresAt)
}

// DeleteAllExpiredSessions implements [SessionRepository].
func (s *SessionRepoMock) DeleteAllExpiredSessions(ctx context.Context) error {
	return s.DeleteAllExpiredSessionsFn(ctx)
}

// CreateLoginState implements [LoginStateRepository].
func (l *LoginStateRepoMock) CreateLoginState(ctx context.Context, loginState model.LoginState) (model.LoginState, error) {
	return l.CreateLoginStateFn(ctx, loginState)
}

// DeleteLoginState implements [LoginStateRepository].
func (l *LoginStateRepoMock) DeleteLoginState(ctx context.Context, id uuid.UUID) error {
	return l.DeleteLoginStateFn(ctx, id)
}

// GetLoginState implements [LoginStateRepository].
func (l *LoginStateRepoMock) GetLoginState(ctx context.Context, id uuid.UUID) (model.LoginState, error) {
	return l.GetLoginStateFn(ctx, id)
}

// DeleteAllExpiredLoginStates implements [LoginStateRepository].
func (l *LoginStateRepoMock) DeleteAllExpiredLoginStates(ctx context.Context) error {
	return l.DeleteAllExpiredLoginStatesFn(ctx)
}
