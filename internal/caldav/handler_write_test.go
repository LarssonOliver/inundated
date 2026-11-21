package caldav_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const newName = "7F3A1C2E-5B4D-4E6F-8A9B-0C1D2E3F4A5B.ics"

var untitledLabelled = model.Timespan{
	Id: uuid.New(), TagIds: []uuid.UUID{label.Id},
	StartTime: time.Date(2026, 11, 3, 8, 0, 0, 0, time.UTC),
	EndTime:   time.Date(2026, 11, 3, 9, 0, 0, 0, time.UTC),
}

func fixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name+".ics"))
	require.NoError(t, err)
	return string(data)
}

func eventBody(summary, start, end string) string {
	return strings.ReplaceAll("BEGIN:VCALENDAR\nVERSION:2.0\nPRODID:-//test//EN\nBEGIN:VEVENT\nUID:x\nDTSTAMP:20261121T000000Z\n"+
		"SUMMARY:"+summary+"\nDTSTART:"+start+"\nDTEND:"+end+"\nEND:VEVENT\nEND:VCALENDAR\n", "\n", "\r\n")
}

func todoBody(lines string) string {
	return strings.ReplaceAll("BEGIN:VCALENDAR\nVERSION:2.0\nPRODID:-//test//EN\nBEGIN:VTODO\nUID:x\nDTSTAMP:20261121T000000Z\n"+
		lines+"END:VTODO\nEND:VCALENDAR\n", "\n", "\r\n")
}

func put(handler http.Handler, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPut, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "text/calendar; charset=utf-8")
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func TestPutCreatesTimespan(t *testing.T) {
	handler, svc := newHandler()

	recorder := put(handler, "/dav/me/calendars/time/"+newName, fixture(t, "apple_event"), map[string]string{"If-None-Match": "*"})

	require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())
	assert.Empty(t, recorder.Header().Get("ETag"))
	require.Len(t, svc.writes, 1)
	created := svc.writes[0].(model.Timespan)
	stockholm, err := time.LoadLocation("Europe/Stockholm")
	require.NoError(t, err)
	assert.Equal(t, "Gym", created.Name)
	assert.True(t, time.Date(2026, 11, 21, 9, 0, 0, 0, stockholm).Equal(created.StartTime))
	assert.True(t, time.Date(2026, 11, 21, 10, 30, 0, 0, stockholm).Equal(created.EndTime))
	require.NotNil(t, created.CaldavName)
	assert.Equal(t, newName, *created.CaldavName)
}

func TestPutCreatesUntitledTimespan(t *testing.T) {
	handler, svc := newHandler()

	recorder := put(handler, "/dav/me/calendars/time/"+newName, eventBody("Untitled", "20261121T090000Z", "20261121T100000Z"), nil)

	require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())
	require.Len(t, svc.writes, 1)
	assert.Equal(t, "", svc.writes[0].(model.Timespan).Name)
}

func TestPutUpdateKeepsTagsAndDerivedTitle(t *testing.T) {
	handler, svc := newHandler()
	svc.timespans = append(svc.timespans, untitledLabelled)

	recorder := put(handler, objectPath("time", untitledLabelled.Id), eventBody("work", "20261103T080000Z", "20261103T093000Z"), nil)

	require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())
	assert.Empty(t, recorder.Header().Get("ETag"))
	require.Len(t, svc.writes, 1)
	updated := svc.writes[0].(model.Timespan)
	assert.Equal(t, untitledLabelled.Id, updated.Id)
	assert.Equal(t, "", updated.Name)
	assert.Equal(t, []uuid.UUID{label.Id}, updated.TagIds)
	assert.Equal(t, time.Date(2026, 11, 3, 9, 30, 0, 0, time.UTC), updated.EndTime.UTC())
}

func TestPutUpdateRenamesTimespan(t *testing.T) {
	handler, svc := newHandler()
	svc.timespans = append(svc.timespans, untitledLabelled)

	recorder := put(handler, objectPath("time", untitledLabelled.Id), eventBody("Planning", "20261103T080000Z", "20261103T090000Z"), nil)

	require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())
	require.Len(t, svc.writes, 1)
	assert.Equal(t, "Planning", svc.writes[0].(model.Timespan).Name)
}

func TestPutRefusesAllDayAndRecurring(t *testing.T) {
	for _, name := range []string{"apple_event_allday", "apple_event_recurring"} {
		t.Run(name, func(t *testing.T) {
			handler, svc := newHandler()

			recorder := put(handler, "/dav/me/calendars/time/"+newName, fixture(t, name), nil)

			assert.Equal(t, http.StatusForbidden, recorder.Code)
			assert.Empty(t, svc.writes)
		})
	}
}

func TestPutRefusesWrongComponent(t *testing.T) {
	handler, svc := newHandler()

	recorder := put(handler, "/dav/me/calendars/tasks/"+newName, fixture(t, "apple_event"), nil)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Empty(t, svc.writes)
}

func TestPutPreconditions(t *testing.T) {
	path := objectPath("time", overlapping.Id)
	body := eventBody("overlapping", "20261109T230000Z", "20261110T020000Z")

	t.Run("create over an existing object", func(t *testing.T) {
		handler, svc := newHandler()
		assert.Equal(t, http.StatusPreconditionFailed, put(handler, path, body, map[string]string{"If-None-Match": "*"}).Code)
		assert.Empty(t, svc.writes)
	})
	t.Run("stale etag", func(t *testing.T) {
		handler, svc := newHandler()
		assert.Equal(t, http.StatusPreconditionFailed, put(handler, path, body, map[string]string{"If-Match": `"stale"`}).Code)
		assert.Empty(t, svc.writes)
	})
	t.Run("etag of a missing object", func(t *testing.T) {
		handler, svc := newHandler()
		assert.Equal(t, http.StatusPreconditionFailed, put(handler, "/dav/me/calendars/time/"+newName, body, map[string]string{"If-Match": `"abc"`}).Code)
		assert.Empty(t, svc.writes)
	})
	t.Run("current etag", func(t *testing.T) {
		handler, svc := newHandler()
		etag := serve(handler, http.MethodGet, path, "", "").Header().Get("ETag")
		require.NotEmpty(t, etag)
		assert.Equal(t, http.StatusCreated, put(handler, path, body, map[string]string{"If-Match": etag}).Code)
		assert.Len(t, svc.writes, 1)
	})
}

func TestPutCreateConflict(t *testing.T) {
	handler, svc := newHandler()
	svc.writeErr = model.ErrAlreadyExists

	recorder := put(handler, "/dav/me/calendars/time/"+newName, fixture(t, "apple_event"), nil)

	assert.Equal(t, http.StatusPreconditionFailed, recorder.Code)
}

func TestPutCreatesTask(t *testing.T) {
	handler, svc := newHandler()

	recorder := put(handler, "/dav/me/calendars/tasks/"+newName, fixture(t, "apple_todo"), nil)

	require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())
	assert.Empty(t, recorder.Header().Get("ETag"))
	require.Len(t, svc.writes, 1)
	created := svc.writes[0].(model.Task)
	assert.Equal(t, "Buy milk", created.Name)
	require.NotNil(t, created.DueDate)
	assert.Equal(t, time.Date(2026, 11, 24, 0, 0, 0, 0, time.UTC), *created.DueDate)
	assert.Nil(t, created.ParentId)
	require.NotNil(t, created.CaldavName)
	assert.Equal(t, newName, *created.CaldavName)
}

func TestPutCreatesUntitledTask(t *testing.T) {
	handler, svc := newHandler()

	recorder := put(handler, "/dav/me/calendars/tasks/"+newName, todoBody(""), nil)

	require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())
	require.Len(t, svc.writes, 1)
	assert.Equal(t, "Untitled", svc.writes[0].(model.Task).Name)
}

func TestPutCompletesAndReopensTask(t *testing.T) {
	doneTask := model.Task{Id: uuid.New(), Name: "done", CloseReason: &done, ClosedAt: new(now.AddDate(0, 0, -1))}

	t.Run("complete", func(t *testing.T) {
		handler, svc := newHandler()
		recorder := put(handler, objectPath("tasks", openTask.Id), todoBody("SUMMARY:open\nSTATUS:COMPLETED\n"), nil)
		require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())
		require.Len(t, svc.writes, 1)
		update := svc.writes[0].(taskUpdate)
		assert.Equal(t, openTask.Id, update.id)
		assert.Equal(t, model.TaskPatch{CloseReason: &done}, update.patch)
	})
	t.Run("reopen", func(t *testing.T) {
		handler, svc := newHandler()
		svc.tasks = append(svc.tasks, doneTask)
		recorder := put(handler, objectPath("tasks", doneTask.Id), todoBody("SUMMARY:done\nSTATUS:NEEDS-ACTION\n"), nil)
		require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())
		require.Len(t, svc.writes, 1)
		assert.Equal(t, model.TaskPatch{Closed: new(false)}, svc.writes[0].(taskUpdate).patch)
	})
	t.Run("unchanged", func(t *testing.T) {
		handler, svc := newHandler()
		recorder := put(handler, objectPath("tasks", openTask.Id), todoBody("SUMMARY:open\nCATEGORIES:other\n"), nil)
		require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())
		assert.Empty(t, svc.writes)
	})
	t.Run("rename and due date", func(t *testing.T) {
		handler, svc := newHandler()
		recorder := put(handler, objectPath("tasks", openTask.Id), todoBody("SUMMARY:renamed\nDUE;VALUE=DATE:20261130\n"), nil)
		require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())
		require.Len(t, svc.writes, 1)
		due := time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, model.TaskPatch{Name: new("renamed"), DueDate: &due}, svc.writes[0].(taskUpdate).patch)
	})
	t.Run("clear due date", func(t *testing.T) {
		handler, svc := newHandler()
		withDue := openTask
		withDue.DueDate = new(time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC))
		svc.tasks = []model.Task{withDue}
		recorder := put(handler, objectPath("tasks", openTask.Id), todoBody("SUMMARY:open\n"), nil)
		require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())
		require.Len(t, svc.writes, 1)
		assert.Equal(t, model.TaskPatch{ClearDueDate: true}, svc.writes[0].(taskUpdate).patch)
	})
}

func TestDeleteTimespan(t *testing.T) {
	handler, svc := newHandler()

	recorder := serve(handler, http.MethodDelete, objectPath("time", overlapping.Id), "", "")

	assert.Equal(t, http.StatusNoContent, recorder.Code)
	assert.Equal(t, []any{timespanDelete(overlapping.Id)}, svc.writes)
}

func TestDeleteTaskWithoutSubtasks(t *testing.T) {
	handler, svc := newHandler()

	recorder := serve(handler, http.MethodDelete, objectPath("tasks", openTask.Id), "", "")

	assert.Equal(t, http.StatusNoContent, recorder.Code)
	assert.Equal(t, []any{taskDelete(openTask.Id)}, svc.writes)
}

func TestDeleteTaskWithSubtasks(t *testing.T) {
	handler, svc := newHandler()
	svc.subtasks = 1

	recorder := serve(handler, http.MethodDelete, objectPath("tasks", openTask.Id), "", "")

	assert.Equal(t, http.StatusForbidden, recorder.Code)
	assert.Empty(t, svc.writes)
}

func TestDeleteMissingObject(t *testing.T) {
	handler, svc := newHandler()

	recorder := serve(handler, http.MethodDelete, "/dav/me/calendars/time/"+newName, "", "")

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	assert.Empty(t, svc.writes)
}

func TestWriteErrorsHideInternals(t *testing.T) {
	handler, svc := newHandler()
	svc.writeErr = errors.New("database password is hunter2")

	recorder := put(handler, "/dav/me/calendars/time/"+newName, fixture(t, "apple_event"), nil)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "hunter2")
}

func TestPutToCollectionIsRefused(t *testing.T) {
	for _, collection := range []string{"/dav/me/calendars/time/", "/dav/me/calendars/tasks/"} {
		t.Run(collection, func(t *testing.T) {
			handler, svc := newHandler()

			recorder := put(handler, collection, fixture(t, "apple_event"), nil)

			assert.Equal(t, http.StatusMethodNotAllowed, recorder.Code)
			assert.Empty(t, svc.writes)
		})
	}
}

func TestPutCreatesClosedTask(t *testing.T) {
	ignored := model.CloseReasonIgnored
	tests := map[string]struct {
		body string
		want *model.CloseReason
	}{
		"completed": {fixture(t, "apple_todo_completed"), &done},
		"cancelled": {todoBody("SUMMARY:Skip\nSTATUS:CANCELLED\n"), &ignored},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			handler, svc := newHandler()

			recorder := put(handler, "/dav/me/calendars/tasks/"+newName, test.body, nil)

			require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())
			require.Len(t, svc.writes, 2)
			assert.IsType(t, model.Task{}, svc.writes[0])
			update := svc.writes[1].(taskUpdate)
			assert.NotEqual(t, uuid.Nil, update.id)
			assert.Equal(t, model.TaskPatch{CloseReason: test.want}, update.patch)
		})
	}
}
