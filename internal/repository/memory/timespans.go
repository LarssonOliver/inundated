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
	if len(tagIds) == 0 {
		return 0, nil
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

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

// AggregateTimeSpentByTagsAndBuckets implements [repository.ProjectStatsRepository].
func (t *MemoryStore) AggregateTimeSpentByTagsAndBuckets(ctx context.Context, tagIds []uuid.UUID, buckets []model.BucketRange) ([]model.BucketValue, error) {
	_ = ctx

	values := make([]model.BucketValue, len(buckets))
	for i, bucket := range buckets {
		if !bucket.End.After(bucket.Start) {
			return nil, model.ErrInvalidArgument
		}

		values[i] = model.BucketValue{
			Bucket: bucket,
			Value:  0,
		}
	}

	if len(tagIds) == 0 {
		return values, nil
	}

	tagSet := make(map[uuid.UUID]struct{}, len(tagIds))
	for _, tagID := range tagIds {
		tagSet[tagID] = struct{}{}
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, timespan := range t.timespans {
		if !timespanHasAnyTag(timespan.TagIds, tagSet) {
			continue
		}

		for i, bucket := range buckets {
			overlapStart := maxTime(timespan.StartTime, bucket.Start)
			overlapEnd := minTime(timespan.EndTime, bucket.End)

			if overlapEnd.After(overlapStart) {
				values[i].Value += overlapEnd.Sub(overlapStart).Seconds()
			}
		}
	}

	return values, nil
}

func timespanHasAnyTag(timespanTagIDs []uuid.UUID, requestedTagSet map[uuid.UUID]struct{}) bool {
	for _, tagID := range timespanTagIDs {
		if _, ok := requestedTagSet[tagID]; ok {
			return true
		}
	}

	return false
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}

	return b
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}

	return b
}
