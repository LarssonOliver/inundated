package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

func (s *ServiceImpl) GetTag(ctx context.Context, id uuid.UUID) (model.Tag, error) {
	return s.repository.GetTag(ctx, id)
}

func (s *ServiceImpl) ListTags(ctx context.Context) ([]model.Tag, error) {
	return s.repository.ListTags(ctx)
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
