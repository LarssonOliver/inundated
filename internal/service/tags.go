package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

func (s *ServiceImpl) GetTag(ctx context.Context, id uuid.UUID, includes *TagServiceGetIncludes) (model.Tag, error) {
	scope := ownerScope(ctx)

	tag, err := s.repository.GetTag(ctx, scope, id)

	if err != nil {
		return model.Tag{}, model.ErrNotFound
	}

	if includes != nil {
		if includes.TotalTime {
			totalTime, err := s.repository.GetTotalDurationByTags(ctx, scope, []uuid.UUID{tag.Id})
			if err != nil {
				return model.Tag{}, model.ErrNotFound
			}
			tag.TotalTime = &totalTime
		}
	}

	return tag, nil
}

func (s *ServiceImpl) ListTags(ctx context.Context, params model.PaginationParams) (model.Page[model.Tag], error) {
	scope := ownerScope(ctx)
	return s.repository.ListTags(ctx, scope, params)
}

func (s *ServiceImpl) CreateTag(ctx context.Context, tag model.Tag) (model.Tag, error) {
	scope := ownerScope(ctx)
	tag.Id = uuid.New()
	return s.repository.CreateTag(ctx, scope, tag)
}

func (s *ServiceImpl) UpdateTag(ctx context.Context, tag model.Tag) (model.Tag, error) {
	scope := ownerScope(ctx)
	return s.repository.UpdateTag(ctx, scope, tag)
}

func (s *ServiceImpl) DeleteTag(ctx context.Context, id uuid.UUID) error {
	scope := ownerScope(ctx)
	return s.repository.DeleteTag(ctx, scope, id)
}
