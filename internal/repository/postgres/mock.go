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

func (m *MockRepository) CreateUser(ctx context.Context, user model.User) error {
	panic("unimplemented")
}

func (m *MockRepository) GetUser(ctx context.Context, id uuid.UUID) (model.User, error) {
	panic("unimplemented")
}

func (m *MockRepository) GetUserBySub(ctx context.Context, sub string) (model.User, error) {
	panic("unimplemented")
}

func (m *MockRepository) UpdateUser(ctx context.Context, user model.User) (model.User, error) {
	panic("unimplemented")
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
