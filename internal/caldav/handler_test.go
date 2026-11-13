package caldav_test

import (
	"context"
	"encoding/xml"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/caldav"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var now = time.Date(2026, 11, 12, 12, 0, 0, 0, time.UTC)

// fakeService answers the calls the backend makes; any other call panics on
// the nil embedded Service.
type fakeService struct {
	service.Service
	timespans []model.Timespan
	tasks     []model.Task
	tags      map[uuid.UUID]model.Tag
	tagLists  int
	from, to  *time.Time
	err       error
}

func (f *fakeService) GetTimespan(ctx context.Context, id uuid.UUID) (model.Timespan, error) {
	if f.err != nil {
		return model.Timespan{}, f.err
	}
	for _, timespan := range f.timespans {
		if timespan.Id == id {
			return timespan, nil
		}
	}
	return model.Timespan{}, model.ErrNotFound
}

func (f *fakeService) GetTask(ctx context.Context, id uuid.UUID, includes *service.TaskServiceIncludes) (model.Task, error) {
	for _, task := range f.tasks {
		if task.Id == id {
			return task, nil
		}
	}
	return model.Task{}, model.ErrNotFound
}

func (f *fakeService) ListCalendarTimespans(ctx context.Context, from, to *time.Time) ([]model.Timespan, error) {
	f.from, f.to = from, to
	return f.timespans, nil
}

func (f *fakeService) ListCalendarTasks(ctx context.Context, closedSince time.Time) ([]model.Task, error) {
	return f.tasks, nil
}

func (f *fakeService) ListCalendarTags(ctx context.Context) (map[uuid.UUID]model.Tag, error) {
	f.tagLists++
	return f.tags, nil
}

var (
	label  = model.Tag{Id: uuid.New(), Name: "work"}
	before = model.Timespan{
		Id: uuid.New(), Name: "before",
		StartTime: time.Date(2026, 11, 1, 8, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC),
	}
	overlapping = model.Timespan{
		Id: uuid.New(), Name: "overlapping", TagIds: []uuid.UUID{label.Id},
		StartTime: time.Date(2026, 11, 9, 23, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 11, 10, 1, 0, 0, 0, time.UTC),
	}
	done     = model.CloseReasonDone
	openTask = model.Task{Id: uuid.New(), Name: "open", TagIds: []uuid.UUID{label.Id}}
	oldTask  = model.Task{Id: uuid.New(), Name: "old", CloseReason: &done, ClosedAt: new(now.AddDate(0, 0, -31))}
)

func newHandler() (http.Handler, *fakeService) {
	svc := &fakeService{
		timespans: []model.Timespan{before, overlapping},
		tasks:     []model.Task{openTask},
		tags:      map[uuid.UUID]model.Tag{label.Id: label},
	}
	return caldav.NewHandler(svc, func() time.Time { return now }), svc
}

func serve(handler http.Handler, method, path, depth, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/xml; charset=utf-8")
	if depth != "" {
		request.Header.Set("Depth", depth)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

type multistatus struct {
	Responses []struct {
		Href      string `xml:"href"`
		Propstats []struct {
			Prop struct {
				Raw []struct {
					XMLName xml.Name
					Inner   string `xml:",innerxml"`
				} `xml:",any"`
				CalendarData string `xml:"urn:ietf:params:xml:ns:caldav calendar-data"`
			} `xml:"prop"`
			Status string `xml:"status"`
		} `xml:"propstat"`
	} `xml:"response"`
}

func decodeMultistatus(t *testing.T, recorder *httptest.ResponseRecorder) multistatus {
	t.Helper()
	require.Equal(t, http.StatusMultiStatus, recorder.Code, recorder.Body.String())
	var result multistatus
	require.NoError(t, xml.Unmarshal(recorder.Body.Bytes(), &result))
	return result
}

func (m multistatus) hrefs() []string {
	var hrefs []string
	for _, response := range m.Responses {
		hrefs = append(hrefs, response.Href)
	}
	return hrefs
}

func objectPath(collection string, id uuid.UUID) string {
	return "/dav/me/calendars/" + collection + "/" + id.String() + ".ics"
}

func TestHomeListsCollections(t *testing.T) {
	handler, _ := newHandler()

	recorder := serve(handler, "PROPFIND", "/dav/me/calendars/", "1", `<?xml version="1.0" encoding="UTF-8"?>
<d:propfind xmlns:d="DAV:" xmlns:c="urn:ietf:params:xml:ns:caldav">
  <d:prop><d:displayname/><d:resourcetype/><c:supported-calendar-component-set/></d:prop>
</d:propfind>`)

	result := decodeMultistatus(t, recorder)
	assert.ElementsMatch(t, []string{"/dav/me/calendars/", "/dav/me/calendars/time/", "/dav/me/calendars/tasks/"}, result.hrefs())
	body := recorder.Body.String()
	assert.Regexp(t, `(?s)/dav/me/calendars/time/</[^>]*href>.*name="VEVENT"`, body)
	assert.Regexp(t, `(?s)/dav/me/calendars/tasks/</[^>]*href>.*name="VTODO"`, body)
}

func TestCalendarQueryUsesTimeRange(t *testing.T) {
	handler, svc := newHandler()

	recorder := serve(handler, "REPORT", "/dav/me/calendars/time/", "1", `<?xml version="1.0" encoding="UTF-8"?>
<c:calendar-query xmlns:d="DAV:" xmlns:c="urn:ietf:params:xml:ns:caldav">
  <d:prop><d:getetag/></d:prop>
  <c:filter>
    <c:comp-filter name="VCALENDAR">
      <c:comp-filter name="VEVENT">
        <c:time-range start="20261110T000000Z" end="20261117T000000Z"/>
      </c:comp-filter>
    </c:comp-filter>
  </c:filter>
</c:calendar-query>`)

	result := decodeMultistatus(t, recorder)
	assert.Equal(t, []string{objectPath("time", overlapping.Id)}, result.hrefs())
	require.NotNil(t, svc.from)
	require.NotNil(t, svc.to)
	assert.Equal(t, time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC), svc.from.UTC())
	assert.Equal(t, time.Date(2026, 11, 17, 0, 0, 0, 0, time.UTC), svc.to.UTC())
}

func TestCalendarMultigetReturnsEncodedTask(t *testing.T) {
	handler, _ := newHandler()
	path := objectPath("tasks", openTask.Id)

	recorder := serve(handler, "REPORT", "/dav/me/calendars/tasks/", "1", `<?xml version="1.0" encoding="UTF-8"?>
<c:calendar-multiget xmlns:d="DAV:" xmlns:c="urn:ietf:params:xml:ns:caldav">
  <d:prop><d:getetag/><c:calendar-data/></d:prop>
  <d:href>`+path+`</d:href>
</c:calendar-multiget>`)

	result := decodeMultistatus(t, recorder)
	require.Equal(t, []string{path}, result.hrefs())
	want, _, err := caldav.Encode(caldav.TaskTodo(openTask, map[uuid.UUID]model.Tag{label.Id: label}))
	require.NoError(t, err)
	assert.Equal(t, string(want), result.Responses[0].Propstats[0].Prop.CalendarData)
}

func TestCalendarMultigetLoadsTagsOncePerRequest(t *testing.T) {
	handler, svc := newHandler()
	body := `<?xml version="1.0" encoding="utf-8"?>
<c:calendar-multiget xmlns:d="DAV:" xmlns:c="urn:ietf:params:xml:ns:caldav">
  <d:prop><d:getetag/><c:calendar-data/></d:prop>
  <d:href>/dav/me/calendars/time/` + before.Id.String() + `.ics</d:href>
  <d:href>/dav/me/calendars/time/` + overlapping.Id.String() + `.ics</d:href>
</c:calendar-multiget>`

	first := serve(handler, "REPORT", "/dav/me/calendars/time/", "1", body)
	require.Equal(t, http.StatusMultiStatus, first.Code)
	assert.Equal(t, 1, svc.tagLists)

	serve(handler, "REPORT", "/dav/me/calendars/time/", "1", body)
	assert.Equal(t, 2, svc.tagLists)
}

func TestGetReturnsEncodedETag(t *testing.T) {
	handler, _ := newHandler()

	recorder := serve(handler, http.MethodGet, objectPath("time", overlapping.Id), "", "")

	require.Equal(t, http.StatusOK, recorder.Code)
	want, etag, err := caldav.Encode(caldav.TimespanEvent(overlapping, map[uuid.UUID]model.Tag{label.Id: label}))
	require.NoError(t, err)
	assert.Equal(t, `"`+etag+`"`, recorder.Header().Get("ETag"))
	assert.Equal(t, string(want), recorder.Body.String())
}

func TestGetMissingObjects(t *testing.T) {
	handler, _ := newHandler()

	tests := map[string]string{
		"task closed 31 days ago": objectPath("tasks", oldTask.Id),
		"malformed id":            "/dav/me/calendars/time/not-a-uuid.ics",
		"missing suffix":          "/dav/me/calendars/time/" + overlapping.Id.String(),
		"undashed id":             "/dav/me/calendars/time/" + strings.ReplaceAll(overlapping.Id.String(), "-", "") + ".ics",
		"urn id":                  "/dav/me/calendars/time/urn:uuid:" + overlapping.Id.String() + ".ics",
		"braced id":               "/dav/me/calendars/time/{" + overlapping.Id.String() + "}.ics",
		"unknown timespan":        objectPath("time", uuid.New()),
		"unknown collection":      objectPath("other", overlapping.Id),
	}
	for name, path := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, http.StatusNotFound, serve(handler, http.MethodGet, path, "", "").Code)
		})
	}
}

func TestGetHidesInternalErrors(t *testing.T) {
	handler, svc := newHandler()
	svc.err = errors.New("database password is hunter2")

	recorder := serve(handler, http.MethodGet, objectPath("time", overlapping.Id), "", "")

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "hunter2")
}

func TestWritesAreForbidden(t *testing.T) {
	handler, _ := newHandler()
	path := objectPath("tasks", openTask.Id)

	body, _, err := caldav.Encode(caldav.TaskTodo(model.Task{Id: openTask.Id, Name: "changed"}, nil))
	require.NoError(t, err)
	put := httptest.NewRequest(http.MethodPut, path, strings.NewReader(string(body)))
	put.Header.Set("Content-Type", "text/calendar")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, put)
	assert.Equal(t, http.StatusForbidden, recorder.Code)

	assert.Equal(t, http.StatusForbidden, serve(handler, http.MethodDelete, path, "", "").Code)
}

func TestPropPatchIsAccepted(t *testing.T) {
	handler, _ := newHandler()

	recorder := serve(handler, "PROPPATCH", "/dav/me/calendars/time/", "", `<?xml version="1.0" encoding="UTF-8"?>
<d:propertyupdate xmlns:d="DAV:" xmlns:a="http://apple.com/ns/ical/">
  <d:set><d:prop>
    <d:displayname>Renamed</d:displayname>
    <a:calendar-color>#FF0000FF</a:calendar-color>
  </d:prop></d:set>
</d:propertyupdate>`)

	result := decodeMultistatus(t, recorder)
	assert.Equal(t, "application/xml; charset=utf-8", recorder.Header().Get("Content-Type"))
	require.Len(t, result.Responses, 1)
	assert.Equal(t, "/dav/me/calendars/time/", result.Responses[0].Href)
	require.Len(t, result.Responses[0].Propstats, 1)
	propstat := result.Responses[0].Propstats[0]
	assert.Equal(t, "HTTP/1.1 200 OK", propstat.Status)
	var names []xml.Name
	for _, property := range propstat.Prop.Raw {
		names = append(names, property.XMLName)
		assert.Empty(t, property.Inner)
	}
	assert.Equal(t, []xml.Name{{Space: "DAV:", Local: "displayname"}, {Space: "http://apple.com/ns/ical/", Local: "calendar-color"}}, names)
}

func TestPropPatchRejectsMalformedBody(t *testing.T) {
	handler, _ := newHandler()

	assert.Equal(t, http.StatusBadRequest, serve(handler, "PROPPATCH", "/dav/me/calendars/time/", "", "<not-xml").Code)
}
