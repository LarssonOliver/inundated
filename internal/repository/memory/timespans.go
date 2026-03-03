package memory

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
)

type TimespanStore struct {
	mu   sync.RWMutex
	data map[uuid.UUID]model.Timespan
	tags repository.TagRepository
}

var _ repository.TimespanRepository = (*TimespanStore)(nil)

func NewTimespanStore(tags repository.TagRepository) *TimespanStore {
	return &TimespanStore{
		mu:   sync.RWMutex{},
		data: make(map[uuid.UUID]model.Timespan),
		tags: tags,
	}
}

// CreateTimespan implements [repository.TimespanRepository].
func (t *TimespanStore) CreateTimespan(ctx context.Context, timespan model.Timespan) (model.Timespan, error) {
	if timespan.StartTime.IsZero() || timespan.EndTime.IsZero() || timespan.EndTime.Before(timespan.StartTime) || timespan.EndTime.Equal(timespan.StartTime) {
		return model.Timespan{}, model.ErrInvalidArgument
	}

	var tagIds []uuid.UUID

	if timespan.TagIds != nil {
		if !tagsExist(ctx, t.tags, timespan.TagIds) {
			return model.Timespan{}, model.ErrInvalidReference
		}

		tagIds = make([]uuid.UUID, len(timespan.TagIds))
		copy(tagIds, timespan.TagIds)
	} else {
		tagIds = []uuid.UUID{}
	}

	newId := uuid.New()
	newTimespan := model.Timespan{
		Id:        newId,
		Name:      timespan.Name,
		StartTime: timespan.StartTime,
		EndTime:   timespan.EndTime,
		TagIds:    tagIds,
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.data[newId] = newTimespan
	return newTimespan, nil
}

// GetTimespan implements [repository.TimespanRepository].
func (t *TimespanStore) GetTimespan(ctx context.Context, id uuid.UUID) (model.Timespan, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	timespan, exists := t.data[id]
	if !exists {
		return model.Timespan{}, model.ErrNotFound
	}

	return timespan, nil
}

// ListTimespans implements [repository.TimespanRepository].
func (t *TimespanStore) ListTimespans(ctx context.Context) ([]model.Timespan, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	timespans := make([]model.Timespan, 0, len(t.data))

	for _, timespan := range t.data {
		timespans = append(timespans, timespan)
	}

	return timespans, nil
}

// UpdateTimespan implements [repository.TimespanRepository].
func (t *TimespanStore) UpdateTimespan(ctx context.Context, timespan model.Timespan) (model.Timespan, error) {
	if timespan.StartTime.IsZero() || timespan.EndTime.IsZero() || timespan.EndTime.Before(timespan.StartTime) || timespan.EndTime.Equal(timespan.StartTime) {
		return model.Timespan{}, model.ErrInvalidArgument
	}

	if timespan.TagIds != nil && !tagsExist(ctx, t.tags, timespan.TagIds) {
		return model.Timespan{}, model.ErrInvalidReference
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	_, exists := t.data[timespan.Id]
	if !exists {
		return model.Timespan{}, model.ErrNotFound
	}

	t.data[timespan.Id] = timespan
	return timespan, nil
}

// DeleteTimespan implements [repository.TimespanRepository].
func (t *TimespanStore) DeleteTimespan(ctx context.Context, id uuid.UUID) error {
	// Skip locking for write if the timespan does not exist
	t.mu.RLock()
	_, exists := t.data[id]
	t.mu.RUnlock()

	if !exists {
		return model.ErrNotFound
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// delete is a noop if the key does not exist
	// thus, it does not matter if it has been deleted by another thread before this line
	delete(t.data, id)
	return nil
}
