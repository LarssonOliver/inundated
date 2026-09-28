package contract_test

import (
	"context"
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

// TestTaskAttributionContract runs the worked example from the tasks design
// on both stores. The project Website has the tag acme. The task Build
// landing page has the tag design and is assigned to Website; its subtask
// Hero section has no regular tags. Time is logged as:
//
//	Hero markup   2h    #hero
//	Copy review   1h    #landing
//	Client call   0.5h  acme
func TestTaskAttributionContract(t *testing.T) {
	ctx := context.Background()

	run := func(t *testing.T, repoName string, newRepo func(t *testing.T) repository.Repository) {
		t.Run(repoName+"WorkedExample", func(t *testing.T) {
			repo := newRepo(t)

			tag := func(name string) model.Tag {
				created, err := repo.CreateTag(ctx, testScope, model.Tag{Name: name, Color: "#88c0d0"})
				require.NoError(t, err)
				return created
			}
			acme, design := tag("acme"), tag("design")

			landing, err := repo.CreateTask(ctx, testScope, model.Task{Name: "Build landing page", TagIds: []uuid.UUID{design.Id}})
			require.NoError(t, err)
			hero, err := repo.CreateTask(ctx, testScope, model.Task{Name: "Hero section", ParentId: &landing.Id})
			require.NoError(t, err)
			website, err := repo.CreateProject(ctx, testScope, model.Project{
				Name: "Website", Color: "#5e81ac", TagIds: []uuid.UUID{acme.Id, landing.TagId},
			})
			require.NoError(t, err)

			start := time.Date(2026, 9, 1, 9, 0, 0, 0, time.UTC)
			logTime := func(d time.Duration, tagId uuid.UUID) {
				_, err := repo.CreateTimespan(ctx, testScope, model.Timespan{
					StartTime: start, EndTime: start.Add(d), TagIds: []uuid.UUID{tagId},
				})
				require.NoError(t, err)
				start = start.Add(d)
			}
			logTime(2*time.Hour, hero.TagId)
			logTime(time.Hour, landing.TagId)
			logTime(30*time.Minute, acme.Id)

			total := func(tagIds ...uuid.UUID) time.Duration {
				d, err := repo.GetTotalDurationByTags(ctx, testScope, tagIds)
				require.NoError(t, err)
				return d
			}
			projectsOf := func(taskIds ...uuid.UUID) map[uuid.UUID][]uuid.UUID {
				got, err := repo.ListTaskProjectIds(ctx, testScope, taskIds)
				require.NoError(t, err)
				return got
			}
			tasksIn := func(projectId uuid.UUID) []string {
				page, err := repo.ListTasks(ctx, testScope, model.TaskListParams{
					PaginationParams: model.PaginationParams{Limit: 100},
					ProjectId:        &projectId,
				})
				require.NoError(t, err)
				out := make([]string, len(page.Data))
				for i, task := range page.Data {
					out[i] = task.Name
				}
				return out
			}

			// Each timespan counts toward everything its effective tags reach.
			require.Equal(t, 2*time.Hour, total(hero.TagId))
			require.Equal(t, 3*time.Hour, total(landing.TagId), "subtask time rolls up")
			require.Equal(t, 3*time.Hour, total(design.Id), "tag time includes time on tasks with the tag")
			require.Equal(t, 30*time.Minute, total(acme.Id))
			require.Equal(t, 3*time.Hour+30*time.Minute, total(website.TagIds...), "each timespan counts once per project")

			// The batch form gives each tag the same total, in one call.
			unused := tag("unused")
			perTag, err := repo.GetTotalDurationPerTag(ctx, testScope, []uuid.UUID{hero.TagId, landing.TagId, design.Id, acme.Id, unused.Id})
			require.NoError(t, err)
			require.Equal(t, map[uuid.UUID]time.Duration{
				hero.TagId:    2 * time.Hour,
				landing.TagId: 3 * time.Hour,
				design.Id:     3 * time.Hour,
				acme.Id:       30 * time.Minute,
			}, perTag)

			series, err := repo.AggregateTimeSpentByTagsAndBuckets(ctx, testScope, []uuid.UUID{design.Id}, []model.BucketRange{
				{Start: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)},
				{Start: time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC), End: time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)},
			})
			require.NoError(t, err)
			require.Equal(t, time.Hour.Seconds(), series[0].Value)
			require.Equal(t, (2 * time.Hour).Seconds(), series[1].Value)

			// Both tasks reach Website through the landing page's task tag.
			require.Equal(t, map[uuid.UUID][]uuid.UUID{
				landing.Id: {website.Id},
				hero.Id:    {website.Id},
			}, projectsOf(landing.Id, hero.Id))
			require.Equal(t, []string{"Build landing page", "Hero section"}, tasksIn(website.Id))

			// A task sharing one of the project's regular tags belongs to it
			// without being assigned, and so does its time.
			invoice, err := repo.CreateTask(ctx, testScope, model.Task{Name: "Send invoice", TagIds: []uuid.UUID{acme.Id}})
			require.NoError(t, err)
			logTime(15*time.Minute, invoice.TagId)
			require.Equal(t, map[uuid.UUID][]uuid.UUID{invoice.Id: {website.Id}}, projectsOf(invoice.Id))
			require.Equal(t, 45*time.Minute, total(acme.Id))
			require.Equal(t, 3*time.Hour+45*time.Minute, total(website.TagIds...))

			// Re-tagging a task moves its past time with it.
			_, err = repo.UpdateTask(ctx, testScope, landing.Id, model.TaskPatch{TagIds: &[]uuid.UUID{}})
			require.NoError(t, err)
			require.Zero(t, total(design.Id))
			require.Equal(t, 3*time.Hour, total(landing.TagId))

			// Moving the subtask out takes its time out of the parent and the
			// project.
			_, err = repo.MoveTask(ctx, testScope, hero.Id, nil, nil)
			require.NoError(t, err)
			require.Equal(t, time.Hour, total(landing.TagId))
			require.Equal(t, time.Hour+45*time.Minute, total(website.TagIds...))
			require.Empty(t, projectsOf(hero.Id))
			require.Equal(t, []string{"Build landing page", "Send invoice"}, tasksIn(website.Id))

			// Moving it back, under a task with a regular tag, brings the
			// time and the tag with it.
			_, err = repo.UpdateTask(ctx, testScope, landing.Id, model.TaskPatch{TagIds: &[]uuid.UUID{design.Id}})
			require.NoError(t, err)
			_, err = repo.MoveTask(ctx, testScope, hero.Id, &landing.Id, nil)
			require.NoError(t, err)
			require.Equal(t, 3*time.Hour, total(design.Id))
			require.Equal(t, 3*time.Hour+45*time.Minute, total(website.TagIds...))

			// A deleted task no longer carries anything.
			scratch, err := repo.CreateTask(ctx, testScope, model.Task{Name: "Scratch", ParentId: &landing.Id})
			require.NoError(t, err)
			require.NoError(t, repo.DeleteTask(ctx, testScope, scratch.Id))
			require.NotContains(t, projectsOf(scratch.Id), scratch.Id)
			require.NotContains(t, tasksIn(website.Id), "Scratch")

			// Another user sees none of it.
			other, err := repo.ListTaskProjectIds(ctx, model.UnownedScope(), []uuid.UUID{landing.Id})
			require.NoError(t, err)
			require.Empty(t, other)
		})

		t.Run(repoName+"NoOpChangesKeepEffectiveTags", func(t *testing.T) {
			repo := newRepo(t)
			a, err := repo.CreateTag(ctx, testScope, model.Tag{Name: "a", Color: "#88c0d0"})
			require.NoError(t, err)
			b, err := repo.CreateTag(ctx, testScope, model.Tag{Name: "b", Color: "#88c0d0"})
			require.NoError(t, err)
			parent, err := repo.CreateTask(ctx, testScope, model.Task{Name: "parent", TagIds: []uuid.UUID{a.Id, b.Id}})
			require.NoError(t, err)
			first, err := repo.CreateTask(ctx, testScope, model.Task{Name: "first", ParentId: &parent.Id})
			require.NoError(t, err)
			second, err := repo.CreateTask(ctx, testScope, model.Task{Name: "second", ParentId: &parent.Id})
			require.NoError(t, err)
			project, err := repo.CreateProject(ctx, testScope, model.Project{Name: "p", Color: "#5e81ac", TagIds: []uuid.UUID{b.Id}})
			require.NoError(t, err)
			allInProject := func() {
				t.Helper()
				ids, err := repo.ListTaskProjectIds(ctx, testScope, []uuid.UUID{parent.Id, first.Id, second.Id})
				require.NoError(t, err)
				for _, id := range []uuid.UUID{parent.Id, first.Id, second.Id} {
					require.Equal(t, []uuid.UUID{project.Id}, ids[id])
				}
			}
			allInProject()

			// Sending the same tags in another order, and reordering the
			// subtasks under the same parent, change nothing they inherit.
			updated, err := repo.UpdateTask(ctx, testScope, parent.Id, model.TaskPatch{TagIds: &[]uuid.UUID{b.Id, a.Id, b.Id}})
			require.NoError(t, err)
			require.ElementsMatch(t, []uuid.UUID{a.Id, b.Id}, updated.TagIds)
			_, err = repo.MoveTask(ctx, testScope, second.Id, &parent.Id, nil)
			require.NoError(t, err)
			allInProject()

			// A real change still takes effect.
			_, err = repo.UpdateTask(ctx, testScope, parent.Id, model.TaskPatch{TagIds: &[]uuid.UUID{a.Id}})
			require.NoError(t, err)
			ids, err := repo.ListTaskProjectIds(ctx, testScope, []uuid.UUID{parent.Id, first.Id, second.Id})
			require.NoError(t, err)
			require.Empty(t, ids)
		})

		t.Run(repoName+"DeletedTagNoLongerLinksTaskAndProject", func(t *testing.T) {
			repo := newRepo(t)
			shared, err := repo.CreateTag(ctx, testScope, model.Tag{Name: "shared", Color: "#88c0d0"})
			require.NoError(t, err)
			task, err := repo.CreateTask(ctx, testScope, model.Task{Name: "t", TagIds: []uuid.UUID{shared.Id}})
			require.NoError(t, err)
			project, err := repo.CreateProject(ctx, testScope, model.Project{Name: "p", Color: "#5e81ac", TagIds: []uuid.UUID{shared.Id}})
			require.NoError(t, err)

			inProject := func() int {
				page, err := repo.ListTasks(ctx, testScope, model.TaskListParams{PaginationParams: model.DefaultPaginationParams(), ProjectId: &project.Id})
				require.NoError(t, err)
				return len(page.Data)
			}
			ids, err := repo.ListTaskProjectIds(ctx, testScope, []uuid.UUID{task.Id})
			require.NoError(t, err)
			require.Equal(t, []uuid.UUID{project.Id}, ids[task.Id])
			require.Equal(t, 1, inProject())

			// The shared tag was the only link; deleting it breaks it.
			require.NoError(t, repo.DeleteTag(ctx, testScope, shared.Id))
			ids, err = repo.ListTaskProjectIds(ctx, testScope, []uuid.UUID{task.Id})
			require.NoError(t, err)
			require.Empty(t, ids[task.Id])
			require.Zero(t, inProject())
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
