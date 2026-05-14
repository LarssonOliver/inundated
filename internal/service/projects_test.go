package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/service"
	"github.com/stretchr/testify/require"
)

func TestProjectService_GetProject(t *testing.T) {
	testId := uuid.New()

	tests := []struct {
		name        string
		id          uuid.UUID
		getFn       func(ctx context.Context, id uuid.UUID) (model.Project, error)
		includes    *service.ProjectServiceGetIncludes
		totalTimeFn func(ctx context.Context, ids []uuid.UUID) (time.Duration, error)
		want        model.Project
		wantErr     bool
	}{
		{
			name: "successful get",
			id:   testId,
			getFn: func(ctx context.Context, id uuid.UUID) (model.Project, error) {
				return model.Project{Id: id, Name: "Test Project", Color: "#abcdef"}, nil
			},
			want:    model.Project{Id: testId, Name: "Test Project", Color: "#abcdef"},
			wantErr: false,
		},
		{
			name: "repository error",
			id:   testId,
			getFn: func(ctx context.Context, id uuid.UUID) (model.Project, error) {
				return model.Project{}, errors.New("not found")
			},
			want:    model.Project{},
			wantErr: true,
		},
		{
			name: "Include total time",
			id:   testId,
			getFn: func(ctx context.Context, id uuid.UUID) (model.Project, error) {
				return model.Project{Id: id, Name: "Test Project", Color: "#abcdef", TagIds: []uuid.UUID{uuid.New()}}, nil
			},
			includes: &service.ProjectServiceGetIncludes{TotalTime: true},
			totalTimeFn: func(ctx context.Context, ids []uuid.UUID) (time.Duration, error) {
				return 2 * time.Hour, nil
			},
			want:    model.Project{Id: testId, Name: "Test Project", Color: "#abcdef", TotalTime: func() *time.Duration { d := 2 * time.Hour; return &d }()},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repository.RepoMock{
				GetProjectFn: tt.getFn,
			}
			if tt.totalTimeFn != nil {
				repo.GetTotalDurationByTagsFn = tt.totalTimeFn
			}

			s := service.NewService(repo)
			got, gotErr := s.GetProject(context.Background(), tt.id, tt.includes)
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			require.NotEqual(t, uuid.Nil, got.Id)
			require.Equal(t, tt.want.Id, got.Id)
			require.Equal(t, tt.want.Name, got.Name)
			require.Equal(t, tt.want.Color, got.Color)

			if tt.includes != nil && tt.includes.TotalTime {
				require.NotNil(t, got.TotalTime)
				require.Equal(t, *tt.want.TotalTime, *got.TotalTime)
			} else {
				require.Nil(t, got.TotalTime)
			}
		})
	}
}

func TestProjectService_ListProjects(t *testing.T) {
	projects := []model.Project{
		{Id: uuid.New(), Name: "Project1", Color: "#ff0000"},
		{Id: uuid.New(), Name: "Project2", Color: "#00ff00"},
	}

	tests := []struct {
		name    string
		listFn  func(ctx context.Context) ([]model.Project, error)
		want    []model.Project
		wantErr bool
	}{
		{
			name: "successful list",
			listFn: func(ctx context.Context) ([]model.Project, error) {
				return projects, nil
			},
			want:    projects,
			wantErr: false,
		},
		{
			name: "repository error",
			listFn: func(ctx context.Context) ([]model.Project, error) {
				return nil, errors.New("database error")
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repository.RepoMock{
				ListProjectFn: tt.listFn,
			}

			s := service.NewService(repo)
			got, gotErr := s.ListProjects(context.Background())
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			require.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestProjectService_CreateProject(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		name     string
		project  model.Project
		createFn func(ctx context.Context, project model.Project) (model.Project, error)
		want     model.Project
		wantErr  bool
	}{
		{
			name:    "successful create",
			project: model.Project{Name: "New Project", Color: "#123456"},
			createFn: func(ctx context.Context, project model.Project) (model.Project, error) {
				project.Id = uuid.New()
				return project, nil
			},
			want:    model.Project{Name: "New Project", Color: "#123456"},
			wantErr: false,
		},
		{
			name:    "ensure create generates Id",
			project: model.Project{Id: id, Name: "New Project", Color: "#123456"},
			createFn: func(ctx context.Context, project model.Project) (model.Project, error) {
				return project, nil
			},
			want:    model.Project{Id: id, Name: "New Project", Color: "#123456"},
			wantErr: false,
		},
		{
			name:    "repository error",
			project: model.Project{Name: "New Project", Color: "#123456"},
			createFn: func(ctx context.Context, project model.Project) (model.Project, error) {
				return model.Project{}, errors.New("database error")
			},
			want:    model.Project{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repository.RepoMock{
				CreateProjectFn: tt.createFn,
			}
			s := service.NewService(repo)
			got, gotErr := s.CreateProject(context.Background(), tt.project)
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			require.NotEqual(t, uuid.Nil, got.Id)
			require.NotEqual(t, tt.want.Id, got.Id)
			require.Equal(t, tt.want.Name, got.Name)
			require.Equal(t, tt.want.Color, got.Color)
		})
	}
}

func TestProjectService_UpdateProject(t *testing.T) {
	projectId := uuid.New()
	tests := []struct {
		name     string
		project  model.Project
		updateFn func(ctx context.Context, project model.Project) (model.Project, error)
		want     model.Project
		wantErr  bool
	}{
		{
			name:    "successful update",
			project: model.Project{Id: projectId, Name: "Updated Project", Color: "#654321"},
			updateFn: func(ctx context.Context, project model.Project) (model.Project, error) {
				return project, nil
			},
			want:    model.Project{Id: projectId, Name: "Updated Project", Color: "#654321"},
			wantErr: false,
		},
		{
			name:    "repository error",
			project: model.Project{Id: projectId, Name: "Updated Project", Color: "#654321"},
			updateFn: func(ctx context.Context, project model.Project) (model.Project, error) {
				return model.Project{}, errors.New("database error")
			},
			want:    model.Project{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repository.RepoMock{
				UpdateProjectFn: tt.updateFn,
			}
			s := service.NewService(repo)
			got, gotErr := s.UpdateProject(context.Background(), tt.project)
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			require.NotEqual(t, uuid.Nil, got.Id)
			require.Equal(t, tt.want.Id, got.Id)
			require.Equal(t, tt.want.Name, got.Name)
			require.Equal(t, tt.want.Color, got.Color)
		})
	}
}

func TestProjectService_DeleteProject(t *testing.T) {
	tests := []struct {
		name     string
		deleteFn func(ctx context.Context, id uuid.UUID) error
		wantErr  bool
	}{
		{
			name: "successful delete",
			deleteFn: func(ctx context.Context, id uuid.UUID) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "repository error",
			deleteFn: func(ctx context.Context, id uuid.UUID) error {
				return errors.New("database error")
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repository.RepoMock{
				DeleteProjectFn: tt.deleteFn,
			}
			s := service.NewService(repo)
			gotErr := s.DeleteProject(context.Background(), uuid.New())
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}

func TestProjectService_GetProjectStats(t *testing.T) {
	projectID := uuid.New()
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	intervalRaw := "2024-01-01T00:00:00Z/2024-01-01T02:00:00Z"
	granularityRaw := "PT1H"
	timezoneRaw := "UTC"

	t.Run("successful get project stats", func(t *testing.T) {
		repo := &repository.RepoMock{
			GetProjectFn: func(ctx context.Context, id uuid.UUID) (model.Project, error) {
				require.Equal(t, projectID, id)
				return model.Project{
					Id:     projectID,
					Name:   "Project A",
					Color:  "#123456",
					TagIds: []uuid.UUID{uuid.New()},
				}, nil
			},
			AggregateTimeSpentByTagsAndBucketsFn: func(ctx context.Context, tagIds []uuid.UUID, buckets []model.BucketRange) ([]model.BucketValue, error) {
				require.NotEmpty(t, tagIds)
				require.Len(t, buckets, 2)
				require.True(t, buckets[0].Start.Equal(base))
				require.True(t, buckets[0].End.Equal(base.Add(1*time.Hour)))
				require.True(t, buckets[1].Start.Equal(base.Add(1*time.Hour)))
				require.True(t, buckets[1].End.Equal(base.Add(2*time.Hour)))

				return []model.BucketValue{
					{Bucket: buckets[0], Value: 1800},
					{Bucket: buckets[1], Value: 3600},
				}, nil
			},
		}

		s := service.NewService(repo)
		got, err := s.GetProjectStats(context.Background(), service.GetProjectStatsInput{
			ProjectID:      projectID,
			Metric:         model.ProjectStatsMetricTimeSpent,
			IntervalRaw:    &intervalRaw,
			GranularityRaw: &granularityRaw,
			TimezoneRaw:    &timezoneRaw,
			Now:            base.Add(10 * time.Hour),
		})

		require.NoError(t, err)
		require.Equal(t, projectID, got.ProjectID)
		require.Equal(t, model.ProjectStatsMetricTimeSpent, got.Metric)
		require.True(t, got.Interval.Start.Equal(base))
		require.True(t, got.Interval.End.Equal(base.Add(2*time.Hour)))
		require.Equal(t, "PT1H", got.Granularity)
		require.Equal(t, "seconds", got.Unit)
		require.Len(t, got.Series, 2)
		require.InDelta(t, 1800, got.Series[0].Value, 0.0001)
		require.InDelta(t, 3600, got.Series[1].Value, 0.0001)
	})

	t.Run("invalid metric", func(t *testing.T) {
		repo := &repository.RepoMock{}
		s := service.NewService(repo)

		_, err := s.GetProjectStats(context.Background(), service.GetProjectStatsInput{
			ProjectID: projectID,
			Metric:    model.ProjectStatsMetric("invalid_metric"),
		})

		require.ErrorIs(t, err, model.ErrInvalidArgument)
	})

	t.Run("project not found", func(t *testing.T) {
		repo := &repository.RepoMock{
			GetProjectFn: func(ctx context.Context, id uuid.UUID) (model.Project, error) {
				return model.Project{}, errors.New("db error")
			},
		}
		s := service.NewService(repo)

		_, err := s.GetProjectStats(context.Background(), service.GetProjectStatsInput{
			ProjectID: projectID,
			Metric:    model.ProjectStatsMetricTimeSpent,
		})

		require.ErrorIs(t, err, model.ErrNotFound)
	})

	t.Run("invalid interval", func(t *testing.T) {
		repo := &repository.RepoMock{
			GetProjectFn: func(ctx context.Context, id uuid.UUID) (model.Project, error) {
				return model.Project{Id: projectID}, nil
			},
		}
		s := service.NewService(repo)
		badInterval := "foo/bar"

		_, err := s.GetProjectStats(context.Background(), service.GetProjectStatsInput{
			ProjectID:   projectID,
			Metric:      model.ProjectStatsMetricTimeSpent,
			IntervalRaw: &badInterval,
		})

		require.ErrorIs(t, err, model.ErrInvalidArgument)
	})

	t.Run("unprocessable interval", func(t *testing.T) {
		repo := &repository.RepoMock{
			GetProjectFn: func(ctx context.Context, id uuid.UUID) (model.Project, error) {
				return model.Project{
					Id:     projectID,
					TagIds: []uuid.UUID{uuid.New()},
				}, nil
			},
		}
		s := service.NewService(repo)
		unprocessableInterval := "2024-01-01T00:00:00Z/2024-01-01T00:00:00Z"

		_, err := s.GetProjectStats(context.Background(), service.GetProjectStatsInput{
			ProjectID:   projectID,
			Metric:      model.ProjectStatsMetricTimeSpent,
			IntervalRaw: &unprocessableInterval,
		})

		require.ErrorIs(t, err, model.ErrUnprocessable)
	})
}
