package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
)

type TimespanServiceImpl struct {
	repository repository.TimespanRepository
}

var _ TimespanService = (*TimespanServiceImpl)(nil)

func NewTimespanService(repo repository.TimespanRepository) *TimespanServiceImpl {
	return &TimespanServiceImpl{
		repository: repo,
	}
}

func (s *TimespanServiceImpl) GetTimespan(ctx context.Context, id uuid.UUID) (model.Timespan, error) {
	return s.repository.GetTimespan(ctx, id)
}

func (s *TimespanServiceImpl) ListTimespans(ctx context.Context) ([]model.Timespan, error) {
	return s.repository.ListTimespans(ctx)
}

func (s *TimespanServiceImpl) CreateTimespan(ctx context.Context, timespan model.Timespan) (model.Timespan, error) {
	return s.repository.CreateTimespan(ctx, timespan)
}

func (s *TimespanServiceImpl) UpdateTimespan(ctx context.Context, timespan model.Timespan) (model.Timespan, error) {
	return s.repository.UpdateTimespan(ctx, timespan)
}

func (s *TimespanServiceImpl) DeleteTimespan(ctx context.Context, id uuid.UUID) error {
	return s.repository.DeleteTimespan(ctx, id)
}
