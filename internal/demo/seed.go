// Package demo seeds a repository with realistic, time-relative sample data
// for demo instances. It is only meant to run against a fresh in-memory
// repository in userless mode - see cmd/server/main.go for the gating.
package demo

import (
	"context"
	"fmt"
	"math/rand"
	"slices"
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

// taskSpec is a demo task. Time logged under a timespan named like the task
// is logged on it. The due date is days from now, when set.
type taskSpec struct {
	name      string
	tagNames  []string
	dueInDays *int
	estimate  *time.Duration
	// project assigns the task by adding its task tag to that project.
	project  string
	closed   model.CloseReason
	subtasks []taskSpec
}

func days(n int) *int { return &n }

func hours(h int) *time.Duration {
	d := time.Duration(h) * time.Hour
	return &d
}

var taskSpecs = []taskSpec{
	{
		name: "Launch new website", tagNames: []string{"Design"}, dueInDays: days(12), estimate: hours(60),
		project: "Website Redesign",
		subtasks: []taskSpec{
			{name: "Wireframes", closed: model.CloseReasonDone},
			{name: "Landing page mockup", dueInDays: days(2), estimate: hours(8)},
			{name: "Implement API endpoint", tagNames: []string{"Development"}, dueInDays: days(6), estimate: hours(16)},
			{name: "Usability review", dueInDays: days(9)},
		},
	},
	{name: "Fix login bug", tagNames: []string{"Bug Fixes"}, dueInDays: days(1), estimate: hours(4)},
	{name: "Patch broken pagination", tagNames: []string{"Bug Fixes"}},
	{name: "Quarterly planning", tagNames: []string{"Planning"}, dueInDays: days(15)},
	{name: "Evaluate new library", tagNames: []string{"Research"}, closed: model.CloseReasonDone},
	{name: "Icon set", closed: model.CloseReasonIgnored},
	{name: "Set up CI pipeline", project: "Internal Tools", estimate: hours(6)},
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

	taskTagsByName, toClose, err := seedTasks(ctx, repo, scope, tagsByName, projects, now)
	if err != nil {
		return err
	}

	if err := seedTimespans(ctx, repo, scope, projects, taskTagsByName, now, rng); err != nil {
		return err
	}

	// Closing archives a task's tag, and new time can't be logged on an
	// archived tag, so tasks are closed last.
	for _, task := range toClose {
		closed := true
		patch := model.TaskPatch{Closed: &closed, CloseReason: task.CloseReason}
		if _, err := repo.UpdateTask(ctx, scope, task.Id, patch); err != nil {
			return fmt.Errorf("demo: closing task %q: %w", task.Name, err)
		}
	}
	return nil
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
	// labelIds are the project's regular tags, without assigned tasks'.
	labelIds []uuid.UUID
	tasks    []string
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

		projects = append(projects, seededProject{Project: project, labelIds: tagIds, tasks: tasks})
	}
	return projects, nil
}

// seedTasks creates taskSpecs, assigning them as specified. It returns each
// task's task tag id by task name, and the tasks to close, already marked
// with their close reason.
func seedTasks(
	ctx context.Context,
	repo repository.Repository,
	scope model.OwnerScope,
	tagsByName map[string]model.Tag,
	projects []seededProject,
	now time.Time,
) (map[string]uuid.UUID, []model.Task, error) {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	taskTags := make(map[string]uuid.UUID)
	var toClose []model.Task

	var create func(spec taskSpec, parentId *uuid.UUID) error
	create = func(spec taskSpec, parentId *uuid.UUID) error {
		task := model.Task{Name: spec.name, ParentId: parentId, Estimate: spec.estimate}
		for _, tagName := range spec.tagNames {
			task.TagIds = append(task.TagIds, tagsByName[tagName].Id)
		}
		if spec.dueInDays != nil {
			due := today.AddDate(0, 0, *spec.dueInDays)
			task.DueDate = &due
		}
		created, err := repo.CreateTask(ctx, scope, task)
		if err != nil {
			return fmt.Errorf("demo: seeding task %q: %w", spec.name, err)
		}
		taskTags[spec.name] = created.TagId

		if spec.project != "" {
			if err := assignTask(ctx, repo, scope, projects, spec.project, created.TagId); err != nil {
				return err
			}
		}
		for _, subtask := range spec.subtasks {
			if err := create(subtask, &created.Id); err != nil {
				return err
			}
		}
		if spec.closed != "" {
			reason := spec.closed
			created.CloseReason = &reason
			toClose = append(toClose, created)
		}
		return nil
	}

	for _, spec := range taskSpecs {
		if err := create(spec, nil); err != nil {
			return nil, nil, err
		}
	}
	return taskTags, toClose, nil
}

// assignTask adds taskTagId to the named project's tags.
func assignTask(ctx context.Context, repo repository.Repository, scope model.OwnerScope, projects []seededProject, projectName string, taskTagId uuid.UUID) error {
	for i := range projects {
		if projects[i].Name != projectName {
			continue
		}
		projects[i].TagIds = append(projects[i].TagIds, taskTagId)
		updated, err := repo.UpdateProject(ctx, scope, projects[i].Project)
		if err != nil {
			return fmt.Errorf("demo: assigning task to project %q: %w", projectName, err)
		}
		projects[i].Project = updated
		return nil
	}
	return fmt.Errorf("demo: no project %q", projectName)
}

func seedTimespans(
	ctx context.Context,
	repo repository.Repository,
	scope model.OwnerScope,
	projects []seededProject,
	taskTagsByName map[string]uuid.UUID,
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

			// Log the time on the project's regular tags, and on the task
			// of the same name if there is one.
			tagIds := slices.Clone(project.labelIds)
			if taskTagId, ok := taskTagsByName[name]; ok {
				tagIds = append(tagIds, taskTagId)
			}

			_, err := repo.CreateTimespan(ctx, scope, model.Timespan{
				Name:      name,
				StartTime: start,
				EndTime:   end,
				TagIds:    tagIds,
			})
			if err != nil {
				return fmt.Errorf("demo: seeding timespan %q on %s: %w", name, day.Format("2006-01-02"), err)
			}
		}
	}

	return nil
}
