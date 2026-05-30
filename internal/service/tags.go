package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

func (s *ServiceImpl) GetTag(ctx context.Context, id uuid.UUID, includes *TagServiceGetIncludes) (model.Tag, error) {
	tag, err := s.repository.GetTag(ctx, id)

	if err != nil {
		return model.Tag{}, model.ErrNotFound
	}

	if includes != nil {
		if includes.TotalTime {
			totalTime, err := s.repository.GetTotalDurationByTags(ctx, []uuid.UUID{tag.Id})
			if err != nil {
				return model.Tag{}, model.ErrNotFound
			}
			tag.TotalTime = &totalTime
		}
	}

	return tag, nil
}

func (s *ServiceImpl) ListTags(ctx context.Context, params model.PaginationParams) (model.Page[model.Tag], error) {
	return s.repository.ListTags(ctx, params)
}

func (s *ServiceImpl) CreateTag(ctx context.Context, tag model.Tag) (model.Tag, error) {
	tag.Id = uuid.New()
	return s.repository.CreateTag(ctx, tag)
}

func (s *ServiceImpl) UpdateTag(ctx context.Context, tag model.Tag) (model.Tag, error) {
	return s.repository.UpdateTag(ctx, tag)
}

func (s *ServiceImpl) DeleteTag(ctx context.Context, id uuid.UUID) error {
	return s.repository.DeleteTag(ctx, id)
}
