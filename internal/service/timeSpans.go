package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
)

type TimeSpanServiceImpl struct {
	repository repository.TimeSpanRepository
}

var _ TimeSpanService = (*TimeSpanServiceImpl)(nil)

func NewTimeSpanService(repo repository.TimeSpanRepository) *TimeSpanServiceImpl {
	return &TimeSpanServiceImpl{
		repository: repo,
	}
}

func (s *TimeSpanServiceImpl) GetTimeSpan(ctx context.Context, id uuid.UUID) (model.TimeSpan, error) {
	return s.repository.GetTimeSpan(ctx, id)
}

func (s *TimeSpanServiceImpl) ListTimeSpans(ctx context.Context) ([]model.TimeSpan, error) {
	return s.repository.ListTimeSpans(ctx)
}

func (s *TimeSpanServiceImpl) CreateTimeSpan(ctx context.Context, timeSpan model.TimeSpan) (model.TimeSpan, error) {
	return s.repository.CreateTimeSpan(ctx, timeSpan)
}

func (s *TimeSpanServiceImpl) UpdateTimeSpan(ctx context.Context, timeSpan model.TimeSpan) (model.TimeSpan, error) {
	return s.repository.UpdateTimeSpan(ctx, timeSpan)
}

func (s *TimeSpanServiceImpl) DeleteTimeSpan(ctx context.Context, id uuid.UUID) error {
	return s.repository.DeleteTimeSpan(ctx, id)
}
