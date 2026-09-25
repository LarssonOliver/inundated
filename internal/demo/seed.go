// Package demo seeds a repository with realistic, time-relative sample data
// for demo instances. It is only meant to run against a fresh in-memory
// repository in userless mode - see cmd/server/main.go for the gating.
package demo

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
)

// seedRandSource is fixed so a demo instance's data looks the same across
// restarts, rather than reshuffling every time the server starts.
const seedRandSource = 20260101

const weeksOfHistory = 6

// Nord aurora/frost colors, matching frontend/src/helpers/nord.ts.
const (
	colorBlue   = "#5e81ac" // nord10
	colorGreen  = "#a3be8c" // nord14
	colorPurple = "#b48ead" // nord15
	colorYellow = "#ebcb8b" // nord13
	colorRed    = "#bf616a" // nord11
	colorTeal   = "#8fbcbb" // nord7
	colorFrost1 = "#88c0d0" // nord8
	colorFrost2 = "#91a1c1" // nord9
	colorOrange = "#d08770" // nord12
)

type tagSpec struct {
	name  string
	color string
	tasks []string
}

var tagSpecs = []tagSpec{
	{name: "Meetings", color: colorBlue, tasks: []string{
		"Client sync", "Sprint planning", "Design review", "Standup", "1:1 with manager",
	}},
	{name: "Development", color: colorGreen, tasks: []string{
		"Implement API endpoint", "Refactor auth module", "Code review", "Write unit tests", "Set up CI pipeline",
	}},
	{name: "Design", color: colorPurple, tasks: []string{
		"Wireframes", "Component library", "Usability review", "Icon set", "Landing page mockup",
	}},
	{name: "Planning", color: colorYellow, tasks: []string{
		"Roadmap update", "Backlog grooming", "Quarterly planning", "Stakeholder alignment",
	}},
	{name: "Bug Fixes", color: colorRed, tasks: []string{
		"Fix login bug", "Investigate memory leak", "Patch broken pagination", "Hotfix release",
	}},
	{name: "Research", color: colorTeal, tasks: []string{
		"Evaluate new library", "Spike: caching strategy", "Competitor analysis", "Read RFC drafts",
	}},
}

type projectSpec struct {
	name       string
	color      string
	tagNames   []string
	timeBudget *time.Duration
}

var websiteRedesignBudget = 80 * time.Hour

var projectSpecs = []projectSpec{
	{name: "Website Redesign", color: colorFrost1, tagNames: []string{"Design", "Development"}, timeBudget: &websiteRedesignBudget},
	{name: "Mobile App", color: colorFrost2, tagNames: []string{"Development", "Bug Fixes"}},
	{name: "Client Onboarding", color: colorOrange, tagNames: []string{"Meetings", "Planning"}},
	{name: "Internal Tools", color: colorTeal, tagNames: []string{"Development", "Research"}},
}

// daySlot is a non-overlapping block of a working day that a timespan can be
// generated into.
type daySlot struct {
	startHour, startMinute int
	endHour, endMinute     int
}

var daySlots = []daySlot{
	{startHour: 9, endHour: 12},
	{startHour: 13, endHour: 15},
	{startHour: 15, startMinute: 30, endHour: 17},
}

// Seed populates repo's unowned scope with a handful of tags, projects, and a
// realistic spread of timesheet entries across the weekdays of the
// weeksOfHistory before now. It is intended to run once against a freshly
// created, empty repository.
func Seed(ctx context.Context, repo repository.Repository, now time.Time) error {
	scope := model.UnownedScope()
	rng := rand.New(rand.NewSource(seedRandSource))

	tagsByName, err := seedTags(ctx, repo, scope)
	if err != nil {
		return err
	}

	projects, err := seedProjects(ctx, repo, scope, tagsByName)
	if err != nil {
		return err
	}

	return seedTimespans(ctx, repo, scope, projects, now, rng)
}

func seedTags(ctx context.Context, repo repository.Repository, scope model.OwnerScope) (map[string]model.Tag, error) {
	tagsByName := make(map[string]model.Tag, len(tagSpecs))
	for _, spec := range tagSpecs {
		tag, err := repo.CreateTag(ctx, scope, model.Tag{Name: spec.name, Color: spec.color})
		if err != nil {
			return nil, fmt.Errorf("demo: seeding tag %q: %w", spec.name, err)
		}
		tagsByName[spec.name] = tag
	}
	return tagsByName, nil
}

type seededProject struct {
	model.Project
	tasks []string
}

func seedProjects(ctx context.Context, repo repository.Repository, scope model.OwnerScope, tagsByName map[string]model.Tag) ([]seededProject, error) {
	projects := make([]seededProject, 0, len(projectSpecs))
	for _, spec := range projectSpecs {
		tagIds := make([]uuid.UUID, 0, len(spec.tagNames))
		var tasks []string
		for _, tagName := range spec.tagNames {
			tag := tagsByName[tagName]
			tagIds = append(tagIds, tag.Id)
			for _, tagSpecEntry := range tagSpecs {
				if tagSpecEntry.name == tagName {
					tasks = append(tasks, tagSpecEntry.tasks...)
				}
			}
		}

		project, err := repo.CreateProject(ctx, scope, model.Project{
			Name:       spec.name,
			Color:      spec.color,
			TagIds:     tagIds,
			TimeBudget: spec.timeBudget,
		})
		if err != nil {
			return nil, fmt.Errorf("demo: seeding project %q: %w", spec.name, err)
		}

		projects = append(projects, seededProject{Project: project, tasks: tasks})
	}
	return projects, nil
}

func seedTimespans(
	ctx context.Context,
	repo repository.Repository,
	scope model.OwnerScope,
	projects []seededProject,
	now time.Time,
	rng *rand.Rand,
) error {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	firstDay := today.AddDate(0, 0, -weeksOfHistory*7)

	for day := firstDay; !day.After(today); day = day.AddDate(0, 0, 1) {
		if day.Weekday() == time.Saturday || day.Weekday() == time.Sunday {
			continue
		}
		// Skip some days at random so the calendar doesn't look
		// unrealistically dense.
		if rng.Float64() < 0.15 {
			continue
		}

		slots := rng.Perm(len(daySlots))
		entryCount := 1 + rng.Intn(3)
		if entryCount > len(daySlots) {
			entryCount = len(daySlots)
		}

		for _, slotIdx := range slots[:entryCount] {
			slot := daySlots[slotIdx]
			start := time.Date(day.Year(), day.Month(), day.Day(), slot.startHour, slot.startMinute, 0, 0, day.Location())
			end := time.Date(day.Year(), day.Month(), day.Day(), slot.endHour, slot.endMinute, 0, 0, day.Location())

			if start.After(now) {
				continue
			}
			if end.After(now) {
				end = now
			}
			if !start.Before(end) {
				continue
			}

			project := projects[rng.Intn(len(projects))]
			name := project.tasks[rng.Intn(len(project.tasks))]

			_, err := repo.CreateTimespan(ctx, scope, model.Timespan{
				Name:      name,
				StartTime: start,
				EndTime:   end,
				TagIds:    project.TagIds,
			})
			if err != nil {
				return fmt.Errorf("demo: seeding timespan %q on %s: %w", name, day.Format("2006-01-02"), err)
			}
		}
	}

	return nil
}
