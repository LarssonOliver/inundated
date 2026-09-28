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
	"github.com/larssonoliver/inundated/internal/utils"
	"github.com/larssonoliver/inundated/test/testutils"
	"github.com/stretchr/testify/require"
)

func TestTaskRepositoryContract(t *testing.T) {
	ctx := context.Background()

	allTasks := model.TaskListParams{
		PaginationParams: model.PaginationParams{Limit: 100},
		IncludeClosed:    true,
	}
	allTags := model.TagListParams{
		PaginationParams: model.PaginationParams{Limit: 100, IncludeArchived: true},
		Kind:             model.TagKindAll,
	}
	day := func(y int, m time.Month, d int) *time.Time {
		t := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
		return &t
	}
	reason := func(r model.CloseReason) *model.CloseReason { return &r }
	names := func(tasks []model.Task) []string {
		out := make([]string, len(tasks))
		for i, task := range tasks {
			out[i] = task.Name
		}
		return out
	}

	run := func(t *testing.T, repoName string, newRepo func(t *testing.T) repository.Repository) {
		t.Run(repoName+"CreateAndGet", func(t *testing.T) {
			repo := newRepo(t)
			label, err := repo.CreateTag(ctx, testScope, model.Tag{Name: "work", Color: "#ff0000"})
			require.NoError(t, err)

			estimate := 90 * time.Minute
			created, err := repo.CreateTask(ctx, testScope, model.Task{
				Name:     "Write report",
				TagIds:   []uuid.UUID{label.Id},
				DueDate:  day(2026, 10, 1),
				Estimate: &estimate,
			})
			require.NoError(t, err)
			require.NotEqual(t, uuid.Nil, created.Id)
			require.NotEqual(t, uuid.Nil, created.TagId)
			require.NotEmpty(t, created.Rank)
			require.False(t, created.Closed())
			require.Equal(t, *testScope.UserID(), *created.UserId)

			got, err := repo.GetTask(ctx, testScope, created.Id)
			require.NoError(t, err)
			require.Equal(t, "Write report", got.Name)
			require.Equal(t, []uuid.UUID{label.Id}, got.TagIds)
			require.True(t, got.DueDate.Equal(*day(2026, 10, 1)))
			require.Equal(t, estimate, *got.Estimate)
			require.Nil(t, got.ParentId)

			// The task tag carries the task's name and points back at it.
			tag, err := repo.GetTag(ctx, testScope, created.TagId)
			require.NoError(t, err)
			require.Equal(t, "Write report", tag.Name)
			require.NotNil(t, tag.TaskId)
			require.Equal(t, created.Id, *tag.TaskId)

			_, err = repo.GetTask(ctx, testScope, uuid.New())
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"TaskTagsInTagList", func(t *testing.T) {
			repo := newRepo(t)
			_, err := repo.CreateTag(ctx, testScope, model.Tag{Name: "Design", Color: "#00ff00"})
			require.NoError(t, err)
			_, err = repo.CreateTag(ctx, testScope, model.Tag{Name: "100% done_ish", Color: "#00ff00"})
			require.NoError(t, err)
			task, err := repo.CreateTask(ctx, testScope, model.Task{Name: "Design review"})
			require.NoError(t, err)

			labels, err := repo.ListTags(ctx, testScope, model.TagListParams{PaginationParams: model.DefaultPaginationParams()})
			require.NoError(t, err)
			require.Equal(t, 2, labels.TotalCount)
			for _, tag := range labels.Data {
				require.Nil(t, tag.TaskId)
			}

			tasks, err := repo.ListTags(ctx, testScope, model.TagListParams{PaginationParams: model.DefaultPaginationParams(), Kind: model.TagKindTask})
			require.NoError(t, err)
			require.Len(t, tasks.Data, 1)
			require.Equal(t, task.TagId, tasks.Data[0].Id)

			// Regular tags come first, and search ignores case.
			both, err := repo.ListTags(ctx, testScope, model.TagListParams{PaginationParams: model.DefaultPaginationParams(), Kind: model.TagKindAll, Query: "DESIGN"})
			require.NoError(t, err)
			require.Equal(t, 2, both.TotalCount)
			require.Equal(t, "Design", both.Data[0].Name)
			require.Equal(t, "Design review", both.Data[1].Name)

			// LIKE wildcards in the search text match literally.
			literal, err := repo.ListTags(ctx, testScope, model.TagListParams{PaginationParams: model.DefaultPaginationParams(), Query: "0% d"})
			require.NoError(t, err)
			require.Len(t, literal.Data, 1)
			none, err := repo.ListTags(ctx, testScope, model.TagListParams{PaginationParams: model.DefaultPaginationParams(), Query: "e_r"})
			require.NoError(t, err)
			require.Empty(t, none.Data)
		})

		t.Run(repoName+"TaskTagColorAndArchivedFollowTask", func(t *testing.T) {
			repo := newRepo(t)
			b, err := repo.CreateTag(ctx, testScope, model.Tag{Name: "b", Color: "#222222"})
			require.NoError(t, err)
			a, err := repo.CreateTag(ctx, testScope, model.Tag{Name: "a", Color: "#111111"})
			require.NoError(t, err)

			bare, err := repo.CreateTask(ctx, testScope, model.Task{Name: "bare"})
			require.NoError(t, err)
			tagged, err := repo.CreateTask(ctx, testScope, model.Task{Name: "tagged", TagIds: []uuid.UUID{b.Id, a.Id}})
			require.NoError(t, err)

			bareTag, err := repo.GetTag(ctx, testScope, bare.TagId)
			require.NoError(t, err)
			require.Equal(t, model.DefaultTaskTagColor, bareTag.Color)
			taggedTag, err := repo.GetTag(ctx, testScope, tagged.TagId)
			require.NoError(t, err)
			require.Equal(t, "#111111", taggedTag.Color)

			_, err = repo.UpdateTask(ctx, testScope, tagged.Id, model.TaskPatch{CloseReason: reason(model.CloseReasonDone)})
			require.NoError(t, err)

			taggedTag, err = repo.GetTag(ctx, testScope, tagged.TagId)
			require.NoError(t, err)
			require.True(t, taggedTag.Archived)
			visible, err := repo.ListTags(ctx, testScope, model.TagListParams{PaginationParams: model.DefaultPaginationParams(), Kind: model.TagKindTask})
			require.NoError(t, err)
			require.Len(t, visible.Data, 1)
			require.Equal(t, bare.TagId, visible.Data[0].Id)

			_, err = repo.UpdateTask(ctx, testScope, tagged.Id, model.TaskPatch{Closed: new(false)})
			require.NoError(t, err)
			taggedTag, err = repo.GetTag(ctx, testScope, tagged.TagId)
			require.NoError(t, err)
			require.False(t, taggedTag.Archived)
		})

		t.Run(repoName+"TaskTagsAttachToTimespansAndProjectsButNotTasks", func(t *testing.T) {
			repo := newRepo(t)
			first, err := repo.CreateTask(ctx, testScope, model.Task{Name: "first"})
			require.NoError(t, err)

			_, err = repo.CreateTask(ctx, testScope, model.Task{Name: "second", TagIds: []uuid.UUID{first.TagId}})
			require.ErrorIs(t, err, model.ErrInvalidReference)
			second, err := repo.CreateTask(ctx, testScope, model.Task{Name: "second"})
			require.NoError(t, err)
			_, err = repo.UpdateTask(ctx, testScope, second.Id, model.TaskPatch{TagIds: &[]uuid.UUID{first.TagId}})
			require.ErrorIs(t, err, model.ErrInvalidReference)

			start := time.Date(2026, 9, 1, 9, 0, 0, 0, time.UTC)
			span, err := repo.CreateTimespan(ctx, testScope, model.Timespan{
				StartTime: start, EndTime: start.Add(time.Hour), TagIds: []uuid.UUID{first.TagId},
			})
			require.NoError(t, err)
			project, err := repo.CreateProject(ctx, testScope, model.Project{
				Name: "p", Color: "#123456", TagIds: []uuid.UUID{first.TagId},
			})
			require.NoError(t, err)
			require.Equal(t, []uuid.UUID{first.TagId}, project.TagIds)

			// A closed task's tag behaves like an archived tag: existing
			// links survive edits, new ones are refused.
			_, err = repo.UpdateTask(ctx, testScope, first.Id, model.TaskPatch{CloseReason: reason(model.CloseReasonIgnored)})
			require.NoError(t, err)

			span.Name = "renamed"
			_, err = repo.UpdateTimespan(ctx, testScope, span)
			require.NoError(t, err)
			_, err = repo.CreateTimespan(ctx, testScope, model.Timespan{
				StartTime: start, EndTime: start.Add(time.Hour), TagIds: []uuid.UUID{first.TagId},
			})
			require.ErrorIs(t, err, model.ErrInvalidReference)
		})

		t.Run(repoName+"UpdateReplacesFieldsAndRenamesTag", func(t *testing.T) {
			repo := newRepo(t)
			label, err := repo.CreateTag(ctx, testScope, model.Tag{Name: "l", Color: "#123456"})
			require.NoError(t, err)
			estimate := time.Hour
			task, err := repo.CreateTask(ctx, testScope, model.Task{
				Name: "old", TagIds: []uuid.UUID{label.Id}, DueDate: day(2026, 1, 1), Estimate: &estimate,
			})
			require.NoError(t, err)

			updated, err := repo.UpdateTask(ctx, testScope, task.Id, model.TaskPatch{
				Name: new("new"), TagIds: &[]uuid.UUID{}, ClearDueDate: true, ClearEstimate: true,
			})
			require.NoError(t, err)
			require.Equal(t, "new", updated.Name)
			require.Empty(t, updated.TagIds)
			require.Nil(t, updated.DueDate)
			require.Nil(t, updated.Estimate)
			require.Equal(t, task.Rank, updated.Rank)

			tag, err := repo.GetTag(ctx, testScope, task.TagId)
			require.NoError(t, err)
			require.Equal(t, "new", tag.Name)

			_, err = repo.UpdateTask(ctx, testScope, uuid.New(), model.TaskPatch{Name: new("new")})
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"CloseAndReopenCascade", func(t *testing.T) {
			repo := newRepo(t)
			root, err := repo.CreateTask(ctx, testScope, model.Task{Name: "root"})
			require.NoError(t, err)
			child, err := repo.CreateTask(ctx, testScope, model.Task{Name: "child", ParentId: &root.Id})
			require.NoError(t, err)
			grandchild, err := repo.CreateTask(ctx, testScope, model.Task{Name: "grandchild", ParentId: &child.Id})
			require.NoError(t, err)
			ignored, err := repo.CreateTask(ctx, testScope, model.Task{Name: "ignored", ParentId: &root.Id})
			require.NoError(t, err)

			_, err = repo.UpdateTask(ctx, testScope, ignored.Id, model.TaskPatch{CloseReason: reason(model.CloseReasonIgnored)})
			require.NoError(t, err)

			closedRoot, err := repo.UpdateTask(ctx, testScope, root.Id, model.TaskPatch{Closed: new(true)})
			require.NoError(t, err)
			require.NotNil(t, closedRoot.ClosedAt)

			for _, id := range []uuid.UUID{child.Id, grandchild.Id} {
				got, err := repo.GetTask(ctx, testScope, id)
				require.NoError(t, err)
				require.Equal(t, model.CloseReasonDone, *got.CloseReason)
			}
			// Already-closed subtasks keep their own reason.
			got, err := repo.GetTask(ctx, testScope, ignored.Id)
			require.NoError(t, err)
			require.Equal(t, model.CloseReasonIgnored, *got.CloseReason)

			open, err := repo.ListTasks(ctx, testScope, model.TaskListParams{PaginationParams: model.DefaultPaginationParams()})
			require.NoError(t, err)
			require.Empty(t, open.Data)

			// Reopening the grandchild reopens its ancestors, but not its
			// siblings' subtrees.
			_, err = repo.UpdateTask(ctx, testScope, grandchild.Id, model.TaskPatch{Closed: new(false)})
			require.NoError(t, err)

			open, err = repo.ListTasks(ctx, testScope, model.TaskListParams{PaginationParams: model.DefaultPaginationParams()})
			require.NoError(t, err)
			require.ElementsMatch(t, []string{"root", "child", "grandchild"}, names(open.Data))

			// A closed parent takes no new open subtasks.
			_, err = repo.UpdateTask(ctx, testScope, root.Id, model.TaskPatch{Closed: new(true)})
			require.NoError(t, err)
			_, err = repo.CreateTask(ctx, testScope, model.Task{Name: "late", ParentId: &root.Id})
			require.ErrorIs(t, err, model.ErrInvalidArgument)
			_, err = repo.CreateTask(ctx, testScope, model.Task{Name: "orphan", ParentId: new(uuid.New())})
			require.ErrorIs(t, err, model.ErrInvalidReference)
		})

		t.Run(repoName+"ListOrderAndFilters", func(t *testing.T) {
			repo := newRepo(t)
			label, err := repo.CreateTag(ctx, testScope, model.Tag{Name: "l", Color: "#123456"})
			require.NoError(t, err)

			first, err := repo.CreateTask(ctx, testScope, model.Task{Name: "first", DueDate: day(2026, 3, 1)})
			require.NoError(t, err)
			_, err = repo.CreateTask(ctx, testScope, model.Task{Name: "second", TagIds: []uuid.UUID{label.Id}, DueDate: day(2026, 3, 15)})
			require.NoError(t, err)
			_, err = repo.CreateTask(ctx, testScope, model.Task{Name: "third"})
			require.NoError(t, err)
			_, err = repo.CreateTask(ctx, testScope, model.Task{Name: "sub-a", ParentId: &first.Id})
			require.NoError(t, err)
			_, err = repo.CreateTask(ctx, testScope, model.Task{Name: "sub-b", ParentId: &first.Id, DueDate: day(2026, 4, 1)})
			require.NoError(t, err)

			page, err := repo.ListTasks(ctx, testScope, allTasks)
			require.NoError(t, err)
			require.Equal(t, 5, page.TotalCount)
			require.Equal(t, []string{"first", "second", "third", "sub-a", "sub-b"}, names(page.Data))

			sub, err := repo.ListTasks(ctx, testScope, model.TaskListParams{PaginationParams: model.DefaultPaginationParams(), ParentId: &first.Id})
			require.NoError(t, err)
			require.Equal(t, []string{"sub-a", "sub-b"}, names(sub.Data))

			byTag, err := repo.ListTasks(ctx, testScope, model.TaskListParams{PaginationParams: model.DefaultPaginationParams(), TagId: &label.Id})
			require.NoError(t, err)
			require.Equal(t, []string{"second"}, names(byTag.Data))

			due, err := repo.ListTasks(ctx, testScope, model.TaskListParams{
				PaginationParams: model.DefaultPaginationParams(), DueFrom: day(2026, 3, 1), DueTo: day(2026, 3, 31),
			})
			require.NoError(t, err)
			require.Equal(t, []string{"first", "second"}, names(due.Data))

			paged, err := repo.ListTasks(ctx, testScope, model.TaskListParams{PaginationParams: model.PaginationParams{Limit: 2, Offset: 2}})
			require.NoError(t, err)
			require.Equal(t, 5, paged.TotalCount)
			require.Equal(t, []string{"third", "sub-a"}, names(paged.Data))
		})

		t.Run(repoName+"Move", func(t *testing.T) {
			repo := newRepo(t)
			a, err := repo.CreateTask(ctx, testScope, model.Task{Name: "a"})
			require.NoError(t, err)
			b, err := repo.CreateTask(ctx, testScope, model.Task{Name: "b"})
			require.NoError(t, err)
			c, err := repo.CreateTask(ctx, testScope, model.Task{Name: "c"})
			require.NoError(t, err)
			topLevel := func() []string {
				page, err := repo.ListTasks(ctx, testScope, allTasks)
				require.NoError(t, err)
				var out []string
				for _, task := range page.Data {
					if task.ParentId == nil {
						out = append(out, task.Name)
					}
				}
				return out
			}

			_, err = repo.MoveTask(ctx, testScope, c.Id, nil, nil)
			require.NoError(t, err)
			require.Equal(t, []string{"c", "a", "b"}, topLevel())

			_, err = repo.MoveTask(ctx, testScope, c.Id, nil, &a.Id)
			require.NoError(t, err)
			require.Equal(t, []string{"a", "c", "b"}, topLevel())

			moved, err := repo.MoveTask(ctx, testScope, b.Id, &a.Id, nil)
			require.NoError(t, err)
			require.Equal(t, a.Id, *moved.ParentId)
			require.Equal(t, []string{"a", "c"}, topLevel())

			// A task can't move under itself or its own subtasks.
			_, err = repo.MoveTask(ctx, testScope, a.Id, &a.Id, nil)
			require.ErrorIs(t, err, model.ErrInvalidArgument)
			_, err = repo.MoveTask(ctx, testScope, a.Id, &b.Id, nil)
			require.ErrorIs(t, err, model.ErrInvalidArgument)

			// afterId must be a sibling under the new parent.
			_, err = repo.MoveTask(ctx, testScope, c.Id, &a.Id, &a.Id)
			require.ErrorIs(t, err, model.ErrInvalidArgument)
			_, err = repo.MoveTask(ctx, testScope, c.Id, nil, new(uuid.New()))
			require.ErrorIs(t, err, model.ErrInvalidArgument)

			// An open task can't move under a closed one.
			_, err = repo.UpdateTask(ctx, testScope, b.Id, model.TaskPatch{Closed: new(true)})
			require.NoError(t, err)
			_, err = repo.MoveTask(ctx, testScope, c.Id, &b.Id, nil)
			require.ErrorIs(t, err, model.ErrInvalidArgument)

			_, err = repo.MoveTask(ctx, testScope, uuid.New(), nil, nil)
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"MoveRespacesCrowdedRanks", func(t *testing.T) {
			repo := newRepo(t)
			a, err := repo.CreateTask(ctx, testScope, model.Task{Name: "a"})
			require.NoError(t, err)
			b, err := repo.CreateTask(ctx, testScope, model.Task{Name: "b"})
			require.NoError(t, err)

			// Swapping the two front-most tasks over and over squeezes
			// ranks toward the start until the list gets respaced.
			for i := range 250 {
				id := a.Id
				if i%2 == 1 {
					id = b.Id
				}
				_, err := repo.MoveTask(ctx, testScope, id, nil, nil)
				require.NoError(t, err)
			}

			page, err := repo.ListTasks(ctx, testScope, allTasks)
			require.NoError(t, err)
			require.Equal(t, []string{"b", "a"}, names(page.Data))
			for _, task := range page.Data {
				require.LessOrEqual(t, len(task.Rank), utils.MaxRankLength)
			}
		})

		t.Run(repoName+"DeleteRefusesLoggedTime", func(t *testing.T) {
			repo := newRepo(t)
			parent, err := repo.CreateTask(ctx, testScope, model.Task{Name: "parent"})
			require.NoError(t, err)
			child, err := repo.CreateTask(ctx, testScope, model.Task{Name: "child", ParentId: &parent.Id})
			require.NoError(t, err)
			project, err := repo.CreateProject(ctx, testScope, model.Project{
				Name: "p", Color: "#123456", TagIds: []uuid.UUID{parent.TagId},
			})
			require.NoError(t, err)

			start := time.Date(2026, 9, 1, 9, 0, 0, 0, time.UTC)
			span, err := repo.CreateTimespan(ctx, testScope, model.Timespan{
				StartTime: start, EndTime: start.Add(time.Hour), TagIds: []uuid.UUID{child.TagId},
			})
			require.NoError(t, err)

			// Time on a subtask blocks deleting its parent.
			require.ErrorIs(t, repo.DeleteTask(ctx, testScope, parent.Id), model.ErrConflict)
			_, err = repo.GetTask(ctx, testScope, child.Id)
			require.NoError(t, err)

			require.NoError(t, repo.DeleteTimespan(ctx, testScope, span.Id))
			require.NoError(t, repo.DeleteTask(ctx, testScope, parent.Id))

			for _, id := range []uuid.UUID{parent.Id, child.Id} {
				_, err := repo.GetTask(ctx, testScope, id)
				require.ErrorIs(t, err, model.ErrNotFound)
			}
			for _, id := range []uuid.UUID{parent.TagId, child.TagId} {
				_, err := repo.GetTag(ctx, testScope, id)
				require.ErrorIs(t, err, model.ErrNotFound)
			}
			got, err := repo.GetProject(ctx, testScope, project.Id)
			require.NoError(t, err)
			require.Empty(t, got.TagIds)

			require.ErrorIs(t, repo.DeleteTask(ctx, testScope, parent.Id), model.ErrNotFound)
		})

		t.Run(repoName+"DeletedTagLeavesTask", func(t *testing.T) {
			repo := newRepo(t)
			keep, err := repo.CreateTag(ctx, testScope, model.Tag{Name: "keep", Color: "#123456"})
			require.NoError(t, err)
			gone, err := repo.CreateTag(ctx, testScope, model.Tag{Name: "gone", Color: "#123456"})
			require.NoError(t, err)
			task, err := repo.CreateTask(ctx, testScope, model.Task{Name: "t", TagIds: []uuid.UUID{keep.Id, gone.Id}})
			require.NoError(t, err)

			require.NoError(t, repo.DeleteTag(ctx, testScope, gone.Id))

			got, err := repo.GetTask(ctx, testScope, task.Id)
			require.NoError(t, err)
			require.Equal(t, []uuid.UUID{keep.Id}, got.TagIds)
			page, err := repo.ListTasks(ctx, testScope, allTasks)
			require.NoError(t, err)
			require.Equal(t, []uuid.UUID{keep.Id}, page.Data[0].TagIds)
			byGone, err := repo.ListTasks(ctx, testScope, model.TaskListParams{PaginationParams: model.DefaultPaginationParams(), TagId: &gone.Id})
			require.NoError(t, err)
			require.Empty(t, byGone.Data)
			require.Zero(t, byGone.TotalCount)

			// Updates that leave the tags alone still work.
			updated, err := repo.UpdateTask(ctx, testScope, task.Id, model.TaskPatch{Name: new("renamed"), Closed: new(true)})
			require.NoError(t, err)
			require.Equal(t, []uuid.UUID{keep.Id}, updated.TagIds)
		})

		t.Run(repoName+"TaskTagsRefuseDirectEdits", func(t *testing.T) {
			repo := newRepo(t)
			task, err := repo.CreateTask(ctx, testScope, model.Task{Name: "t"})
			require.NoError(t, err)

			_, err = repo.UpdateTag(ctx, testScope, model.Tag{Id: task.TagId, Name: "x", Color: "#123456"})
			require.ErrorIs(t, err, model.ErrInvalidArgument)
			require.ErrorIs(t, repo.DeleteTag(ctx, testScope, task.TagId), model.ErrInvalidArgument)

			tag, err := repo.GetTag(ctx, testScope, task.TagId)
			require.NoError(t, err)
			require.Equal(t, "t", tag.Name)

			// Another user's task tag is simply not found.
			other := model.UserScope(uuid.New())
			seedScopeUser(t, ctx, repo, other)
			_, err = repo.UpdateTag(ctx, other, model.Tag{Id: task.TagId, Name: "x", Color: "#123456"})
			require.ErrorIs(t, err, model.ErrNotFound)
			require.ErrorIs(t, repo.DeleteTag(ctx, other, task.TagId), model.ErrNotFound)
		})

		t.Run(repoName+"TagNamesOrderIgnoringCase", func(t *testing.T) {
			repo := newRepo(t)
			banana, err := repo.CreateTag(ctx, testScope, model.Tag{Name: "Banana", Color: "#0000ff"})
			require.NoError(t, err)
			apple, err := repo.CreateTag(ctx, testScope, model.Tag{Name: "apple", Color: "#ff0000"})
			require.NoError(t, err)
			_, err = repo.CreateTag(ctx, testScope, model.Tag{Name: "cherry", Color: "#00ff00"})
			require.NoError(t, err)

			page, err := repo.ListTags(ctx, testScope, model.TagListParams{PaginationParams: model.DefaultPaginationParams()})
			require.NoError(t, err)
			var got []string
			for _, tag := range page.Data {
				got = append(got, tag.Name)
			}
			require.Equal(t, []string{"apple", "Banana", "cherry"}, got)

			task, err := repo.CreateTask(ctx, testScope, model.Task{Name: "t", TagIds: []uuid.UUID{banana.Id, apple.Id}})
			require.NoError(t, err)
			taskTag, err := repo.GetTag(ctx, testScope, task.TagId)
			require.NoError(t, err)
			require.Equal(t, apple.Color, taskTag.Color)
		})

		t.Run(repoName+"ConcurrentCreatesGetDistinctRanks", func(t *testing.T) {
			repo := newRepo(t)
			const n = 20
			var wg sync.WaitGroup
			errs := make([]error, n)
			for i := range n {
				wg.Go(func() {
					_, errs[i] = repo.CreateTask(ctx, testScope, model.Task{Name: "t"})
				})
			}
			wg.Wait()
			for _, err := range errs {
				require.NoError(t, err)
			}

			page, err := repo.ListTasks(ctx, testScope, allTasks)
			require.NoError(t, err)
			ranks := map[string]bool{}
			for _, task := range page.Data {
				ranks[task.Rank] = true
			}
			require.Len(t, ranks, n)
		})

		t.Run(repoName+"ConcurrentMovesCantFormACycle", func(t *testing.T) {
			repo := newRepo(t)
			for range 10 {
				// B > C and A > D. Moving A under C and B under D are each
				// fine alone, but together they'd make A > C... > D > A.
				a, err := repo.CreateTask(ctx, testScope, model.Task{Name: "a"})
				require.NoError(t, err)
				b, err := repo.CreateTask(ctx, testScope, model.Task{Name: "b"})
				require.NoError(t, err)
				c, err := repo.CreateTask(ctx, testScope, model.Task{Name: "c", ParentId: &b.Id})
				require.NoError(t, err)
				d, err := repo.CreateTask(ctx, testScope, model.Task{Name: "d", ParentId: &a.Id})
				require.NoError(t, err)

				var wg sync.WaitGroup
				var errA, errB error
				wg.Go(func() { _, errA = repo.MoveTask(ctx, testScope, a.Id, &c.Id, nil) })
				wg.Go(func() { _, errB = repo.MoveTask(ctx, testScope, b.Id, &d.Id, nil) })
				wg.Wait()

				require.True(t, (errA == nil) != (errB == nil), "exactly one move must win: %v, %v", errA, errB)
				for _, err := range []error{errA, errB} {
					if err != nil {
						require.ErrorIs(t, err, model.ErrInvalidArgument)
					}
				}
			}
		})

		t.Run(repoName+"PatchesKeepOtherFields", func(t *testing.T) {
			repo := newRepo(t)
			parent, err := repo.CreateTask(ctx, testScope, model.Task{Name: "parent"})
			require.NoError(t, err)
			task, err := repo.CreateTask(ctx, testScope, model.Task{Name: "task", ParentId: &parent.Id, DueDate: day(2026, 5, 1)})
			require.NoError(t, err)

			// A rename after a close leaves the close alone, instead of
			// writing back an open state read before it.
			_, err = repo.UpdateTask(ctx, testScope, parent.Id, model.TaskPatch{CloseReason: reason(model.CloseReasonIgnored)})
			require.NoError(t, err)
			renamed, err := repo.UpdateTask(ctx, testScope, task.Id, model.TaskPatch{Name: new("renamed")})
			require.NoError(t, err)
			require.Equal(t, "renamed", renamed.Name)
			require.Equal(t, model.CloseReasonIgnored, *renamed.CloseReason)
			require.True(t, renamed.DueDate.Equal(*day(2026, 5, 1)))
			got, err := repo.GetTask(ctx, testScope, parent.Id)
			require.NoError(t, err)
			require.True(t, got.Closed())

			// closed: true on a closed task keeps its reason.
			again, err := repo.UpdateTask(ctx, testScope, task.Id, model.TaskPatch{Closed: new(true)})
			require.NoError(t, err)
			require.Equal(t, model.CloseReasonIgnored, *again.CloseReason)

			// Concurrent patches to different fields both land.
			fresh, err := repo.CreateTask(ctx, testScope, model.Task{Name: "fresh"})
			require.NoError(t, err)
			var wg sync.WaitGroup
			var errClose, errRename error
			wg.Go(func() { _, errClose = repo.UpdateTask(ctx, testScope, fresh.Id, model.TaskPatch{Closed: new(true)}) })
			wg.Go(func() {
				_, errRename = repo.UpdateTask(ctx, testScope, fresh.Id, model.TaskPatch{Name: new("fresh renamed")})
			})
			wg.Wait()
			require.NoError(t, errClose)
			require.NoError(t, errRename)
			got, err = repo.GetTask(ctx, testScope, fresh.Id)
			require.NoError(t, err)
			require.True(t, got.Closed())
			require.Equal(t, "fresh renamed", got.Name)

			_, err = repo.UpdateTask(ctx, testScope, task.Id, model.TaskPatch{Name: new("")})
			require.ErrorIs(t, err, model.ErrInvalidArgument)
		})

		t.Run(repoName+"DeleteRacesTimespanSafely", func(t *testing.T) {
			repo := newRepo(t)
			start := time.Date(2026, 9, 1, 9, 0, 0, 0, time.UTC)
			for range 20 {
				task, err := repo.CreateTask(ctx, testScope, model.Task{Name: "t"})
				require.NoError(t, err)

				// Deleting a task and logging time on it at once: one must
				// lose, or time ends up logged on a deleted task.
				var wg sync.WaitGroup
				var errDelete, errLog error
				wg.Go(func() { errDelete = repo.DeleteTask(ctx, testScope, task.Id) })
				wg.Go(func() {
					_, errLog = repo.CreateTimespan(ctx, testScope, model.Timespan{
						StartTime: start, EndTime: start.Add(time.Hour), TagIds: []uuid.UUID{task.TagId},
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

		t.Run(repoName+"ScopeIsolation", func(t *testing.T) {
			repo := newRepo(t)
			scopeA := model.UserScope(uuid.New())
			scopeB := model.UserScope(uuid.New())
			seedScopeUser(t, ctx, repo, scopeA)
			seedScopeUser(t, ctx, repo, scopeB)

			taskA, err := repo.CreateTask(ctx, scopeA, model.Task{Name: "a"})
			require.NoError(t, err)
			taskB, err := repo.CreateTask(ctx, scopeB, model.Task{Name: "b"})
			require.NoError(t, err)

			pageA, err := repo.ListTasks(ctx, scopeA, allTasks)
			require.NoError(t, err)
			require.Equal(t, []string{"a"}, names(pageA.Data))

			_, err = repo.GetTask(ctx, scopeB, taskA.Id)
			require.ErrorIs(t, err, model.ErrNotFound)
			_, err = repo.GetTag(ctx, scopeB, taskA.TagId)
			require.ErrorIs(t, err, model.ErrNotFound)
			tagsB, err := repo.ListTags(ctx, scopeB, allTags)
			require.NoError(t, err)
			require.Len(t, tagsB.Data, 1)
			require.Equal(t, taskB.TagId, tagsB.Data[0].Id)

			_, err = repo.UpdateTask(ctx, scopeB, taskA.Id, model.TaskPatch{Name: new("hijack")})
			require.ErrorIs(t, err, model.ErrNotFound)
			_, err = repo.MoveTask(ctx, scopeB, taskA.Id, nil, nil)
			require.ErrorIs(t, err, model.ErrNotFound)
			require.ErrorIs(t, repo.DeleteTask(ctx, scopeB, taskA.Id), model.ErrNotFound)

			// Another user's task can't be a parent, and its tag can't be
			// attached.
			_, err = repo.CreateTask(ctx, scopeB, model.Task{Name: "c", ParentId: &taskA.Id})
			require.ErrorIs(t, err, model.ErrInvalidReference)
			_, err = repo.MoveTask(ctx, scopeB, taskB.Id, &taskA.Id, nil)
			require.ErrorIs(t, err, model.ErrInvalidArgument)
			_, err = repo.CreateProject(ctx, scopeB, model.Project{Name: "p", Color: "#123456", TagIds: []uuid.UUID{taskA.TagId}})
			require.ErrorIs(t, err, model.ErrInvalidReference)

			got, err := repo.GetTask(ctx, scopeA, taskA.Id)
			require.NoError(t, err)
			require.Equal(t, "a", got.Name)
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
