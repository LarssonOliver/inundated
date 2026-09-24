package demo_test

import (
	"context"
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

func TestSeed_CreatesTagsProjectsAndTimespans(t *testing.T) {
	repo := memory.NewMemoryStore()
	now := time.Date(2026, 9, 24, 15, 0, 0, 0, time.UTC)

	require.NoError(t, demo.Seed(context.Background(), repo, now))

	tags, err := repo.ListTags(context.Background(), model.UnownedScope(), largePage)
	require.NoError(t, err)
	assert.NotEmpty(t, tags.Data)

	projects, err := repo.ListProjects(context.Background(), model.UnownedScope(), largePage)
	require.NoError(t, err)
	assert.NotEmpty(t, projects.Data)

	timespans, err := repo.ListTimespans(context.Background(), model.UnownedScope(), largePage)
	require.NoError(t, err)
	assert.NotEmpty(t, timespans.Data)
}

func TestSeed_DataIsUnowned(t *testing.T) {
	repo := memory.NewMemoryStore()
	now := time.Date(2026, 9, 24, 15, 0, 0, 0, time.UTC)
	require.NoError(t, demo.Seed(context.Background(), repo, now))

	tags, err := repo.ListTags(context.Background(), model.UnownedScope(), largePage)
	require.NoError(t, err)
	for _, tag := range tags.Data {
		assert.Nil(t, tag.UserId)
	}

	timespans, err := repo.ListTimespans(context.Background(), model.UnownedScope(), largePage)
	require.NoError(t, err)
	for _, ts := range timespans.Data {
		assert.Nil(t, ts.UserId)
	}
}

func TestSeed_ProjectsReferenceExistingTags(t *testing.T) {
	repo := memory.NewMemoryStore()
	now := time.Date(2026, 9, 24, 15, 0, 0, 0, time.UTC)
	require.NoError(t, demo.Seed(context.Background(), repo, now))

	tags, err := repo.ListTags(context.Background(), model.UnownedScope(), largePage)
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

	tags, err := repo.ListTags(context.Background(), model.UnownedScope(), largePage)
	require.NoError(t, err)
	knownTags := make(map[uuid.UUID]bool, len(tags.Data))
	for _, tag := range tags.Data {
		knownTags[tag.Id] = true
	}

	timespans, err := repo.ListTimespans(context.Background(), model.UnownedScope(), largePage)
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

	timespans, err := repo.ListTimespans(context.Background(), model.UnownedScope(), largePage)
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
	timespansA, err := repoA.ListTimespans(context.Background(), model.UnownedScope(), largePage)
	require.NoError(t, err)

	repoB := memory.NewMemoryStore()
	require.NoError(t, demo.Seed(context.Background(), repoB, now))
	timespansB, err := repoB.ListTimespans(context.Background(), model.UnownedScope(), largePage)
	require.NoError(t, err)

	assert.Equal(t, timespansA.TotalCount, timespansB.TotalCount)
}
