package memory

import (
	"context"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

// CreateTimespan implements [repository.TimespanRepository].
func (t *MemoryStore) CreateTimespan(ctx context.Context, timespan model.Timespan) (model.Timespan, error) {
	if timespan.StartTime.IsZero() || timespan.EndTime.IsZero() || timespan.EndTime.Before(timespan.StartTime) || timespan.EndTime.Equal(timespan.StartTime) {
		return model.Timespan{}, model.ErrInvalidArgument
	}

	var tagIds []uuid.UUID

	if timespan.TagIds != nil {
		if !t.tagsExist(ctx, timespan.TagIds) {
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

	t.timespans[newId] = newTimespan
	return newTimespan, nil
}

// GetTimespan implements [repository.TimespanRepository].
func (t *MemoryStore) GetTimespan(ctx context.Context, id uuid.UUID) (model.Timespan, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	timespan, exists := t.timespans[id]
	if !exists {
		return model.Timespan{}, model.ErrNotFound
	}

	return timespan, nil
}

// ListTimespans implements [repository.TimespanRepository].
func (t *MemoryStore) ListTimespans(ctx context.Context) ([]model.Timespan, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	timespans := make([]model.Timespan, 0, len(t.timespans))

	for _, timespan := range t.timespans {
		timespans = append(timespans, timespan)
	}

	return timespans, nil
}

// UpdateTimespan implements [repository.TimespanRepository].
func (t *MemoryStore) UpdateTimespan(ctx context.Context, timespan model.Timespan) (model.Timespan, error) {
	if timespan.StartTime.IsZero() || timespan.EndTime.IsZero() || timespan.EndTime.Before(timespan.StartTime) || timespan.EndTime.Equal(timespan.StartTime) {
		return model.Timespan{}, model.ErrInvalidArgument
	}

	if timespan.TagIds != nil && !t.tagsExist(ctx, timespan.TagIds) {
		return model.Timespan{}, model.ErrInvalidReference
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	_, exists := t.timespans[timespan.Id]
	if !exists {
		return model.Timespan{}, model.ErrNotFound
	}

	t.timespans[timespan.Id] = timespan
	return timespan, nil
}

// DeleteTimespan implements [repository.TimespanRepository].
func (t *MemoryStore) DeleteTimespan(ctx context.Context, id uuid.UUID) error {
	// Skip locking for write if the timespan does not exist
	t.mu.RLock()
	_, exists := t.timespans[id]
	t.mu.RUnlock()

	if !exists {
		return model.ErrNotFound
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// delete is a noop if the key does not exist
	// thus, it does not matter if it has been deleted by another thread before this line
	delete(t.timespans, id)
	return nil
}

// GetTotalDurationByTags implements [repository.Repository].
func (t *MemoryStore) GetTotalDurationByTags(ctx context.Context, tagIds []uuid.UUID) (time.Duration, error) {
	if tagIds == nil {
		return 0, model.ErrInvalidArgument
	}

	if len(tagIds) == 0 {
		return 0, nil
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	// Check for invalid tag IDs
	for _, tagId := range tagIds {
		if _, exists := t.tags[tagId]; !exists {
			return 0, model.ErrInvalidReference
		}
	}

	timespanIds := []uuid.UUID{}
	for _, inputTagId := range tagIds {
		for _, timespan := range t.timespans {
			if slices.Contains(timespan.TagIds, inputTagId) && !slices.Contains(timespanIds, timespan.Id) {
				timespanIds = append(timespanIds, timespan.Id)
			}
		}
	}

	totalDuration := time.Duration(0)
	for _, timespanId := range timespanIds {
		ts := t.timespans[timespanId]
		totalDuration += ts.EndTime.Sub(ts.StartTime)
	}

	return totalDuration, nil
}
