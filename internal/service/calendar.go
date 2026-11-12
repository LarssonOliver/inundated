package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

const calendarPageSize = 500

func (s *ServiceImpl) ListCalendarTimespans(ctx context.Context, from, to *time.Time) ([]model.Timespan, error) {
	scope, err := ownerScope(ctx)
	if err != nil {
		return nil, err
	}
	return readAllPages(func(page model.PaginationParams) (model.Page[model.Timespan], error) {
		return s.repository.ListTimespans(ctx, scope, model.TimespanListParams{PaginationParams: page, From: from, To: to})
	})
}

func (s *ServiceImpl) ListCalendarTasks(ctx context.Context, closedSince time.Time) ([]model.Task, error) {
	scope, err := ownerScope(ctx)
	if err != nil {
		return nil, err
	}
	tasks, err := readAllPages(func(page model.PaginationParams) (model.Page[model.Task], error) {
		return s.repository.ListTasks(ctx, scope, model.TaskListParams{PaginationParams: page, IncludeClosed: true})
	})
	if err != nil {
		return nil, err
	}
	kept := tasks[:0]
	for _, task := range tasks {
		if !task.ClosedBefore(closedSince) {
			kept = append(kept, task)
		}
	}
	return kept, nil
}

func (s *ServiceImpl) ListCalendarTags(ctx context.Context) (map[uuid.UUID]model.Tag, error) {
	scope, err := ownerScope(ctx)
	if err != nil {
		return nil, err
	}
	tags, err := readAllPages(func(page model.PaginationParams) (model.Page[model.Tag], error) {
		page.IncludeArchived = true
		return s.repository.ListTags(ctx, scope, model.TagListParams{PaginationParams: page, Kinds: []model.TagKind{model.TagKindAll}})
	})
	if err != nil {
		return nil, err
	}
	byId := make(map[uuid.UUID]model.Tag, len(tags))
	for _, tag := range tags {
		byId[tag.Id] = tag
	}
	return byId, nil
}

func readAllPages[T any](list func(page model.PaginationParams) (model.Page[T], error)) ([]T, error) {
	var all []T
	for {
		page, err := list(model.PaginationParams{Limit: calendarPageSize, Offset: len(all)})
		if err != nil {
			return nil, err
		}
		all = append(all, page.Data...)
		if len(page.Data) == 0 || len(all) >= page.TotalCount {
			return all, nil
		}
	}
}
