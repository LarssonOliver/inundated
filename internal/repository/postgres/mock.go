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

var _ repository.Repository = (*MockRepository)(nil)

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

func (m *MockRepository) GetTimespan(ctx context.Context, id uuid.UUID) (model.Timespan, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(model.Timespan), args.Error(1)
}

func (m *MockRepository) ListTimespans(ctx context.Context) ([]model.Timespan, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Timespan), args.Error(1)
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
