package caldav

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"path"
	"slices"
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
	errReadOnly           = webdav.NewHTTPError(http.StatusForbidden, errors.New("caldav: calendars can't be created"))
	errNotFound           = webdav.NewHTTPError(http.StatusNotFound, nil)
	errPreconditionFailed = webdav.NewHTTPError(http.StatusPreconditionFailed, nil)
	errHasSubtasks        = webdav.NewHTTPError(http.StatusForbidden, errors.New("caldav: tasks with subtasks can't be deleted here"))
	errInternal           = webdav.NewHTTPError(http.StatusInternalServerError, errors.New("internal error"))
)

// clientErrors are safe to send to clients as they are.
var clientErrors = []error{errReadOnly, errNotFound, errPreconditionFailed, errHasSubtasks, errUnsupported, errInvalid}

var calendars = []gocaldav.Calendar{
	{Path: timePath, Name: "Time", SupportedComponentSet: []string{ical.CompEvent}},
	{Path: tasksPath, Name: "Tasks", SupportedComponentSet: []string{ical.CompToDo}},
}

// backend serves one request: a calendar-multiget fetches each object
// separately, and all of them share the tags and location loaded for the
// first.
type backend struct {
	svc      service.Service
	now      func() time.Time
	tags     map[uuid.UUID]model.Tag
	location *time.Location
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
	if collection != timePath && collection != tasksPath {
		return nil, errNotFound
	}
	tags, err := b.loadTags(ctx)
	if err != nil {
		return nil, err
	}

	if collection == timePath {
		location, err := b.loadLocation(ctx)
		if err != nil {
			return nil, err
		}
		timespan, err := b.lookupTimespan(ctx, name)
		if err != nil {
			return nil, err
		}
		return timespanObject(timespan, tags, location)
	}
	task, err := b.lookupTask(ctx, name)
	if err != nil {
		return nil, err
	}
	if task.ClosedBefore(b.closedSince()) {
		return nil, errNotFound
	}
	return taskObject(task, tags)
}

// lookupTimespan finds the timespan served under name: by id for
// "<id>.ics", else by the name a client created it under.
func (b *backend) lookupTimespan(ctx context.Context, name string) (model.Timespan, error) {
	if id, ok := objectId(name); ok {
		timespan, err := b.svc.GetTimespan(ctx, id)
		if !errors.Is(err, model.ErrNotFound) {
			return timespan, err
		}
	}
	return b.svc.GetCalendarTimespan(ctx, name)
}

// lookupTask is [backend.lookupTimespan] for tasks.
func (b *backend) lookupTask(ctx context.Context, name string) (model.Task, error) {
	if id, ok := objectId(name); ok {
		task, err := b.svc.GetTask(ctx, id, nil)
		if !errors.Is(err, model.ErrNotFound) {
			return task, err
		}
	}
	return b.svc.GetCalendarTask(ctx, name)
}

// objectId parses the names this server gives objects, "<id>.ics" with a
// lower-case id. A client may still have created an object under such a name.
func objectId(name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(strings.TrimSuffix(name, objectSuffix))
	return id, err == nil && id.String()+objectSuffix == name
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

// PutCalendarObject answers without an ETag: what is stored never matches
// what the client sent byte for byte, so the client fetches it again.
func (b *backend) PutCalendarObject(ctx context.Context, objectPath string, calendar *ical.Calendar, options *gocaldav.PutCalendarObjectOptions) (*gocaldav.CalendarObject, error) {
	var err error
	switch collection, name := path.Split(objectPath); collection {
	case timePath:
		err = b.putTimespan(ctx, name, calendar, options)
	case tasksPath:
		err = b.putTask(ctx, name, calendar, options)
	default:
		err = errNotFound
	}
	if err != nil {
		return nil, clientError(ctx, err)
	}
	return &gocaldav.CalendarObject{Path: objectPath}, nil
}

func (b *backend) DeleteCalendarObject(ctx context.Context, objectPath string) error {
	var err error
	switch collection, name := path.Split(objectPath); collection {
	case timePath:
		err = b.deleteTimespan(ctx, name)
	case tasksPath:
		err = b.deleteTask(ctx, name)
	default:
		err = errNotFound
	}
	return clientError(ctx, err)
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
		location, err := b.loadLocation(ctx)
		if err != nil {
			return nil, err
		}
		timespans, err := b.svc.ListCalendarTimespans(ctx, from, to)
		if err != nil {
			return nil, err
		}
		return toObjects(timespans, func(timespan model.Timespan) (*gocaldav.CalendarObject, error) {
			return timespanObject(timespan, tags, location)
		})
	}
	tasks, err := b.svc.ListCalendarTasks(ctx, b.closedSince())
	if err != nil {
		return nil, err
	}
	return toObjects(tasks, func(task model.Task) (*gocaldav.CalendarObject, error) {
		return taskObject(task, tags)
	})
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

// loadLocation falls back to UTC when the timezone setting is "browser",
// since only the browser knows its zone.
func (b *backend) loadLocation(ctx context.Context) (*time.Location, error) {
	if b.location == nil {
		settings, err := b.svc.GetSettings(ctx)
		if err != nil {
			return nil, err
		}
		location, err := time.LoadLocation(settings.Timezone)
		if settings.Timezone == model.TimezoneBrowser || err != nil {
			location = time.UTC
		}
		b.location = location
	}
	return b.location, nil
}

func (b *backend) closedSince() time.Time {
	return b.now().Add(-closedTaskWindow)
}

func timespanObject(timespan model.Timespan, tags map[uuid.UUID]model.Tag, location *time.Location) (*gocaldav.CalendarObject, error) {
	return newObject(timePath+objectName(timespan.Id, timespan.CaldavName), TimespanEvent(timespan, tags, location))
}

func taskObject(task model.Task, tags map[uuid.UUID]model.Tag) (*gocaldav.CalendarObject, error) {
	return newObject(tasksPath+objectName(task.Id, task.CaldavName), TaskTodo(task, tags))
}

func toObjects[T any](items []T, toObject func(T) (*gocaldav.CalendarObject, error)) ([]gocaldav.CalendarObject, error) {
	objects := make([]gocaldav.CalendarObject, 0, len(items))
	for _, item := range items {
		object, err := toObject(item)
		if err != nil {
			return nil, err
		}
		objects = append(objects, *object)
	}
	return objects, nil
}

// objectName keeps the name a client created an object under, so it finds
// the object again.
func objectName(id uuid.UUID, caldavName *string) string {
	if caldavName != nil {
		return *caldavName
	}
	return id.String() + objectSuffix
}

func newObject(objectPath string, calendar *ical.Calendar) (*gocaldav.CalendarObject, error) {
	data, etag, err := Encode(calendar)
	if err != nil {
		return nil, err
	}
	return &gocaldav.CalendarObject{
		Path:          objectPath,
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
	case slices.ContainsFunc(clientErrors, func(clientErr error) bool { return errors.Is(err, clientErr) }):
		return err
	case errors.Is(err, model.ErrNotFound):
		return errNotFound
	case errors.Is(err, model.ErrAlreadyExists):
		return errPreconditionFailed
	case errors.Is(err, model.ErrInvalidArgument), errors.Is(err, model.ErrInvalidReference):
		return errInvalid
	default:
		slog.ErrorContext(ctx, "caldav request failed", "error", err)
		return errInternal
	}
}
