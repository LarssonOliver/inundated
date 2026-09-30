package contract_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/repository/memory"
	"github.com/larssonoliver/inundated/internal/repository/postgres"
	"github.com/larssonoliver/inundated/test/testutils"
	"github.com/stretchr/testify/require"
)

// TestProjectTagContract covers the project tag every project owns: shown
// as "@name", following its project, and counting toward it wherever a
// timespan or task carries it.
func TestProjectTagContract(t *testing.T) {
	ctx := context.Background()
	pagination := model.DefaultPaginationParams()

	run := func(t *testing.T, repoName string, newRepo func(t *testing.T) repository.Repository) {
		t.Run(repoName+"ProjectTagFollowsProject", func(t *testing.T) {
			repo := newRepo(t)

			project, err := repo.CreateProject(ctx, testScope, model.Project{Name: "Website", Color: "#bf616a"})
			require.NoError(t, err)
			require.NotEqual(t, uuid.Nil, project.TagId)
			require.Empty(t, project.TagIds, "the project tag is never among the linked tags")

			tag, err := repo.GetTag(ctx, testScope, project.TagId)
			require.NoError(t, err)
			require.Equal(t, "Website", tag.Name)
			require.Equal(t, "#bf616a", tag.Color)
			require.Equal(t, &model.TagOwner{Kind: model.TagOwnerProject, Id: project.Id}, tag.Owner)
			require.False(t, tag.IsDerived())

			got, err := repo.GetProject(ctx, testScope, project.Id)
			require.NoError(t, err)
			require.Equal(t, project.TagId, got.TagId)
			require.Empty(t, got.TagIds)

			// Project tags list under their own kind, not as labels.
			projects, err := repo.ListTags(ctx, testScope, model.TagListParams{PaginationParams: pagination, Kind: model.TagKindProject})
			require.NoError(t, err)
			require.Len(t, projects.Data, 1)
			require.Equal(t, project.TagId, projects.Data[0].Id)
			labels, err := repo.ListTags(ctx, testScope, model.TagListParams{PaginationParams: pagination})
			require.NoError(t, err)
			require.Empty(t, labels.Data)

			// Renaming, recoloring and archiving the project carry over.
			project.Name, project.Color, project.Archived = "Web site", "#a3be8c", true
			_, err = repo.UpdateProject(ctx, testScope, project)
			require.NoError(t, err)
			tag, err = repo.GetTag(ctx, testScope, project.TagId)
			require.NoError(t, err)
			require.Equal(t, "Web site", tag.Name)
			require.Equal(t, "#a3be8c", tag.Color)
			require.True(t, tag.Archived)

			// It can only change through its project.
			_, err = repo.UpdateTag(ctx, testScope, model.Tag{Id: tag.Id, Name: "x", Color: "#000000"})
			require.ErrorIs(t, err, model.ErrInvalidArgument)
			require.ErrorIs(t, repo.DeleteTag(ctx, testScope, tag.Id), model.ErrInvalidArgument)

			// Deleting the project deletes its tag.
			require.NoError(t, repo.DeleteProject(ctx, testScope, project.Id))
			_, err = repo.GetTag(ctx, testScope, project.TagId)
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"ProjectTagsCountTowardTheirProject", func(t *testing.T) {
			repo := newRepo(t)

			website, err := repo.CreateProject(ctx, testScope, model.Project{Name: "Website", Color: "#bf616a"})
			require.NoError(t, err)
			parent, err := repo.CreateTask(ctx, testScope, model.Task{Name: "Launch", TagIds: []uuid.UUID{website.TagId}})
			require.NoError(t, err)
			child, err := repo.CreateTask(ctx, testScope, model.Task{Name: "Hero", ParentId: &parent.Id})
			require.NoError(t, err)

			start := time.Date(2026, 9, 1, 9, 0, 0, 0, time.UTC)
			logTime := func(d time.Duration, tagId uuid.UUID) uuid.UUID {
				span, err := repo.CreateTimespan(ctx, testScope, model.Timespan{
					StartTime: start, EndTime: start.Add(d), TagIds: []uuid.UUID{tagId},
				})
				require.NoError(t, err)
				start = start.Add(d)
				return span.Id
			}
			other, err := repo.CreateTag(ctx, testScope, model.Tag{Name: "other", Color: "#88c0d0"})
			require.NoError(t, err)
			direct := logTime(time.Hour, website.TagId) // straight on the project
			onTask := logTime(2*time.Hour, child.TagId) // on a subtask of a task carrying @Website
			logTime(4*time.Hour, other.Id)              // not the project's

			total, err := repo.GetTotalDurationByTags(ctx, testScope, website.EffectiveTagIds())
			require.NoError(t, err)
			require.Equal(t, 3*time.Hour, total)
			onTasks, err := repo.GetTaskDurationByTags(ctx, testScope, website.EffectiveTagIds())
			require.NoError(t, err)
			require.Equal(t, 2*time.Hour, onTasks)

			// The task and its subtask belong to the project.
			ids, err := repo.ListTaskProjectIds(ctx, testScope, []uuid.UUID{parent.Id, child.Id})
			require.NoError(t, err)
			require.Equal(t, []uuid.UUID{website.Id}, ids[parent.Id])
			require.Equal(t, []uuid.UUID{website.Id}, ids[child.Id])
			page, err := repo.ListTasks(ctx, testScope, model.TaskListParams{PaginationParams: pagination, ProjectId: &website.Id})
			require.NoError(t, err)
			require.Len(t, page.Data, 2)

			// Projects can't carry project tags, their own included.
			_, err = repo.CreateProject(ctx, testScope, model.Project{Name: "Shop", Color: "#ebcb8b", TagIds: []uuid.UUID{website.TagId}})
			require.ErrorIs(t, err, model.ErrInvalidReference)
			website.TagIds = []uuid.UUID{website.TagId}
			_, err = repo.UpdateProject(ctx, testScope, website)
			require.ErrorIs(t, err, model.ErrInvalidReference)

			// A project with time attributed to it can't be deleted, whether
			// the time is on the project or on one of its tasks.
			website.TagIds = nil
			require.ErrorIs(t, repo.DeleteProject(ctx, testScope, website.Id), model.ErrConflict)
			require.NoError(t, repo.DeleteTimespan(ctx, testScope, direct))
			require.ErrorIs(t, repo.DeleteProject(ctx, testScope, website.Id), model.ErrConflict)
			require.NoError(t, repo.DeleteTimespan(ctx, testScope, onTask))

			// Once the project is deleted, its tasks no longer belong to it.
			require.NoError(t, repo.DeleteProject(ctx, testScope, website.Id))
			ids, err = repo.ListTaskProjectIds(ctx, testScope, []uuid.UUID{parent.Id, child.Id})
			require.NoError(t, err)
			require.Empty(t, ids)
		})

		t.Run(repoName+"DeleteIgnoresTimeThroughLinkedTags", func(t *testing.T) {
			repo := newRepo(t)

			label, err := repo.CreateTag(ctx, testScope, model.Tag{Name: "client", Color: "#88c0d0"})
			require.NoError(t, err)
			project, err := repo.CreateProject(ctx, testScope, model.Project{
				Name: "Website", Color: "#bf616a", TagIds: []uuid.UUID{label.Id},
			})
			require.NoError(t, err)
			start := time.Date(2026, 9, 1, 9, 0, 0, 0, time.UTC)
			span, err := repo.CreateTimespan(ctx, testScope, model.Timespan{
				StartTime: start, EndTime: start.Add(time.Hour), TagIds: []uuid.UUID{label.Id},
			})
			require.NoError(t, err)

			// Time reaching the project only through a linked tag stays with
			// that tag, so the project can go.
			require.NoError(t, repo.DeleteProject(ctx, testScope, project.Id))
			_, err = repo.GetProject(ctx, testScope, project.Id)
			require.ErrorIs(t, err, model.ErrNotFound)
			_, err = repo.GetTag(ctx, testScope, label.Id)
			require.NoError(t, err)
			got, err := repo.GetTimespan(ctx, testScope, span.Id)
			require.NoError(t, err)
			require.Equal(t, []uuid.UUID{label.Id}, got.TagIds)
		})

		t.Run(repoName+"DeleteRacesUpdateSafely", func(t *testing.T) {
			repo := newRepo(t)
			label, err := repo.CreateTag(ctx, testScope, model.Tag{Name: "client", Color: "#88c0d0"})
			require.NoError(t, err)
			for range 20 {
				project, err := repo.CreateProject(ctx, testScope, model.Project{
					Name: "Website", Color: "#bf616a", TagIds: []uuid.UUID{label.Id},
				})
				require.NoError(t, err)

				// Saving and deleting a project at once: the delete always
				// wins, and the save either lands first or finds it gone.
				var wg sync.WaitGroup
				var errDelete, errUpdate error
				wg.Go(func() { errDelete = repo.DeleteProject(ctx, testScope, project.Id) })
				wg.Go(func() {
					project.Name = "Web site"
					_, errUpdate = repo.UpdateProject(ctx, testScope, project)
				})
				wg.Wait()

				require.NoError(t, errDelete)
				if errUpdate != nil {
					require.ErrorIs(t, errUpdate, model.ErrNotFound)
				}
			}
		})

		t.Run(repoName+"UpdateRacesTaskUpdateSafely", func(t *testing.T) {
			repo := newRepo(t)
			label, err := repo.CreateTag(ctx, testScope, model.Tag{Name: "client", Color: "#88c0d0"})
			require.NoError(t, err)
			for range 50 {
				project, err := repo.CreateProject(ctx, testScope, model.Project{Name: "Website", Color: "#bf616a"})
				require.NoError(t, err)
				task, err := repo.CreateTask(ctx, testScope, model.Task{Name: "Launch", TagIds: []uuid.UUID{project.TagId}})
				require.NoError(t, err)
				project.TagIds = []uuid.UUID{task.TagId}
				project, err = repo.UpdateProject(ctx, testScope, project)
				require.NoError(t, err)

				// The project links the task's tag and the task carries the
				// project's tag: renaming both at once, each keeping the
				// other's tag, rewrites each owned tag while checking the
				// other, and both saves land.
				var wg sync.WaitGroup
				var errProject, errTask error
				wg.Go(func() {
					project.Name = "Web site"
					_, errProject = repo.UpdateProject(ctx, testScope, project)
				})
				wg.Go(func() {
					_, errTask = repo.UpdateTask(ctx, testScope, task.Id, model.TaskPatch{
						Name:   new("Go live"),
						TagIds: &[]uuid.UUID{project.TagId, label.Id},
					})
				})
				wg.Wait()

				require.NoError(t, errProject)
				require.NoError(t, errTask)
			}
		})

		t.Run(repoName+"DeleteRacesTimespanSafely", func(t *testing.T) {
			repo := newRepo(t)
			start := time.Date(2026, 9, 1, 9, 0, 0, 0, time.UTC)
			for range 20 {
				project, err := repo.CreateProject(ctx, testScope, model.Project{Name: "Website", Color: "#bf616a"})
				require.NoError(t, err)

				// Deleting a project and logging time on it at once: one
				// must lose, or time ends up logged on a deleted tag.
				var wg sync.WaitGroup
				var errDelete, errLog error
				wg.Go(func() { errDelete = repo.DeleteProject(ctx, testScope, project.Id) })
				wg.Go(func() {
					_, errLog = repo.CreateTimespan(ctx, testScope, model.Timespan{
						StartTime: start, EndTime: start.Add(time.Hour), TagIds: []uuid.UUID{project.TagId},
					})
				})
				wg.Wait()

				if errDelete == nil {
					require.ErrorIs(t, errLog, model.ErrInvalidReference)
				} else {
					require.ErrorIs(t, errDelete, model.ErrConflict)
					require.NoError(t, errLog)
				}
			}
		})

		t.Run(repoName+"DeleteRacesTimespanOnTaskSafely", func(t *testing.T) {
			repo := newRepo(t)
			start := time.Date(2026, 9, 1, 9, 0, 0, 0, time.UTC)
			for range 20 {
				project, err := repo.CreateProject(ctx, testScope, model.Project{Name: "Website", Color: "#bf616a"})
				require.NoError(t, err)
				task, err := repo.CreateTask(ctx, testScope, model.Task{Name: "Launch", TagIds: []uuid.UUID{project.TagId}})
				require.NoError(t, err)

				// Logging time on a task of a project being deleted: the log
				// always lands, and counts toward the project only if the
				// project survives.
				var wg sync.WaitGroup
				var errDelete, errLog error
				wg.Go(func() { errDelete = repo.DeleteProject(ctx, testScope, project.Id) })
				wg.Go(func() {
					_, errLog = repo.CreateTimespan(ctx, testScope, model.Timespan{
						StartTime: start, EndTime: start.Add(time.Hour), TagIds: []uuid.UUID{task.TagId},
					})
				})
				wg.Wait()

				require.NoError(t, errLog)
				total, err := repo.GetTotalDurationByTags(ctx, testScope, []uuid.UUID{project.TagId})
				require.NoError(t, err)
				if errDelete == nil {
					require.Zero(t, total)
					ids, err := repo.ListTaskProjectIds(ctx, testScope, []uuid.UUID{task.Id})
					require.NoError(t, err)
					require.Empty(t, ids)
				} else {
					require.ErrorIs(t, errDelete, model.ErrConflict)
					require.Equal(t, time.Hour, total)
				}
			}
		})
	}

	run(t, "memory", func(t *testing.T) repository.Repository {
		return memory.NewMemoryStore()
	})

	run(t, "postgres", func(t *testing.T) repository.Repository {
		t.Parallel()
		pool := testutils.StartPostgresContainerWithMigrationsApplied(ctx, t)
		repo := postgres.NewPostgresStoreFromPool(pool)
		seedScopeUser(t, ctx, repo, testScope)
		return repo
	})
}
