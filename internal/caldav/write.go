package caldav

import (
	"context"
	"errors"

	"github.com/emersion/go-ical"
	gocaldav "github.com/emersion/go-webdav/caldav"
	"github.com/larssonoliver/inundated/internal/model"
)

// putTimespan creates or updates the timespan served under name. Tags and
// anything else the event can't express are left as they are.
func (b *backend) putTimespan(ctx context.Context, name string, calendar *ical.Calendar, options *gocaldav.PutCalendarObjectOptions) error {
	location, err := b.loadLocation(ctx)
	if err != nil {
		return err
	}
	event, err := ParseEvent(calendar, location)
	if err != nil {
		return err
	}
	tags, err := b.loadTags(ctx)
	if err != nil {
		return err
	}

	stored, err := b.lookupTimespan(ctx, name)
	if errors.Is(err, model.ErrNotFound) {
		if err := checkPreconditions(options, nil); err != nil {
			return err
		}
		if event.Summary == untitled {
			event.Summary = ""
		}
		_, err = b.svc.CreateTimespan(ctx, model.Timespan{Name: event.Summary, StartTime: event.Start, EndTime: event.End, CaldavName: &name})
		return err
	}
	if err != nil {
		return err
	}
	object, err := timespanObject(stored, tags, location)
	if err != nil {
		return err
	}
	if err := checkPreconditions(options, object); err != nil {
		return err
	}

	// A client sends back the summary it was given, which for an unnamed
	// timespan is made up from its tags.
	if event.Summary != timespanSummary(stored, tags) {
		stored.Name = event.Summary
	}
	stored.StartTime, stored.EndTime = event.Start, event.End
	_, err = b.svc.UpdateTimespan(ctx, stored)
	return err
}

// putTask creates a top-level task, or updates the task served under name.
func (b *backend) putTask(ctx context.Context, name string, calendar *ical.Calendar, options *gocaldav.PutCalendarObjectOptions) error {
	location, err := b.loadLocation(ctx)
	if err != nil {
		return err
	}
	todo, err := ParseTodo(calendar, location)
	if err != nil {
		return err
	}
	tags, err := b.loadTags(ctx)
	if err != nil {
		return err
	}

	stored, err := b.lookupTask(ctx, name)
	if errors.Is(err, model.ErrNotFound) {
		if err := checkPreconditions(options, nil); err != nil {
			return err
		}
		if todo.Summary == "" {
			todo.Summary = untitled
		}
		_, err = b.svc.CreateTask(ctx, model.Task{Name: todo.Summary, DueDate: todo.Due, CaldavName: &name})
		return err
	}
	if err != nil {
		return err
	}
	object, err := taskObject(stored, tags)
	if err != nil {
		return err
	}
	if err := checkPreconditions(options, object); err != nil {
		return err
	}

	patch := taskPatch(stored, todo)
	if patch == (model.TaskPatch{}) {
		return nil
	}
	_, err = b.svc.UpdateTask(ctx, stored.Id, patch)
	return err
}

// taskPatch changes only what differs, since closing and reopening ripple
// through a task's subtasks and ancestors.
func taskPatch(task model.Task, todo Todo) model.TaskPatch {
	var patch model.TaskPatch
	if todo.Summary != "" && todo.Summary != task.Name {
		patch.Name = &todo.Summary
	}

	switch {
	case todo.Due == nil && task.DueDate != nil:
		patch.ClearDueDate = true
	case todo.Due != nil && (task.DueDate == nil || !todo.Due.Equal(*task.DueDate)):
		patch.DueDate = todo.Due
	}

	switch {
	case todo.CloseReason != nil && (task.CloseReason == nil || *task.CloseReason != *todo.CloseReason):
		patch.CloseReason = todo.CloseReason
	case todo.CloseReason == nil && task.Closed():
		reopen := false
		patch.Closed = &reopen
	}
	return patch
}

// checkPreconditions applies If-Match and If-None-Match, which go-webdav
// leaves to the backend. current is nil for an object that doesn't exist.
func checkPreconditions(options *gocaldav.PutCalendarObjectOptions, current *gocaldav.CalendarObject) error {
	etag := ""
	if current != nil {
		etag = current.ETag
	}
	if options.IfNoneMatch.IsSet() {
		if options.IfNoneMatch.IsWildcard() && current != nil {
			return errPreconditionFailed
		}
		if matched, _ := options.IfNoneMatch.MatchETag(etag); matched {
			return errPreconditionFailed
		}
	}
	if options.IfMatch.IsSet() {
		if matched, err := options.IfMatch.MatchETag(etag); err != nil || !matched {
			return errPreconditionFailed
		}
	}
	return nil
}

func (b *backend) deleteTimespan(ctx context.Context, name string) error {
	timespan, err := b.lookupTimespan(ctx, name)
	if err != nil {
		return err
	}
	return b.svc.DeleteTimespan(ctx, timespan.Id)
}

// deleteTask refuses tasks with subtasks, open or closed: deleting a task
// deletes its subtasks, which a calendar app shows as unrelated items.
func (b *backend) deleteTask(ctx context.Context, name string) error {
	task, err := b.lookupTask(ctx, name)
	if err != nil {
		return err
	}
	subtasks, err := b.svc.ListTasks(ctx, model.TaskListParams{
		PaginationParams: model.PaginationParams{Limit: 1},
		ParentId:         &task.Id,
		IncludeClosed:    true,
	}, nil)
	if err != nil {
		return err
	}
	if subtasks.TotalCount > 0 {
		return errHasSubtasks
	}
	return b.svc.DeleteTask(ctx, task.Id)
}
