// Package caldav exposes timespans and tasks as read-only CalDAV resources.
package caldav

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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

// TimespanEvent maps a timespan to a calendar holding one VEVENT, with its
// times in location. tags holds every tag the timespan may reference, keyed
// by id.
func TimespanEvent(timespan model.Timespan, tags map[uuid.UUID]model.Tag, location *time.Location) *ical.Calendar {
	start := timespan.StartTime.In(location)
	end := timespan.EndTime.In(location)

	event := ical.NewEvent()
	event.Props.SetText(ical.PropUID, timespan.Id.String())
	event.Props.SetDateTime(ical.PropDateTimeStamp, timestamp)
	event.Props.SetDateTime(ical.PropDateTimeStart, start)
	event.Props.SetDateTime(ical.PropDateTimeEnd, end)

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

	cal := newCalendar(event.Component)
	if location != time.UTC {
		cal.Children = slices.Insert(cal.Children, 0, timezone(start, end))
	}
	return cal
}

// timezone describes the zone of the given times, all in one location, with
// only the observances in effect at those times. RFC 5545 requires a
// VTIMEZONE for every TZID, but not one covering all of history.
func timezone(times ...time.Time) *ical.Component {
	component := ical.NewComponent(ical.CompTimezone)
	component.Props.SetText(ical.PropTimezoneID, times[0].Location().String())

	var starts []time.Time
	for _, t := range times {
		start := zoneStart(t)
		if slices.ContainsFunc(starts, start.Equal) {
			continue
		}
		starts = append(starts, start)
		component.Children = append(component.Children, observance(t, start))
	}
	return component
}

// zoneStart returns when the zone in effect at t began, or zero if the
// location has always used it. With slim tzdata, as in Go's own zoneinfo.zip,
// ZoneBounds stops at year boundaries the zone continues across, so keep
// stepping back until the zone actually changes.
func zoneStart(t time.Time) time.Time {
	name, offset := t.Zone()
	start, _ := t.ZoneBounds()
	for !start.IsZero() {
		before := start.Add(-time.Second)
		if beforeName, beforeOffset := before.Zone(); beforeName != name || beforeOffset != offset {
			break
		}
		start, _ = before.ZoneBounds()
	}
	return start
}

// observance describes the zone in effect at t, which began at start (zero
// if the location has always used it).
func observance(t, start time.Time) *ical.Component {
	name, offset := t.Zone()
	offsetFrom := offset
	if start.IsZero() {
		start = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	} else {
		_, offsetFrom = start.Add(-time.Second).Zone()
	}

	kind := ical.CompTimezoneStandard
	if t.IsDST() {
		kind = ical.CompTimezoneDaylight
	}
	component := ical.NewComponent(kind)
	// An observance's DTSTART is local time under the offset it replaces.
	onset := ical.NewProp(ical.PropDateTimeStart)
	onset.Value = start.In(time.FixedZone("", offsetFrom)).Format("20060102T150405")
	component.Props.Set(onset)
	setOffset(component, ical.PropTimezoneOffsetFrom, offsetFrom)
	setOffset(component, ical.PropTimezoneOffsetTo, offset)
	component.Props.SetText(ical.PropTimezoneName, name)
	return component
}

// setOffset sets the value directly, since SetText would mark the
// UTC-OFFSET property as TEXT.
func setOffset(component *ical.Component, name string, seconds int) {
	prop := ical.NewProp(name)
	prop.Value = formatOffset(seconds)
	component.Props.Set(prop)
}

func formatOffset(seconds int) string {
	sign := '+'
	if seconds < 0 {
		sign = '-'
		seconds = -seconds
	}
	formatted := fmt.Sprintf("%c%02d%02d", sign, seconds/3600, seconds%3600/60)
	if seconds%60 != 0 {
		formatted += fmt.Sprintf("%02d", seconds%60)
	}
	return formatted
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
