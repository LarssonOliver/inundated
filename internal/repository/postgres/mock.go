package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a testify mock implementing Repository.
// Use it in service-layer tests to stub the whole repository without touching SQL.
type MockRepository struct {
	mock.Mock
}

var _ repository.Repository = (*MockRepository)(nil)
var _ repository.SessionRepository = (*MockRepository)(nil)
var _ repository.LoginStateRepository = (*MockRepository)(nil)

func (m *MockRepository) CreateUser(ctx context.Context, user model.User) (model.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockRepository) GetUser(ctx context.Context, id uuid.UUID) (model.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockRepository) GetUserBySub(ctx context.Context, sub string) (model.User, error) {
	args := m.Called(ctx, sub)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockRepository) UpdateUser(ctx context.Context, user model.User) (model.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockRepository) GetTag(ctx context.Context, id uuid.UUID) (model.Tag, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(model.Tag), args.Error(1)
}

func (m *MockRepository) ListTags(ctx context.Context, params model.PaginationParams) (model.Page[model.Tag], error) {
	args := m.Called(ctx, params)
	return args.Get(0).(model.Page[model.Tag]), args.Error(1)
}

func (m *MockRepository) CreateTag(ctx context.Context, tag model.Tag) (model.Tag, error) {
	args := m.Called(ctx, tag)
	return args.Get(0).(model.Tag), args.Error(1)
}

func (m *MockRepository) UpdateTag(ctx context.Context, tag model.Tag) (model.Tag, error) {
	args := m.Called(ctx, tag)
	return args.Get(0).(model.Tag), args.Error(1)
}

func (m *MockRepository) DeleteTag(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockRepository) GetProject(ctx context.Context, id uuid.UUID) (model.Project, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(model.Project), args.Error(1)
}

func (m *MockRepository) ListProjects(ctx context.Context, params model.PaginationParams) (model.Page[model.Project], error) {
	args := m.Called(ctx, params)
	return args.Get(0).(model.Page[model.Project]), args.Error(1)
}

func (m *MockRepository) CreateProject(ctx context.Context, project model.Project) (model.Project, error) {
	args := m.Called(ctx, project)
	return args.Get(0).(model.Project), args.Error(1)
}

func (m *MockRepository) UpdateProject(ctx context.Context, project model.Project) (model.Project, error) {
	args := m.Called(ctx, project)
	return args.Get(0).(model.Project), args.Error(1)
}

func (m *MockRepository) DeleteProject(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockRepository) GetTimespan(ctx context.Context, id uuid.UUID) (model.Timespan, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(model.Timespan), args.Error(1)
}

func (m *MockRepository) ListTimespans(ctx context.Context, params model.PaginationParams) (model.Page[model.Timespan], error) {
	args := m.Called(ctx, params)
	return args.Get(0).(model.Page[model.Timespan]), args.Error(1)
}

func (m *MockRepository) CreateTimespan(ctx context.Context, timespan model.Timespan) (model.Timespan, error) {
	args := m.Called(ctx, timespan)
	return args.Get(0).(model.Timespan), args.Error(1)
}

func (m *MockRepository) UpdateTimespan(ctx context.Context, timespan model.Timespan) (model.Timespan, error) {
	args := m.Called(ctx, timespan)
	return args.Get(0).(model.Timespan), args.Error(1)
}

func (m *MockRepository) DeleteTimespan(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockRepository) GetTotalDurationByTags(ctx context.Context, tagIds []uuid.UUID) (time.Duration, error) {
	args := m.Called(ctx, tagIds)
	return args.Get(0).(time.Duration), args.Error(1)
}

func (m *MockRepository) AggregateTimeSpentByTagsAndBuckets(ctx context.Context, tagIds []uuid.UUID, buckets []model.BucketRange) ([]model.BucketValue, error) {
	args := m.Called(ctx, tagIds, buckets)
	return args.Get(0).([]model.BucketValue), args.Error(1)
}

// CreateSession implements [repository.SessionRepository].
func (m *MockRepository) CreateSession(ctx context.Context, session model.Session) (model.Session, error) {
	args := m.Called(ctx, session)
	return args.Get(0).(model.Session), args.Error(1)
}

// DeleteSession implements [repository.SessionRepository].
func (m *MockRepository) DeleteSession(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// GetSession implements [repository.SessionRepository].
func (m *MockRepository) GetSession(ctx context.Context, id uuid.UUID) (model.Session, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(model.Session), args.Error(1)
}

// TouchSession implements [repository.SessionRepository].
func (m *MockRepository) TouchSession(ctx context.Context, id uuid.UUID, expiresAt time.Time) (model.Session, error) {
	args := m.Called(ctx, id, expiresAt)
	return args.Get(0).(model.Session), args.Error(1)
}

// CreateLoginState implements [repository.LoginStateRepository].
func (m *MockRepository) CreateLoginState(ctx context.Context, loginState model.LoginState) (model.LoginState, error) {
	args := m.Called(ctx, loginState)
	return args.Get(0).(model.LoginState), args.Error(1)
}

// DeleteLoginState implements [repository.LoginStateRepository].
func (m *MockRepository) DeleteLoginState(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// GetLoginState implements [repository.LoginStateRepository].
func (m *MockRepository) GetLoginState(ctx context.Context, id uuid.UUID) (model.LoginState, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(model.LoginState), args.Error(1)
}
