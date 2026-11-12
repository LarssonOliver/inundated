package caldav

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav"
	gocaldav "github.com/emersion/go-webdav/caldav"
	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/service"
)

// go-webdav tells resources apart by path depth: principal, home, collection,
// object.
const (
	prefix        = "/dav"
	principalPath = prefix + "/me/"
	homePath      = principalPath + "calendars/"
	timePath      = homePath + "time/"
	tasksPath     = homePath + "tasks/"
	objectSuffix  = ".ics"
)

const closedTaskWindow = 30 * 24 * time.Hour

var (
	errReadOnly = webdav.NewHTTPError(http.StatusForbidden, errors.New("caldav: read-only"))
	errNotFound = webdav.NewHTTPError(http.StatusNotFound, nil)
	errInternal = webdav.NewHTTPError(http.StatusInternalServerError, errors.New("internal error"))
)

var calendars = []gocaldav.Calendar{
	{Path: timePath, Name: "Time", SupportedComponentSet: []string{ical.CompEvent}},
	{Path: tasksPath, Name: "Tasks", SupportedComponentSet: []string{ical.CompToDo}},
}

// backend serves one request: a calendar-multiget fetches each object
// separately, and all of them share the tags loaded for the first.
type backend struct {
	svc  service.Service
	now  func() time.Time
	tags map[uuid.UUID]model.Tag
}

var _ gocaldav.Backend = (*backend)(nil)

func (b *backend) CurrentUserPrincipal(ctx context.Context) (string, error) {
	return principalPath, nil
}

func (b *backend) CalendarHomeSetPath(ctx context.Context) (string, error) {
	return homePath, nil
}

func (b *backend) CreateCalendar(ctx context.Context, calendar *gocaldav.Calendar) error {
	return errReadOnly
}

func (b *backend) ListCalendars(ctx context.Context) ([]gocaldav.Calendar, error) {
	return calendars, nil
}

func (b *backend) GetCalendar(ctx context.Context, calendarPath string) (*gocaldav.Calendar, error) {
	collection := collectionPath(calendarPath)
	for _, calendar := range calendars {
		if calendar.Path == collection {
			return &calendar, nil
		}
	}
	return nil, errNotFound
}

func (b *backend) GetCalendarObject(ctx context.Context, objectPath string, request *gocaldav.CalendarCompRequest) (*gocaldav.CalendarObject, error) {
	object, err := b.getObject(ctx, objectPath)
	return object, clientError(ctx, err)
}

func (b *backend) getObject(ctx context.Context, objectPath string) (*gocaldav.CalendarObject, error) {
	collection, name := path.Split(objectPath)
	id, err := uuid.Parse(strings.TrimSuffix(name, objectSuffix))
	if err != nil || id.String()+objectSuffix != name || (collection != timePath && collection != tasksPath) {
		return nil, errNotFound
	}
	tags, err := b.loadTags(ctx)
	if err != nil {
		return nil, err
	}

	if collection == timePath {
		timespan, err := b.svc.GetTimespan(ctx, id)
		if err != nil {
			return nil, err
		}
		return timespanObject(timespan, tags)
	}
	task, err := b.svc.GetTask(ctx, id, nil)
	if err != nil {
		return nil, err
	}
	if task.ClosedBefore(b.closedSince()) {
		return nil, errNotFound
	}
	return taskObject(task, tags)
}

func (b *backend) ListCalendarObjects(ctx context.Context, calendarPath string, request *gocaldav.CalendarCompRequest) ([]gocaldav.CalendarObject, error) {
	objects, err := b.listObjects(ctx, collectionPath(calendarPath), nil, nil)
	return objects, clientError(ctx, err)
}

func (b *backend) QueryCalendarObjects(ctx context.Context, calendarPath string, query *gocaldav.CalendarQuery) ([]gocaldav.CalendarObject, error) {
	var from, to *time.Time
	for _, filter := range query.CompFilter.Comps {
		if filter.Name != ical.CompEvent {
			continue
		}
		if !filter.Start.IsZero() {
			from = &filter.Start
		}
		if !filter.End.IsZero() {
			to = &filter.End
		}
	}

	objects, err := b.listObjects(ctx, collectionPath(calendarPath), from, to)
	if err == nil {
		objects, err = gocaldav.Filter(query, objects)
	}
	return objects, clientError(ctx, err)
}

func (b *backend) PutCalendarObject(ctx context.Context, objectPath string, calendar *ical.Calendar, options *gocaldav.PutCalendarObjectOptions) (*gocaldav.CalendarObject, error) {
	return nil, errReadOnly
}

func (b *backend) DeleteCalendarObject(ctx context.Context, objectPath string) error {
	return errReadOnly
}

// listObjects lists a collection's objects; from and to bound timespans only.
func (b *backend) listObjects(ctx context.Context, collection string, from, to *time.Time) ([]gocaldav.CalendarObject, error) {
	if collection != timePath && collection != tasksPath {
		return nil, errNotFound
	}
	tags, err := b.loadTags(ctx)
	if err != nil {
		return nil, err
	}

	if collection == timePath {
		timespans, err := b.svc.ListCalendarTimespans(ctx, from, to)
		if err != nil {
			return nil, err
		}
		return toObjects(timespans, tags, timespanObject)
	}
	tasks, err := b.svc.ListCalendarTasks(ctx, b.closedSince())
	if err != nil {
		return nil, err
	}
	return toObjects(tasks, tags, taskObject)
}

func (b *backend) loadTags(ctx context.Context) (map[uuid.UUID]model.Tag, error) {
	if b.tags == nil {
		tags, err := b.svc.ListCalendarTags(ctx)
		if err != nil {
			return nil, err
		}
		b.tags = tags
	}
	return b.tags, nil
}

func (b *backend) closedSince() time.Time {
	return b.now().Add(-closedTaskWindow)
}

func timespanObject(timespan model.Timespan, tags map[uuid.UUID]model.Tag) (*gocaldav.CalendarObject, error) {
	return newObject(timePath, timespan.Id, TimespanEvent(timespan, tags))
}

func taskObject(task model.Task, tags map[uuid.UUID]model.Tag) (*gocaldav.CalendarObject, error) {
	return newObject(tasksPath, task.Id, TaskTodo(task, tags))
}

func toObjects[T any](items []T, tags map[uuid.UUID]model.Tag, toObject func(T, map[uuid.UUID]model.Tag) (*gocaldav.CalendarObject, error)) ([]gocaldav.CalendarObject, error) {
	objects := make([]gocaldav.CalendarObject, 0, len(items))
	for _, item := range items {
		object, err := toObject(item, tags)
		if err != nil {
			return nil, err
		}
		objects = append(objects, *object)
	}
	return objects, nil
}

func newObject(collection string, id uuid.UUID, calendar *ical.Calendar) (*gocaldav.CalendarObject, error) {
	data, etag, err := Encode(calendar)
	if err != nil {
		return nil, err
	}
	return &gocaldav.CalendarObject{
		Path:          collection + id.String() + objectSuffix,
		ContentLength: int64(len(data)),
		ETag:          etag,
		Data:          calendar,
	}, nil
}

// collectionPath accepts a collection path with or without its trailing slash.
func collectionPath(calendarPath string) string {
	return strings.TrimSuffix(calendarPath, "/") + "/"
}

// clientError keeps internal error text out of responses, since go-webdav
// writes an error's text into the body.
func clientError(ctx context.Context, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, errNotFound), errors.Is(err, model.ErrNotFound):
		return errNotFound
	default:
		slog.ErrorContext(ctx, "caldav request failed", "error", err)
		return errInternal
	}
}
