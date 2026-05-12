package service

import (
	"context"

	"github.com/larssonoliver/inundated/internal/model"
)

func (s *ServiceImpl) GetProjectStats(ctx context.Context, input GetProjectStatsInput) (model.ProjectStats, error) {
	return model.ProjectStats{}, model.ErrNotImplemented
}
