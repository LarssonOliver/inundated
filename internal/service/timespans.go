package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/utils"
)

func (s *ServiceImpl) GetTimespan(ctx context.Context, id uuid.UUID) (model.Timespan, error) {
	scope, err := ownerScope(ctx)
	if err != nil {
		return model.Timespan{}, err
	}
	return s.repository.GetTimespan(ctx, scope, id)
}

func (s *ServiceImpl) ListTimespans(ctx context.Context, params model.PaginationParams, intervalRaw *string) (model.Page[model.Timespan], error) {
	scope, err := ownerScope(ctx)
	if err != nil {
		return model.Page[model.Timespan]{}, err
	}

	listParams := model.TimespanListParams{PaginationParams: params}

	interval := ""
	if intervalRaw != nil {
		interval = strings.TrimSpace(*intervalRaw)
	}
	if interval != "" {
		resolved, err := utils.ParseISO8601Interval(interval, time.Now().UTC(), time.UTC)
		if err != nil {
			if errors.Is(err, utils.ErrISO8601Unprocessable) {
				return model.Page[model.Timespan]{}, model.ErrUnprocessable
			}
			return model.Page[model.Timespan]{}, model.ErrInvalidArgument
		}
		listParams.From = &resolved.Start
		listParams.To = &resolved.End
	}

	return s.repository.ListTimespans(ctx, scope, listParams)
}

func (s *ServiceImpl) CreateTimespan(ctx context.Context, timespan model.Timespan) (model.Timespan, error) {
	scope, err := ownerScope(ctx)
	if err != nil {
		return model.Timespan{}, err
	}
	timespan.Id = uuid.New()
	return s.repository.CreateTimespan(ctx, scope, timespan)
}

func (s *ServiceImpl) UpdateTimespan(ctx context.Context, timespan model.Timespan) (model.Timespan, error) {
	scope, err := ownerScope(ctx)
	if err != nil {
		return model.Timespan{}, err
	}
	return s.repository.UpdateTimespan(ctx, scope, timespan)
}

func (s *ServiceImpl) DeleteTimespan(ctx context.Context, id uuid.UUID) error {
	scope, err := ownerScope(ctx)
	if err != nil {
		return err
	}
	return s.repository.DeleteTimespan(ctx, scope, id)
}
