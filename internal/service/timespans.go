package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

func (s *ServiceImpl) GetTimespan(ctx context.Context, id uuid.UUID) (model.Timespan, error) {
	scope := ownerScope(ctx)
	return s.repository.GetTimespan(ctx, scope, id)
}

func (s *ServiceImpl) ListTimespans(ctx context.Context, params model.PaginationParams) (model.Page[model.Timespan], error) {
	scope := ownerScope(ctx)
	return s.repository.ListTimespans(ctx, scope, params)
}

func (s *ServiceImpl) CreateTimespan(ctx context.Context, timespan model.Timespan) (model.Timespan, error) {
	scope := ownerScope(ctx)
	timespan.Id = uuid.New()
	return s.repository.CreateTimespan(ctx, scope, timespan)
}

func (s *ServiceImpl) UpdateTimespan(ctx context.Context, timespan model.Timespan) (model.Timespan, error) {
	scope := ownerScope(ctx)
	return s.repository.UpdateTimespan(ctx, scope, timespan)
}

func (s *ServiceImpl) DeleteTimespan(ctx context.Context, id uuid.UUID) error {
	scope := ownerScope(ctx)
	return s.repository.DeleteTimespan(ctx, scope, id)
}
