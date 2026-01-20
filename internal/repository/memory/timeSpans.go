package memory

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
)

type TimeSpanStore struct {
	mu   sync.RWMutex
	data map[uuid.UUID]model.TimeSpan
}

var _ repository.TimeSpanRepository = (*TimeSpanStore)(nil)

func NewTimeSpanStore() *TimeSpanStore {
	return &TimeSpanStore{
		mu:   sync.RWMutex{},
		data: make(map[uuid.UUID]model.TimeSpan),
	}
}

// CreateTimeSpan implements [repository.TimeSpanRepository].
func (t *TimeSpanStore) CreateTimeSpan(ctx context.Context, timeSpan model.TimeSpan) (model.TimeSpan, error) {
	if timeSpan.Name == "" || timeSpan.StartTime.IsZero() || timeSpan.EndTime.IsZero() || timeSpan.EndTime.Before(timeSpan.StartTime) || timeSpan.EndTime.Equal(timeSpan.StartTime) {
		return model.TimeSpan{}, model.ErrInvalidArgument
	}

	var tagIds []uuid.UUID

	if timeSpan.TagIds != nil {
		tagIds = make([]uuid.UUID, len(timeSpan.TagIds))
		copy(tagIds, timeSpan.TagIds)
	} else {
		tagIds = []uuid.UUID{}
	}

	newId := uuid.New()
	newTimeSpan := model.TimeSpan{
		Id:        newId,
		Name:      timeSpan.Name,
		StartTime: timeSpan.StartTime,
		EndTime:   timeSpan.EndTime,
		TagIds:    tagIds,
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.data[newId] = newTimeSpan
	return newTimeSpan, nil
}

// GetTimeSpan implements [repository.TimeSpanRepository].
func (t *TimeSpanStore) GetTimeSpan(ctx context.Context, id uuid.UUID) (model.TimeSpan, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	timeSpan, exists := t.data[id]
	if !exists {
		return model.TimeSpan{}, model.ErrNotFound
	}

	return timeSpan, nil
}

// ListTimeSpans implements [repository.TimeSpanRepository].
func (t *TimeSpanStore) ListTimeSpans(ctx context.Context) ([]model.TimeSpan, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	timeSpans := make([]model.TimeSpan, 0, len(t.data))

	for _, timeSpan := range t.data {
		timeSpans = append(timeSpans, timeSpan)
	}

	return timeSpans, nil
}

// UpdateTimeSpan implements [repository.TimeSpanRepository].
func (t *TimeSpanStore) UpdateTimeSpan(ctx context.Context, timeSpan model.TimeSpan) (model.TimeSpan, error) {
	if timeSpan.Name == "" || timeSpan.StartTime.IsZero() || timeSpan.EndTime.IsZero() || timeSpan.EndTime.Before(timeSpan.StartTime) || timeSpan.EndTime.Equal(timeSpan.StartTime) {
		return model.TimeSpan{}, model.ErrInvalidArgument
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	_, exists := t.data[timeSpan.Id]
	if !exists {
		return model.TimeSpan{}, model.ErrNotFound
	}

	t.data[timeSpan.Id] = timeSpan
	return timeSpan, nil
}

// DeleteTimeSpan implements [repository.TimeSpanRepository].
func (t *TimeSpanStore) DeleteTimeSpan(ctx context.Context, id uuid.UUID) error {
	// Skip locking for write if the timeSpan does not exist
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
