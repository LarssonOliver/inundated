package memory

import (
	"context"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/utils"
)

// CreateTimespan implements [repository.TimespanRepository].
func (t *MemoryStore) CreateTimespan(ctx context.Context, scope model.OwnerScope, timespan model.Timespan) (model.Timespan, error) {
	if timespan.StartTime.IsZero() || timespan.EndTime.IsZero() || timespan.EndTime.Before(timespan.StartTime) || timespan.EndTime.Equal(timespan.StartTime) {
		return model.Timespan{}, model.ErrInvalidArgument
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	var tagIds []uuid.UUID

	if timespan.TagIds != nil {
		if !t.tagsUsable(scope, timespan.TagIds, nil) {
			return model.Timespan{}, model.ErrInvalidReference
		}

		deduped := utils.DedupeUUIDs(timespan.TagIds)
		tagIds = make([]uuid.UUID, len(deduped))
		copy(tagIds, deduped)
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
		UserId:    scope.UserID(),
	}

	t.timespans = append(t.timespans, newTimespan)
	return newTimespan, nil
}

// GetTimespan implements [repository.TimespanRepository].
func (t *MemoryStore) GetTimespan(ctx context.Context, scope model.OwnerScope, id uuid.UUID) (model.Timespan, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	idx := slices.IndexFunc(t.timespans, func(ts model.Timespan) bool { return ts.Id == id })
	if idx == -1 || !matchesScope(t.timespans[idx].UserId, scope) {
		return model.Timespan{}, model.ErrNotFound
	}

	return t.timespans[idx], nil
}

// ListTimespans implements [repository.TimespanRepository].
func (t *MemoryStore) ListTimespans(ctx context.Context, scope model.OwnerScope, params model.TimespanListParams) (model.Page[model.Timespan], error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	all := make([]model.Timespan, 0, len(t.timespans))
	for _, ts := range t.timespans {
		if !matchesScope(ts.UserId, scope) {
			continue
		}
		if params.From != nil && !ts.EndTime.After(*params.From) {
			continue
		}
		if params.To != nil && !ts.StartTime.Before(*params.To) {
			continue
		}
		all = append(all, ts)
	}
	slices.SortFunc(all, func(a, b model.Timespan) int {
		return -a.StartTime.Compare(b.StartTime) // Negate to sort in descending order
	})

	total := len(all)
	start := min(params.Offset, total)
	end := min(start+params.Limit, total)

	return model.Page[model.Timespan]{
		Data:       all[start:end],
		TotalCount: total,
		Limit:      params.Limit,
		Offset:     params.Offset,
	}, nil
}

// UpdateTimespan implements [repository.TimespanRepository].
func (t *MemoryStore) UpdateTimespan(ctx context.Context, scope model.OwnerScope, timespan model.Timespan) (model.Timespan, error) {
	if timespan.StartTime.IsZero() || timespan.EndTime.IsZero() || timespan.EndTime.Before(timespan.StartTime) || timespan.EndTime.Equal(timespan.StartTime) {
		return model.Timespan{}, model.ErrInvalidArgument
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	idx := slices.IndexFunc(t.timespans, func(ts model.Timespan) bool {
		return ts.Id == timespan.Id && matchesScope(ts.UserId, scope)
	})
	if idx == -1 {
		return model.Timespan{}, model.ErrNotFound
	}
	if timespan.TagIds != nil {
		if !t.tagsUsable(scope, timespan.TagIds, t.timespans[idx].TagIds) {
			return model.Timespan{}, model.ErrInvalidReference
		}
		timespan.TagIds = utils.DedupeUUIDs(timespan.TagIds)
	}

	timespan.UserId = t.timespans[idx].UserId
	t.timespans[idx] = timespan
	return timespan, nil
}

// DeleteTimespan implements [repository.TimespanRepository].
func (t *MemoryStore) DeleteTimespan(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	idx := slices.IndexFunc(t.timespans, func(ts model.Timespan) bool {
		return ts.Id == id && matchesScope(ts.UserId, scope)
	})
	if idx == -1 {
		return model.ErrNotFound
	}

	t.timespans = slices.Delete(t.timespans, idx, idx+1)
	return nil
}

// GetTotalDurationByTags implements [repository.Repository].
func (t *MemoryStore) GetTotalDurationByTags(ctx context.Context, scope model.OwnerScope, tagIds []uuid.UUID) (time.Duration, error) {
	if len(tagIds) == 0 {
		return 0, nil
	}

	tagSet := make(map[uuid.UUID]struct{}, len(tagIds))
	for _, tagID := range tagIds {
		tagSet[tagID] = struct{}{}
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	totalDuration := time.Duration(0)
	for _, timespan := range t.timespans {
		if !matchesScope(timespan.UserId, scope) || !t.timespanHasAnyTag(timespan, tagSet) {
			continue
		}
		totalDuration += timespan.EndTime.Sub(timespan.StartTime)
	}

	return totalDuration, nil
}

// AggregateTimeSpentByTagsAndBuckets implements [repository.ProjectStatsRepository].
func (t *MemoryStore) AggregateTimeSpentByTagsAndBuckets(ctx context.Context, scope model.OwnerScope, tagIds []uuid.UUID, buckets []model.BucketRange) ([]model.BucketValue, error) {
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
		if !matchesScope(timespan.UserId, scope) {
			continue
		}
		if !t.timespanHasAnyTag(timespan, tagSet) {
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

// timespanHasAnyTag reports whether any of the requested tags is among the
// timespan's effective tags: the tags it carries, plus the effective tags of
// every task whose task tag it carries. Callers must hold t.mu.
func (t *MemoryStore) timespanHasAnyTag(timespan model.Timespan, requestedTagSet map[uuid.UUID]struct{}) bool {
	for _, tagID := range timespan.TagIds {
		if _, ok := requestedTagSet[tagID]; ok {
			return true
		}
		idx := slices.IndexFunc(t.tasks, func(task model.Task) bool { return task.TagId == tagID })
		if idx == -1 {
			continue
		}
		for effective := range t.taskEffectiveTags(t.tasks[idx]) {
			if _, ok := requestedTagSet[effective]; ok {
				return true
			}
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
