package postgres

import (
	"context"

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

var _ repository.Repository= (*MockRepository)(nil)

func (m *MockRepository) GetTag(ctx context.Context, id uuid.UUID) (model.Tag, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(model.Tag), args.Error(1)
}

func (m *MockRepository) ListTags(ctx context.Context) ([]model.Tag, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Tag), args.Error(1)
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

func (m *MockRepository) ListProjects(ctx context.Context) ([]model.Project, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Project), args.Error(1)
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

func (m *MockRepository) GetTimeSpan(ctx context.Context, id uuid.UUID) (model.TimeSpan, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(model.TimeSpan), args.Error(1)
}

func (m *MockRepository) ListTimeSpans(ctx context.Context) ([]model.TimeSpan, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.TimeSpan), args.Error(1)
}

func (m *MockRepository) CreateTimeSpan(ctx context.Context, timeSpan model.TimeSpan) (model.TimeSpan, error) {
	args := m.Called(ctx, timeSpan)
	return args.Get(0).(model.TimeSpan), args.Error(1)
}

func (m *MockRepository) UpdateTimeSpan(ctx context.Context, timeSpan model.TimeSpan) (model.TimeSpan, error) {
	args := m.Called(ctx, timeSpan)
	return args.Get(0).(model.TimeSpan), args.Error(1)
}

func (m *MockRepository) DeleteTimeSpan(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
