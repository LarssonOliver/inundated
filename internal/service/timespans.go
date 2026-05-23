package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

func (s *ServiceImpl) GetTimespan(ctx context.Context, id uuid.UUID) (model.Timespan, error) {
	return s.repository.GetTimespan(ctx, id)
}

func (s *ServiceImpl) ListTimespans(ctx context.Context, params model.PaginationParams) (model.Page[model.Timespan], error) {
	return s.repository.ListTimespans(ctx, params)
}

func (s *ServiceImpl) CreateTimespan(ctx context.Context, timespan model.Timespan) (model.Timespan, error) {
	timespan.Id = uuid.New()
	return s.repository.CreateTimespan(ctx, timespan)
}

func (s *ServiceImpl) UpdateTimespan(ctx context.Context, timespan model.Timespan) (model.Timespan, error) {
	return s.repository.UpdateTimespan(ctx, timespan)
}

func (s *ServiceImpl) DeleteTimespan(ctx context.Context, id uuid.UUID) error {
	return s.repository.DeleteTimespan(ctx, id)
}
