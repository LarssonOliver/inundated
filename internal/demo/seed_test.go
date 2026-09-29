package demo_test

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/larssonoliver/inundated/internal/demo"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository/memory"
)

var largePage = model.PaginationParams{Limit: 10000}
var largeTimespanPage = model.TimespanListParams{PaginationParams: largePage}

func TestSeed_CreatesTagsProjectsAndTimespans(t *testing.T) {
	repo := memory.NewMemoryStore()
	now := time.Date(2026, 9, 24, 15, 0, 0, 0, time.UTC)

	require.NoError(t, demo.Seed(context.Background(), repo, now))

	tags, err := repo.ListTags(context.Background(), model.UnownedScope(), model.TagListParams{PaginationParams: largePage})
	require.NoError(t, err)
	assert.NotEmpty(t, tags.Data)

	projects, err := repo.ListProjects(context.Background(), model.UnownedScope(), largePage)
	require.NoError(t, err)
	assert.NotEmpty(t, projects.Data)

	timespans, err := repo.ListTimespans(context.Background(), model.UnownedScope(), largeTimespanPage)
	require.NoError(t, err)
	assert.NotEmpty(t, timespans.Data)
}

func TestSeed_DataIsUnowned(t *testing.T) {
	repo := memory.NewMemoryStore()
	now := time.Date(2026, 9, 24, 15, 0, 0, 0, time.UTC)
	require.NoError(t, demo.Seed(context.Background(), repo, now))

	tags, err := repo.ListTags(context.Background(), model.UnownedScope(), model.TagListParams{PaginationParams: largePage})
	require.NoError(t, err)
	for _, tag := range tags.Data {
		assert.Nil(t, tag.UserId)
	}

	timespans, err := repo.ListTimespans(context.Background(), model.UnownedScope(), largeTimespanPage)
	require.NoError(t, err)
	for _, ts := range timespans.Data {
		assert.Nil(t, ts.UserId)
	}
}

func TestSeed_ProjectsReferenceExistingTags(t *testing.T) {
	repo := memory.NewMemoryStore()
	now := time.Date(2026, 9, 24, 15, 0, 0, 0, time.UTC)
	require.NoError(t, demo.Seed(context.Background(), repo, now))

	// Projects carry the task tags of the tasks assigned to them.
	tags, err := repo.ListTags(context.Background(), model.UnownedScope(), model.TagListParams{PaginationParams: largePage, Kind: model.TagKindAll})
	require.NoError(t, err)
	knownTags := make(map[uuid.UUID]bool, len(tags.Data))
	for _, tag := range tags.Data {
		knownTags[tag.Id] = true
	}

	projects, err := repo.ListProjects(context.Background(), model.UnownedScope(), largePage)
	require.NoError(t, err)
	require.NotEmpty(t, projects.Data)
	for _, p := range projects.Data {
		for _, tagID := range p.TagIds {
			assert.True(t, knownTags[tagID], "project %q references unknown tag %s", p.Name, tagID)
		}
	}
}

func TestSeed_TimespansReferenceExistingTags(t *testing.T) {
	repo := memory.NewMemoryStore()
	now := time.Date(2026, 9, 24, 15, 0, 0, 0, time.UTC)
	require.NoError(t, demo.Seed(context.Background(), repo, now))

	// Time is logged on task tags too, including closed (archived) tasks'.
	allTags := model.PaginationParams{Limit: largePage.Limit, IncludeArchived: true}
	tags, err := repo.ListTags(context.Background(), model.UnownedScope(), model.TagListParams{PaginationParams: allTags, Kind: model.TagKindAll})
	require.NoError(t, err)
	knownTags := make(map[uuid.UUID]bool, len(tags.Data))
	for _, tag := range tags.Data {
		knownTags[tag.Id] = true
	}

	timespans, err := repo.ListTimespans(context.Background(), model.UnownedScope(), largeTimespanPage)
	require.NoError(t, err)
	require.NotEmpty(t, timespans.Data)
	for _, ts := range timespans.Data {
		require.NotEmpty(t, ts.TagIds, "timespan %q has no tags", ts.Name)
		for _, tagID := range ts.TagIds {
			assert.True(t, knownTags[tagID], "timespan %q references unknown tag %s", ts.Name, tagID)
		}
	}
}

func TestSeed_TimespansAreWithinSixWeekWindowRelativeToNow(t *testing.T) {
	repo := memory.NewMemoryStore()
	now := time.Date(2026, 9, 24, 15, 0, 0, 0, time.UTC)
	require.NoError(t, demo.Seed(context.Background(), repo, now))

	windowStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -6*7)

	timespans, err := repo.ListTimespans(context.Background(), model.UnownedScope(), largeTimespanPage)
	require.NoError(t, err)
	require.NotEmpty(t, timespans.Data)
	for _, ts := range timespans.Data {
		assert.True(t, ts.StartTime.Before(ts.EndTime), "timespan %q: start not before end", ts.Name)
		assert.False(t, ts.StartTime.Before(windowStart), "timespan %q starts before the 6-week window", ts.Name)
		assert.False(t, ts.EndTime.After(now), "timespan %q ends after now", ts.Name)
	}
}

func TestSeed_IsDeterministicForAGivenNow(t *testing.T) {
	now := time.Date(2026, 9, 24, 15, 0, 0, 0, time.UTC)

	repoA := memory.NewMemoryStore()
	require.NoError(t, demo.Seed(context.Background(), repoA, now))
	timespansA, err := repoA.ListTimespans(context.Background(), model.UnownedScope(), largeTimespanPage)
	require.NoError(t, err)

	repoB := memory.NewMemoryStore()
	require.NoError(t, demo.Seed(context.Background(), repoB, now))
	timespansB, err := repoB.ListTimespans(context.Background(), model.UnownedScope(), largeTimespanPage)
	require.NoError(t, err)

	assert.Equal(t, timespansA.TotalCount, timespansB.TotalCount)
}

func TestSeed_CreatesTasksWithTimeAndProjects(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewMemoryStore()
	now := time.Date(2026, 9, 24, 15, 0, 0, 0, time.UTC)
	require.NoError(t, demo.Seed(ctx, repo, now))

	all, err := repo.ListTasks(ctx, model.UnownedScope(), model.TaskListParams{PaginationParams: largePage, IncludeClosed: true})
	require.NoError(t, err)
	open, err := repo.ListTasks(ctx, model.UnownedScope(), model.TaskListParams{PaginationParams: largePage})
	require.NoError(t, err)
	assert.Less(t, len(open.Data), len(all.Data), "some tasks are closed")

	byName := make(map[string]model.Task, len(all.Data))
	for _, task := range all.Data {
		byName[task.Name] = task
	}
	launch := byName["Launch new website"]
	require.NotNil(t, launch.DueDate)
	assert.Equal(t, launch.Id, *byName["Landing page mockup"].ParentId)

	// Subtask time rolls up into the parent.
	launchTime, err := repo.GetTotalDurationByTags(ctx, model.UnownedScope(), []uuid.UUID{launch.TagId})
	require.NoError(t, err)
	assert.Positive(t, launchTime)

	projectIds, err := repo.ListTaskProjectIds(ctx, model.UnownedScope(), []uuid.UUID{launch.Id})
	require.NoError(t, err)
	assert.NotEmpty(t, projectIds[launch.Id], "the launch task is assigned to a project")
}

func TestSeed_TaskTimeStaysInItsProject(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewMemoryStore()
	now := time.Date(2026, 9, 24, 15, 0, 0, 0, time.UTC)
	require.NoError(t, demo.Seed(ctx, repo, now))
	scope := model.UnownedScope()

	tasks, err := repo.ListTasks(ctx, scope, model.TaskListParams{PaginationParams: largePage, IncludeClosed: true})
	require.NoError(t, err)
	taskByTag := make(map[uuid.UUID]uuid.UUID, len(tasks.Data))
	for _, task := range tasks.Data {
		taskByTag[task.TagId] = task.Id
	}

	// Each timespan carries one project's regular tags; key projects by them.
	projects, err := repo.ListProjects(ctx, scope, largePage)
	require.NoError(t, err)
	labelKey := func(tagIds []uuid.UUID) string {
		var labels []string
		for _, id := range tagIds {
			if _, isTaskTag := taskByTag[id]; !isTaskTag {
				labels = append(labels, id.String())
			}
		}
		slices.Sort(labels)
		return strings.Join(labels, ",")
	}
	projectByLabels := make(map[string]uuid.UUID, len(projects.Data))
	for _, project := range projects.Data {
		projectByLabels[labelKey(project.TagIds)] = project.Id
	}

	timespans, err := repo.ListTimespans(ctx, scope, largeTimespanPage)
	require.NoError(t, err)
	onTasks := 0
	for _, ts := range timespans.Data {
		projectId, ok := projectByLabels[labelKey(ts.TagIds)]
		require.True(t, ok, "timespan %q matches no project", ts.Name)
		for _, tagId := range ts.TagIds {
			taskId, isTaskTag := taskByTag[tagId]
			if !isTaskTag {
				continue
			}
			onTasks++
			projectIds, err := repo.ListTaskProjectIds(ctx, scope, []uuid.UUID{taskId})
			require.NoError(t, err)
			assert.Equal(t, []uuid.UUID{projectId}, projectIds[taskId],
				"timespan %q is logged on a task outside its project", ts.Name)
		}
	}
	assert.Positive(t, onTasks, "some demo time is logged on tasks")
}
