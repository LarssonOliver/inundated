package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

type TimespanServiceMock struct {
	CreateFn func(ctx context.Context, timespan model.Timespan) (model.Timespan, error)
	DeleteFn func(ctx context.Context, id uuid.UUID) error
	GetFn    func(ctx context.Context, id uuid.UUID) (model.Timespan, error)
	ListFn   func(ctx context.Context, params model.PaginationParams) (model.Page[model.Timespan], error)
	UpdateFn func(ctx context.Context, timespan model.Timespan) (model.Timespan, error)
}

var _ TimespanService = (*TimespanServiceMock)(nil)

// CreateTimespan implements [service.TimespanService].
func (m *TimespanServiceMock) CreateTimespan(ctx context.Context, timespan model.Timespan) (model.Timespan, error) {
	return m.CreateFn(ctx, timespan)
}

// DeleteTimespan implements [service.TimespanService].
func (m *TimespanServiceMock) DeleteTimespan(ctx context.Context, id uuid.UUID) error {
	return m.DeleteFn(ctx, id)
}

// GetTimespan implements [service.TimespanService].
func (m *TimespanServiceMock) GetTimespan(ctx context.Context, id uuid.UUID) (model.Timespan, error) {
	return m.GetFn(ctx, id)
}

// ListTimespans implements [service.TimespanService].
func (m *TimespanServiceMock) ListTimespans(ctx context.Context, params model.PaginationParams) (model.Page[model.Timespan], error) {
	return m.ListFn(ctx, params)
}

// UpdateTimespan implements [service.TimespanService].
func (m *TimespanServiceMock) UpdateTimespan(ctx context.Context, timespan model.Timespan) (model.Timespan, error) {
	return m.UpdateFn(ctx, timespan)
}

type TagServiceMock struct {
	CreateFn func(ctx context.Context, tag model.Tag) (model.Tag, error)
	DeleteFn func(ctx context.Context, id uuid.UUID) error
	GetFn    func(ctx context.Context, id uuid.UUID, includes *TagServiceGetIncludes) (model.Tag, error)
	ListFn   func(ctx context.Context, params model.PaginationParams) (model.Page[model.Tag], error)
	UpdateFn func(ctx context.Context, tag model.Tag) (model.Tag, error)
}

var _ TagService = (*TagServiceMock)(nil)

// CreateTag implements [service.TagService].
func (m *TagServiceMock) CreateTag(ctx context.Context, tag model.Tag) (model.Tag, error) {
	return m.CreateFn(ctx, tag)
}

// DeleteTag implements [service.TagService].
func (m *TagServiceMock) DeleteTag(ctx context.Context, id uuid.UUID) error {
	return m.DeleteFn(ctx, id)
}

// GetTag implements [service.TagService].
func (m *TagServiceMock) GetTag(ctx context.Context, id uuid.UUID, includes *TagServiceGetIncludes) (model.Tag, error) {
	return m.GetFn(ctx, id, includes)
}

// ListTags implements [service.TagService].
func (m *TagServiceMock) ListTags(ctx context.Context, params model.PaginationParams) (model.Page[model.Tag], error) {
	return m.ListFn(ctx, params)
}

// UpdateTag implements [service.TagService].
func (m *TagServiceMock) UpdateTag(ctx context.Context, tag model.Tag) (model.Tag, error) {
	return m.UpdateFn(ctx, tag)
}

type ProjectServiceMock struct {
	CreateFn   func(ctx context.Context, project model.Project) (model.Project, error)
	DeleteFn   func(ctx context.Context, id uuid.UUID) error
	GetFn      func(ctx context.Context, id uuid.UUID, i *ProjectServiceGetIncludes) (model.Project, error)
	GetStatsFn func(ctx context.Context, input GetProjectStatsInput) (model.ProjectStats, error)
	ListFn     func(ctx context.Context, params model.PaginationParams) (model.Page[model.Project], error)
	UpdateFn   func(ctx context.Context, project model.Project) (model.Project, error)
}

var _ ProjectService = (*ProjectServiceMock)(nil)

// CreateProject implements [service.ProjectService].
func (m *ProjectServiceMock) CreateProject(ctx context.Context, project model.Project) (model.Project, error) {
	return m.CreateFn(ctx, project)
}

// DeleteProject implements [service.ProjectService].
func (m *ProjectServiceMock) DeleteProject(ctx context.Context, id uuid.UUID) error {
	return m.DeleteFn(ctx, id)
}

// GetProject implements [service.ProjectService].
func (m *ProjectServiceMock) GetProject(ctx context.Context, id uuid.UUID, i *ProjectServiceGetIncludes) (model.Project, error) {
	return m.GetFn(ctx, id, i)
}

// ListProjects implements [service.ProjectService].
func (m *ProjectServiceMock) ListProjects(ctx context.Context, params model.PaginationParams) (model.Page[model.Project], error) {
	return m.ListFn(ctx, params)
}

// GetProjectStats implements [service.ProjectService].
func (m *ProjectServiceMock) GetProjectStats(ctx context.Context, input GetProjectStatsInput) (model.ProjectStats, error) {
	if m.GetStatsFn == nil {
		return model.ProjectStats{}, model.ErrNotImplemented
	}
	return m.GetStatsFn(ctx, input)
}

// UpdateProject implements [service.ProjectService].
func (m *ProjectServiceMock) UpdateProject(ctx context.Context, project model.Project) (model.Project, error) {
	return m.UpdateFn(ctx, project)
}

type UserServiceMock struct {
	GetCurrentUserFn       func(ctx context.Context) (model.User, error)
	UpdateCurrentUserFn    func(ctx context.Context, user model.User) (model.User, error)
	GetOrCreateUserBySubFn func(ctx context.Context, subject string) (model.User, error)
}

var _ UserService = (*UserServiceMock)(nil)

// GetCurrentUser implements [UserService].
func (u *UserServiceMock) GetCurrentUser(ctx context.Context) (model.User, error) {
	return u.GetCurrentUserFn(ctx)
}

// GetOrCreateUserBySub implements [UserService].
func (u *UserServiceMock) GetOrCreateUserBySub(ctx context.Context, subject string) (model.User, error) {
	return u.GetOrCreateUserBySubFn(ctx, subject)
}

// UpdateCurrentUser implements [UserService].
func (u *UserServiceMock) UpdateCurrentUser(ctx context.Context, user model.User) (model.User, error) {
	return u.UpdateCurrentUserFn(ctx, user)
}

