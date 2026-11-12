// Package caldav exposes timespans and tasks as read-only CalDAV resources.
package caldav

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"slices"
	"strings"
	"time"

	"github.com/emersion/go-ical"
	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

// Constant so that an object's bytes, and so its ETag, change only when the
// object does.
var timestamp = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

const untitled = "Untitled"

// TimespanEvent maps a timespan to a calendar holding one VEVENT. tags holds
// every tag the timespan may reference, keyed by id.
func TimespanEvent(timespan model.Timespan, tags map[uuid.UUID]model.Tag) *ical.Calendar {
	event := ical.NewEvent()
	event.Props.SetText(ical.PropUID, timespan.Id.String())
	event.Props.SetDateTime(ical.PropDateTimeStamp, timestamp)
	event.Props.SetDateTime(ical.PropDateTimeStart, timespan.StartTime.UTC())
	event.Props.SetDateTime(ical.PropDateTimeEnd, timespan.EndTime.UTC())

	tagsOfTimespan := lookupTags(timespan.TagIds, tags)
	summary := timespan.Name
	if summary == "" {
		summary = strings.Join(tagNames(tagsOfTimespan, func(tag model.Tag) bool { return tag.Owner != nil }), ", ")
	}
	if summary == "" {
		summary = untitled
	}
	event.Props.SetText(ical.PropSummary, summary)
	setCategories(event.Component, tagsOfTimespan)

	return newCalendar(event.Component)
}

// TaskTodo maps a task to a calendar holding one VTODO. tags holds every tag
// the task may reference, keyed by id.
func TaskTodo(task model.Task, tags map[uuid.UUID]model.Tag) *ical.Calendar {
	todo := ical.NewComponent(ical.CompToDo)
	todo.Props.SetText(ical.PropUID, task.Id.String())
	todo.Props.SetDateTime(ical.PropDateTimeStamp, timestamp)
	todo.Props.SetText(ical.PropSummary, task.Name)
	setCategories(todo, lookupTags(task.TagIds, tags))

	if task.DueDate != nil {
		todo.Props.SetDate(ical.PropDue, *task.DueDate)
	}
	if task.ParentId != nil {
		related := ical.NewProp(ical.PropRelatedTo)
		related.Value = task.ParentId.String()
		related.Params.Set(ical.ParamRelationshipType, "PARENT")
		todo.Props.Set(related)
	}

	switch {
	case task.CloseReason == nil:
		todo.Props.SetText(ical.PropStatus, "NEEDS-ACTION")
	case *task.CloseReason == model.CloseReasonDone:
		todo.Props.SetText(ical.PropStatus, "COMPLETED")
		if task.ClosedAt != nil {
			todo.Props.SetDateTime(ical.PropCompleted, task.ClosedAt.UTC())
		}
	default:
		todo.Props.SetText(ical.PropStatus, "CANCELLED")
	}

	return newCalendar(todo)
}

// Encode serializes cal and returns its bytes and their hex SHA-256 as the ETag (unquoted).
func Encode(cal *ical.Calendar) ([]byte, string, error) {
	var buffer bytes.Buffer
	if err := ical.NewEncoder(&buffer).Encode(cal); err != nil {
		return nil, "", err
	}
	sum := sha256.Sum256(buffer.Bytes())
	return buffer.Bytes(), hex.EncodeToString(sum[:]), nil
}

func newCalendar(component *ical.Component) *ical.Calendar {
	cal := ical.NewCalendar()
	cal.Props.SetText(ical.PropVersion, "2.0")
	cal.Props.SetText(ical.PropProductID, "-//inundated//caldav//EN")
	cal.Children = append(cal.Children, component)
	return cal
}

// lookupTags skips ids missing from tags.
func lookupTags(ids []uuid.UUID, tags map[uuid.UUID]model.Tag) []model.Tag {
	found := make([]model.Tag, 0, len(ids))
	for _, tagId := range ids {
		if tag, ok := tags[tagId]; ok {
			found = append(found, tag)
		}
	}
	return found
}

func tagNames(tags []model.Tag, include func(model.Tag) bool) []string {
	var names []string
	for _, tag := range tags {
		if include(tag) {
			names = append(names, tag.Name)
		}
	}
	slices.Sort(names)
	return names
}

func setCategories(component *ical.Component, tags []model.Tag) {
	names := tagNames(tags, func(tag model.Tag) bool { return tag.Owner == nil })
	if len(names) == 0 {
		return
	}
	categories := ical.NewProp(ical.PropCategories)
	categories.SetTextList(names)
	component.Props.Set(categories)
}
