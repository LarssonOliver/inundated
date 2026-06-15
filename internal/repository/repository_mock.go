package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

var _ Repository = (*RepoMock)(nil)

type RepoMock struct {
	CreateProjectFn func(ctx context.Context, project model.Project) (model.Project, error)
	DeleteProjectFn func(ctx context.Context, id uuid.UUID) error
	GetProjectFn    func(ctx context.Context, id uuid.UUID) (model.Project, error)
	ListProjectFn   func(ctx context.Context, params model.PaginationParams) (model.Page[model.Project], error)
	UpdateProjectFn func(ctx context.Context, project model.Project) (model.Project, error)

	CreateTagFn func(ctx context.Context, tag model.Tag) (model.Tag, error)
	DeleteTagFn func(ctx context.Context, id uuid.UUID) error
	GetTagFn    func(ctx context.Context, id uuid.UUID) (model.Tag, error)
	ListTagFn   func(ctx context.Context, params model.PaginationParams) (model.Page[model.Tag], error)
	UpdateTagFn func(ctx context.Context, tag model.Tag) (model.Tag, error)

	GetUserByIDFn   func(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetUserBySubFn  func(ctx context.Context, sub string) (*model.User, error)
	CreateUserFn    func(ctx context.Context, user *model.User) error
	UpdateUserFn    func(ctx context.Context, id uuid.UUID, update *model.UpdateUser) (*model.User, error)

	CreateTimespanFn         func(ctx context.Context, timespan model.Timespan) (model.Timespan, error)
	DeleteTimespanFn         func(ctx context.Context, id uuid.UUID) error
	GetTimespanFn            func(ctx context.Context, id uuid.UUID) (model.Timespan, error)
	ListTimespanFn           func(ctx context.Context, params model.PaginationParams) (model.Page[model.Timespan], error)
	UpdateTimespanFn         func(ctx context.Context, timespan model.Timespan) (model.Timespan, error)
	GetTotalDurationByTagsFn func(ctx context.Context, tagIds []uuid.UUID) (time.Duration, error)

	AggregateTimeSpentByTagsAndBucketsFn func(ctx context.Context, tagIds []uuid.UUID, buckets []model.BucketRange) ([]model.BucketValue, error)
}

// CreateProject implements repository.ProjectRepository.
func (t *RepoMock) CreateProject(ctx context.Context, project model.Project) (model.Project, error) {
	return t.CreateProjectFn(ctx, project)
}

// DeleteProject implements repository.ProjectRepository.
func (t *RepoMock) DeleteProject(ctx context.Context, id uuid.UUID) error {
	return t.DeleteProjectFn(ctx, id)
}

// GetProject implements repository.ProjectRepository.
func (t *RepoMock) GetProject(ctx context.Context, id uuid.UUID) (model.Project, error) {
	return t.GetProjectFn(ctx, id)
}

// ListProjects implements repository.ProjectRepository.
func (t *RepoMock) ListProjects(ctx context.Context, params model.PaginationParams) (model.Page[model.Project], error) {
	return t.ListProjectFn(ctx, params)
}

// UpdateProject implements repository.ProjectRepository.
func (t *RepoMock) UpdateProject(ctx context.Context, project model.Project) (model.Project, error) {
	return t.UpdateProjectFn(ctx, project)
}

// CreateTag implements repository.TagRepository.
func (t *RepoMock) CreateTag(ctx context.Context, tag model.Tag) (model.Tag, error) {
	return t.CreateTagFn(ctx, tag)
}

// DeleteTag implements repository.TagRepository.
func (t *RepoMock) DeleteTag(ctx context.Context, id uuid.UUID) error {
	return t.DeleteTagFn(ctx, id)
}

// GetTag implements repository.TagRepository.
func (t *RepoMock) GetTag(ctx context.Context, id uuid.UUID) (model.Tag, error) {
	return t.GetTagFn(ctx, id)
}

// ListTags implements repository.TagRepository.
func (t *RepoMock) ListTags(ctx context.Context, params model.PaginationParams) (model.Page[model.Tag], error) {
	return t.ListTagFn(ctx, params)
}

// UpdateTag implements repository.TagRepository.
func (t *RepoMock) UpdateTag(ctx context.Context, tag model.Tag) (model.Tag, error) {
	return t.UpdateTagFn(ctx, tag)
}

// GetByID implements repository.UserRepository.
func (t *RepoMock) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return t.GetUserByIDFn(ctx, id)
}

// GetBySub implements repository.UserRepository.
func (t *RepoMock) GetBySub(ctx context.Context, sub string) (*model.User, error) {
	return t.GetUserBySubFn(ctx, sub)
}

// Create implements repository.UserRepository.
func (t *RepoMock) Create(ctx context.Context, user *model.User) error {
	return t.CreateUserFn(ctx, user)
}

// Update implements repository.UserRepository.
func (t *RepoMock) Update(ctx context.Context, id uuid.UUID, update *model.UpdateUser) (*model.User, error) {
	return t.UpdateUserFn(ctx, id, update)
}

// CreateTimespan implements repository.TimespanRepository.
func (t *RepoMock) CreateTimespan(ctx context.Context, timespan model.Timespan) (model.Timespan, error) {
	return t.CreateTimespanFn(ctx, timespan)
}

// DeleteTimespan implements repository.TimespanRepository.
func (t *RepoMock) DeleteTimespan(ctx context.Context, id uuid.UUID) error {
	return t.DeleteTimespanFn(ctx, id)
}

// GetTimespan implements repository.TimespanRepository.
func (t *RepoMock) GetTimespan(ctx context.Context, id uuid.UUID) (model.Timespan, error) {
	return t.GetTimespanFn(ctx, id)
}

// ListTimespans implements repository.TimespanRepository.
func (t *RepoMock) ListTimespans(ctx context.Context, params model.PaginationParams) (model.Page[model.Timespan], error) {
	return t.ListTimespanFn(ctx, params)
}

// UpdateTimespan implements repository.TimespanRepository.
func (t *RepoMock) UpdateTimespan(ctx context.Context, timespan model.Timespan) (model.Timespan, error) {
	return t.UpdateTimespanFn(ctx, timespan)
}

// GetTotalDurationByTags implements repository.TimespanRepository.
func (t *RepoMock) GetTotalDurationByTags(ctx context.Context, tagIds []uuid.UUID) (time.Duration, error) {
	return t.GetTotalDurationByTagsFn(ctx, tagIds)
}

// AggregateTimeSpentByTagsAndBuckets implements repository.ProjectStatsRepository.
func (t *RepoMock) AggregateTimeSpentByTagsAndBuckets(ctx context.Context, tagIds []uuid.UUID, buckets []model.BucketRange) ([]model.BucketValue, error) {
	return t.AggregateTimeSpentByTagsAndBucketsFn(ctx, tagIds, buckets)
}
