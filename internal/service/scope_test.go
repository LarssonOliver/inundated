package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/service"
	"github.com/stretchr/testify/require"
)

// scopedMethod is one service entry point that must derive an owner scope from
// the request context and forward it to every repository call it makes.
type scopedMethod struct {
	name string
	// invoke calls the method; return values and errors are ignored — the test
	// only inspects the scopes the recording repo captured.
	invoke func(ctx context.Context, s *service.ServiceImpl)
	// repoCalls is the exact number of scoped repository calls the method makes.
	repoCalls int
}

func scopedMethods() []scopedMethod {
	validTimespan := func() model.Timespan {
		start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		return model.Timespan{
			Id:        uuid.New(),
			Name:      "ts",
			StartTime: start,
			EndTime:   start.Add(time.Hour),
		}
	}

	intervalRaw := "2024-01-01T00:00:00Z/2024-01-01T02:00:00Z"
	granularityRaw := "PT1H"
	timezoneRaw := "UTC"

	return []scopedMethod{
		{"GetTag", func(ctx context.Context, s *service.ServiceImpl) {
			_, _ = s.GetTag(ctx, uuid.New(), nil)
		}, 1},
		{"ListTags", func(ctx context.Context, s *service.ServiceImpl) {
			_, _ = s.ListTags(ctx, model.DefaultPaginationParams())
		}, 1},
		{"CreateTag", func(ctx context.Context, s *service.ServiceImpl) {
			_, _ = s.CreateTag(ctx, model.Tag{Name: "t", Color: "#abcdef"})
		}, 1},
		{"UpdateTag", func(ctx context.Context, s *service.ServiceImpl) {
			_, _ = s.UpdateTag(ctx, model.Tag{Id: uuid.New(), Name: "t", Color: "#abcdef"})
		}, 1},
		{"DeleteTag", func(ctx context.Context, s *service.ServiceImpl) {
			_ = s.DeleteTag(ctx, uuid.New())
		}, 1},

		{"GetProject", func(ctx context.Context, s *service.ServiceImpl) {
			_, _ = s.GetProject(ctx, uuid.New(), nil)
		}, 1},
		{"ListProjects", func(ctx context.Context, s *service.ServiceImpl) {
			_, _ = s.ListProjects(ctx, model.DefaultPaginationParams())
		}, 1},
		{"CreateProject", func(ctx context.Context, s *service.ServiceImpl) {
			_, _ = s.CreateProject(ctx, model.Project{Name: "p", Color: "#abcdef"})
		}, 1},
		{"UpdateProject", func(ctx context.Context, s *service.ServiceImpl) {
			_, _ = s.UpdateProject(ctx, model.Project{Id: uuid.New(), Name: "p", Color: "#abcdef"})
		}, 1},
		{"DeleteProject", func(ctx context.Context, s *service.ServiceImpl) {
			_ = s.DeleteProject(ctx, uuid.New())
		}, 1},

		{"GetTimespan", func(ctx context.Context, s *service.ServiceImpl) {
			_, _ = s.GetTimespan(ctx, uuid.New())
		}, 1},
		{"ListTimespans", func(ctx context.Context, s *service.ServiceImpl) {
			_, _ = s.ListTimespans(ctx, model.DefaultPaginationParams())
		}, 1},
		{"CreateTimespan", func(ctx context.Context, s *service.ServiceImpl) {
			_, _ = s.CreateTimespan(ctx, validTimespan())
		}, 1},
		{"UpdateTimespan", func(ctx context.Context, s *service.ServiceImpl) {
			_, _ = s.UpdateTimespan(ctx, validTimespan())
		}, 1},
		{"DeleteTimespan", func(ctx context.Context, s *service.ServiceImpl) {
			_ = s.DeleteTimespan(ctx, uuid.New())
		}, 1},

		// GetProjectStats must pass the SAME scope to both GetProject and
		// AggregateTimeSpentByTagsAndBuckets (spec Testing section).
		{"GetProjectStats", func(ctx context.Context, s *service.ServiceImpl) {
			_, _ = s.GetProjectStats(ctx, service.GetProjectStatsInput{
				ProjectID:      uuid.New(),
				Metric:         model.ProjectStatsMetricTimeSpent,
				IntervalRaw:    &intervalRaw,
				GranularityRaw: &granularityRaw,
				TimezoneRaw:    &timezoneRaw,
				Now:            time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			})
		}, 2},
	}
}

// recordingRepo returns a RepoMock whose every scoped closure appends the scope
// it received to rec and returns a minimal valid result.
func recordingRepo(rec *[]model.OwnerScope) *repository.RepoMock {
	record := func(scope model.OwnerScope) { *rec = append(*rec, scope) }

	return &repository.RepoMock{
		GetTagFn: func(_ context.Context, scope model.OwnerScope, id uuid.UUID) (model.Tag, error) {
			record(scope)
			return model.Tag{Id: id, Name: "t", Color: "#abcdef"}, nil
		},
		ListTagFn: func(_ context.Context, scope model.OwnerScope, _ model.PaginationParams) (model.Page[model.Tag], error) {
			record(scope)
			return model.Page[model.Tag]{}, nil
		},
		CreateTagFn: func(_ context.Context, scope model.OwnerScope, tag model.Tag) (model.Tag, error) {
			record(scope)
			return tag, nil
		},
		UpdateTagFn: func(_ context.Context, scope model.OwnerScope, tag model.Tag) (model.Tag, error) {
			record(scope)
			return tag, nil
		},
		DeleteTagFn: func(_ context.Context, scope model.OwnerScope, _ uuid.UUID) error {
			record(scope)
			return nil
		},

		GetProjectFn: func(_ context.Context, scope model.OwnerScope, id uuid.UUID) (model.Project, error) {
			record(scope)
			return model.Project{Id: id, Name: "p", Color: "#abcdef", TagIds: []uuid.UUID{uuid.New()}}, nil
		},
		ListProjectFn: func(_ context.Context, scope model.OwnerScope, _ model.PaginationParams) (model.Page[model.Project], error) {
			record(scope)
			return model.Page[model.Project]{}, nil
		},
		CreateProjectFn: func(_ context.Context, scope model.OwnerScope, project model.Project) (model.Project, error) {
			record(scope)
			return project, nil
		},
		UpdateProjectFn: func(_ context.Context, scope model.OwnerScope, project model.Project) (model.Project, error) {
			record(scope)
			return project, nil
		},
		DeleteProjectFn: func(_ context.Context, scope model.OwnerScope, _ uuid.UUID) error {
			record(scope)
			return nil
		},

		GetTimespanFn: func(_ context.Context, scope model.OwnerScope, id uuid.UUID) (model.Timespan, error) {
			record(scope)
			return model.Timespan{Id: id}, nil
		},
		ListTimespanFn: func(_ context.Context, scope model.OwnerScope, _ model.PaginationParams) (model.Page[model.Timespan], error) {
			record(scope)
			return model.Page[model.Timespan]{}, nil
		},
		CreateTimespanFn: func(_ context.Context, scope model.OwnerScope, timespan model.Timespan) (model.Timespan, error) {
			record(scope)
			return timespan, nil
		},
		UpdateTimespanFn: func(_ context.Context, scope model.OwnerScope, timespan model.Timespan) (model.Timespan, error) {
			record(scope)
			return timespan, nil
		},
		DeleteTimespanFn: func(_ context.Context, scope model.OwnerScope, _ uuid.UUID) error {
			record(scope)
			return nil
		},

		AggregateTimeSpentByTagsAndBucketsFn: func(_ context.Context, scope model.OwnerScope, _ []uuid.UUID, buckets []model.BucketRange) ([]model.BucketValue, error) {
			record(scope)
			out := make([]model.BucketValue, len(buckets))
			for i, b := range buckets {
				out[i] = model.BucketValue{Bucket: b, Value: 0}
			}
			return out, nil
		},
	}
}

// TestService_ScopedMethodsThreadContextScope proves that every scoped service
// method resolves the owner scope from the request context and forwards that
// exact scope to each repository call — a hardcoded scope or a stale variable in
// any one method would surface here rather than as a silent cross-user leak.
func TestService_ScopedMethodsThreadContextScope(t *testing.T) {
	userID := uuid.New()

	scenarios := []struct {
		name        string
		ctx         func() context.Context
		assertScope func(t *testing.T, scope model.OwnerScope)
	}{
		{
			name: "user in context resolves to that user's scope",
			ctx: func() context.Context {
				return model.SetUserInContext(context.Background(), model.User{Id: userID})
			},
			assertScope: func(t *testing.T, scope model.OwnerScope) {
				require.NotNil(t, scope.UserID())
				require.Equal(t, userID, *scope.UserID())
			},
		},
		{
			name: "no user in context resolves to the unowned scope",
			ctx: func() context.Context {
				return context.Background()
			},
			assertScope: func(t *testing.T, scope model.OwnerScope) {
				require.Nil(t, scope.UserID())
			},
		},
	}

	for _, sc := range scenarios {
		for _, m := range scopedMethods() {
			t.Run(sc.name+"/"+m.name, func(t *testing.T) {
				var rec []model.OwnerScope
				s := service.NewService(recordingRepo(&rec))

				m.invoke(sc.ctx(), s)

				require.Len(t, rec, m.repoCalls, "unexpected number of scoped repository calls")
				for i, scope := range rec {
					sc.assertScope(t, scope)
					require.Equal(t, rec[0], scope, "repository call %d received a different scope", i)
				}
			})
		}
	}
}
