package caldav

import (
	"errors"
	"net/http"
	"time"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav"
	"github.com/larssonoliver/inundated/internal/model"
)

var (
	errUnsupported = webdav.NewHTTPError(http.StatusForbidden, errors.New("caldav: recurring and all-day events are not supported"))
	errInvalid     = webdav.NewHTTPError(http.StatusBadRequest, errors.New("caldav: invalid calendar object"))
)

// Event holds what a client may change on a timespan.
type Event struct {
	Summary    string
	Start, End time.Time
}

// Todo holds what a client may change on a task.
type Todo struct {
	Summary string
	// Due is midnight UTC of the due date, as [model.Task.DueDate].
	Due *time.Time
	// CloseReason is nil while the to-do is open.
	CloseReason *model.CloseReason
}

// ParseEvent reads the VEVENT in cal. location reads floating times.
func ParseEvent(cal *ical.Calendar, location *time.Location) (Event, error) {
	component := child(cal, ical.CompEvent)
	if component == nil {
		return Event{}, errInvalid
	}
	props := component.Props
	if props.Get(ical.PropRecurrenceRule) != nil || props.Get(ical.PropRecurrenceDates) != nil {
		return Event{}, errUnsupported
	}

	startProp := props.Get(ical.PropDateTimeStart)
	if startProp == nil {
		return Event{}, errInvalid
	}
	if isDate(startProp) {
		return Event{}, errUnsupported
	}
	start, err := startProp.DateTime(location)
	if err != nil {
		return Event{}, errInvalid
	}

	var end time.Time
	if endProp := props.Get(ical.PropDateTimeEnd); endProp != nil {
		if end, err = endProp.DateTime(location); err != nil {
			return Event{}, errInvalid
		}
	} else if durationProp := props.Get(ical.PropDuration); durationProp != nil {
		duration, err := durationProp.Duration()
		if err != nil {
			return Event{}, errInvalid
		}
		end = start.Add(duration)
	}
	if !end.After(start) {
		return Event{}, errInvalid
	}

	return Event{Summary: text(props, ical.PropSummary), Start: start, End: end}, nil
}

// ParseTodo reads the VTODO in cal. location dates due times.
func ParseTodo(cal *ical.Calendar, location *time.Location) (Todo, error) {
	component := child(cal, ical.CompToDo)
	if component == nil {
		return Todo{}, errInvalid
	}
	props := component.Props
	todo := Todo{Summary: text(props, ical.PropSummary)}

	if dueProp := props.Get(ical.PropDue); dueProp != nil {
		due, err := dueProp.DateTime(location)
		if err != nil {
			return Todo{}, errInvalid
		}
		if !isDate(dueProp) {
			due = due.In(location)
		}
		date := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, time.UTC)
		todo.Due = &date
	}

	var reason model.CloseReason
	switch text(props, ical.PropStatus) {
	case "COMPLETED":
		reason = model.CloseReasonDone
		todo.CloseReason = &reason
	case "CANCELLED":
		reason = model.CloseReasonIgnored
		todo.CloseReason = &reason
	}
	return todo, nil
}

func child(cal *ical.Calendar, name string) *ical.Component {
	for _, component := range cal.Children {
		if component.Name == name {
			return component
		}
	}
	return nil
}

// isDate reports a DATE value, marked or recognised by its length as
// go-ical does.
func isDate(prop *ical.Prop) bool {
	return prop.ValueType() == ical.ValueDate || len(prop.Value) == len("20060102")
}

func text(props ical.Props, name string) string {
	value, err := props.Text(name)
	if err != nil {
		return ""
	}
	return value
}
