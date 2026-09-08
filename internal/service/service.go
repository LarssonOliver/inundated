package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
)

type ServiceImpl struct {
	repository           repository.Repository
	registrationDisabled bool
}

var _ Service = (*ServiceImpl)(nil)

// ServiceOption tweaks a ServiceImpl at construction time.
type ServiceOption func(*ServiceImpl)

// WithRegistrationDisabled makes GetOrCreateUserByIdentity reject an OIDC
// identity that doesn't already map to a user, instead of creating one.
// Existing users are unaffected.
func WithRegistrationDisabled(disabled bool) ServiceOption {
	return func(s *ServiceImpl) { s.registrationDisabled = disabled }
}

func NewService(repository repository.Repository, opts ...ServiceOption) *ServiceImpl {
	s := &ServiceImpl{
		repository: repository,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

type Service interface {
	UserService
	TagService
	ProjectService
	TimespanService
}

type UserService interface {
	GetCurrentUser(ctx context.Context) (model.User, error)
	GetUserBySub(ctx context.Context, sub string) (model.User, error)
	GetOrCreateUserByIdentity(ctx context.Context, identity model.UserIdentity) (model.User, error)
}

type TagServiceGetIncludes struct {
	TotalTime bool
}

type TagService interface {
	GetTag(ctx context.Context, id uuid.UUID, includes *TagServiceGetIncludes) (model.Tag, error)
	ListTags(ctx context.Context, params model.PaginationParams) (model.Page[model.Tag], error)
	CreateTag(ctx context.Context, tag model.Tag) (model.Tag, error)
	UpdateTag(ctx context.Context, tag model.Tag) (model.Tag, error)
	DeleteTag(ctx context.Context, id uuid.UUID) error
}

type ProjectServiceGetIncludes struct {
	TotalTime bool
}

type ProjectService interface {
	GetProject(ctx context.Context, id uuid.UUID, includes *ProjectServiceGetIncludes) (model.Project, error)
	ListProjects(ctx context.Context, params model.PaginationParams) (model.Page[model.Project], error)
	CreateProject(ctx context.Context, project model.Project) (model.Project, error)
	UpdateProject(ctx context.Context, project model.Project) (model.Project, error)
	DeleteProject(ctx context.Context, id uuid.UUID) error
	GetProjectStats(ctx context.Context, input GetProjectStatsInput) (model.ProjectStats, error)
}

type GetProjectStatsInput struct {
	ProjectID      uuid.UUID
	Metric         model.ProjectStatsMetric
	IntervalRaw    *string
	GranularityRaw *string
	TimezoneRaw    *string
	Now            time.Time
}

type TimespanService interface {
	GetTimespan(ctx context.Context, id uuid.UUID) (model.Timespan, error)
	ListTimespans(ctx context.Context, params model.PaginationParams) (model.Page[model.Timespan], error)
	CreateTimespan(ctx context.Context, timespan model.Timespan) (model.Timespan, error)
	UpdateTimespan(ctx context.Context, timespan model.Timespan) (model.Timespan, error)
	DeleteTimespan(ctx context.Context, id uuid.UUID) error
}
