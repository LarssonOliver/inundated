package caldav

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/emersion/go-ical"
	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func stockholm(t *testing.T) *time.Location {
	t.Helper()
	location, err := time.LoadLocation("Europe/Stockholm")
	require.NoError(t, err)
	return location
}

func readCalendar(t *testing.T, name string) *ical.Calendar {
	t.Helper()
	file, err := os.Open(filepath.Join("testdata", name+".ics"))
	require.NoError(t, err)
	defer func() { _ = file.Close() }()
	cal, err := ical.NewDecoder(file).Decode()
	require.NoError(t, err)
	return cal
}

// decodeCalendar parses inline iCalendar written with "\n" line endings.
func decodeCalendar(t *testing.T, text string) *ical.Calendar {
	t.Helper()
	cal, err := ical.NewDecoder(strings.NewReader(strings.ReplaceAll(text, "\n", "\r\n"))).Decode()
	require.NoError(t, err)
	return cal
}

func eventCalendar(t *testing.T, lines string) *ical.Calendar {
	return decodeCalendar(t, "BEGIN:VCALENDAR\nVERSION:2.0\nPRODID:-//test//EN\nBEGIN:VEVENT\nUID:x\nDTSTAMP:20261121T000000Z\n"+lines+"END:VEVENT\nEND:VCALENDAR\n")
}

func todoCalendar(t *testing.T, lines string) *ical.Calendar {
	return decodeCalendar(t, "BEGIN:VCALENDAR\nVERSION:2.0\nPRODID:-//test//EN\nBEGIN:VTODO\nUID:x\nDTSTAMP:20261121T000000Z\nSUMMARY:Buy milk\n"+lines+"END:VTODO\nEND:VCALENDAR\n")
}

func TestParseEvent(t *testing.T) {
	location := stockholm(t)

	t.Run("apple event", func(t *testing.T) {
		event, err := ParseEvent(readCalendar(t, "apple_event"), time.UTC)
		require.NoError(t, err)
		assert.Equal(t, "Gym", event.Summary)
		assert.True(t, time.Date(2026, 11, 21, 9, 0, 0, 0, location).Equal(event.Start), event.Start)
		assert.True(t, time.Date(2026, 11, 21, 10, 30, 0, 0, location).Equal(event.End), event.End)
	})

	t.Run("floating times use the given location", func(t *testing.T) {
		event, err := ParseEvent(readCalendar(t, "apple_event_floating"), location)
		require.NoError(t, err)
		assert.True(t, time.Date(2026, 11, 21, 9, 0, 0, 0, location).Equal(event.Start), event.Start)
	})

	t.Run("duration instead of end", func(t *testing.T) {
		event, err := ParseEvent(eventCalendar(t, "DTSTART:20261121T090000Z\nDURATION:PT45M\n"), time.UTC)
		require.NoError(t, err)
		assert.Equal(t, 45*time.Minute, event.End.Sub(event.Start))
	})

	t.Run("missing summary", func(t *testing.T) {
		event, err := ParseEvent(eventCalendar(t, "DTSTART:20261121T090000Z\nDTEND:20261121T100000Z\n"), time.UTC)
		require.NoError(t, err)
		assert.Equal(t, "", event.Summary)
	})

	unsupported := map[string]*ical.Calendar{
		"all-day":   readCalendar(t, "apple_event_allday"),
		"recurring": readCalendar(t, "apple_event_recurring"),
		"rdate":     eventCalendar(t, "DTSTART:20261121T090000Z\nDTEND:20261121T100000Z\nRDATE:20261128T090000Z\n"),
	}
	for name, cal := range unsupported {
		t.Run(name, func(t *testing.T) {
			_, err := ParseEvent(cal, time.UTC)
			assert.ErrorIs(t, err, errUnsupported)
		})
	}

	invalid := map[string]*ical.Calendar{
		"no event":         todoCalendar(t, ""),
		"missing start":    eventCalendar(t, "DTEND:20261121T100000Z\n"),
		"missing end":      eventCalendar(t, "DTSTART:20261121T090000Z\n"),
		"end before start": eventCalendar(t, "DTSTART:20261121T090000Z\nDTEND:20261121T080000Z\n"),
		"empty":            eventCalendar(t, "DTSTART:20261121T090000Z\nDTEND:20261121T090000Z\n"),
		"unknown zone":     eventCalendar(t, "DTSTART;TZID=Nowhere/Special:20261121T090000\nDTEND;TZID=Nowhere/Special:20261121T100000\n"),
	}
	for name, cal := range invalid {
		t.Run(name, func(t *testing.T) {
			_, err := ParseEvent(cal, time.UTC)
			assert.ErrorIs(t, err, errInvalid)
		})
	}
}

func TestParseTodo(t *testing.T) {
	location := stockholm(t)
	due := time.Date(2026, 11, 24, 0, 0, 0, 0, time.UTC)
	ignored := model.CloseReasonIgnored

	t.Run("apple todo", func(t *testing.T) {
		todo, err := ParseTodo(readCalendar(t, "apple_todo"), time.UTC)
		require.NoError(t, err)
		assert.Equal(t, "Buy milk", todo.Summary)
		require.NotNil(t, todo.Due)
		assert.Equal(t, due, *todo.Due)
		assert.Nil(t, todo.CloseReason)
	})

	t.Run("apple completed todo", func(t *testing.T) {
		todo, err := ParseTodo(readCalendar(t, "apple_todo_completed"), time.UTC)
		require.NoError(t, err)
		require.NotNil(t, todo.CloseReason)
		assert.Equal(t, model.CloseReasonDone, *todo.CloseReason)
	})

	t.Run("due date", func(t *testing.T) {
		todo, err := ParseTodo(todoCalendar(t, "DUE;VALUE=DATE:20261124\n"), location)
		require.NoError(t, err)
		require.NotNil(t, todo.Due)
		assert.Equal(t, due, *todo.Due)
	})

	t.Run("utc due is dated in the given location", func(t *testing.T) {
		todo, err := ParseTodo(todoCalendar(t, "DUE:20261124T233000Z\n"), location)
		require.NoError(t, err)
		require.NotNil(t, todo.Due)
		assert.Equal(t, due.AddDate(0, 0, 1), *todo.Due)
	})

	t.Run("no due", func(t *testing.T) {
		todo, err := ParseTodo(todoCalendar(t, ""), location)
		require.NoError(t, err)
		assert.Nil(t, todo.Due)
	})

	statuses := map[string]*model.CloseReason{
		"STATUS:CANCELLED\n":    &ignored,
		"STATUS:IN-PROCESS\n":   nil,
		"STATUS:NEEDS-ACTION\n": nil,
		"":                      nil,
	}
	for lines, want := range statuses {
		t.Run("status "+strings.TrimSpace(lines), func(t *testing.T) {
			todo, err := ParseTodo(todoCalendar(t, lines), time.UTC)
			require.NoError(t, err)
			assert.Equal(t, want, todo.CloseReason)
		})
	}

	t.Run("no todo", func(t *testing.T) {
		_, err := ParseTodo(eventCalendar(t, "DTSTART:20261121T090000Z\nDTEND:20261121T100000Z\n"), time.UTC)
		assert.ErrorIs(t, err, errInvalid)
	})
}

func TestTimespanSummary(t *testing.T) {
	tests := map[string]struct {
		timespan model.Timespan
		want     string
	}{
		"named":      {model.Timespan{Name: "Deep work", TagIds: []uuid.UUID{work.Id}}, "Deep work"},
		"owned tags": {model.Timespan{TagIds: []uuid.UUID{work.Id, taskTag.Id, project.Id}}, "Apollo, Write report"},
		"labels":     {model.Timespan{TagIds: []uuid.UUID{work.Id, alpha.Id}}, "alpha"},
		"untitled":   {model.Timespan{}, "Untitled"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.want, timespanSummary(test.timespan, tags))
		})
	}
}
