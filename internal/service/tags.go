package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
)

type TagService struct {
	repository repository.TagRepository
}

func NewTagService(repo repository.TagRepository) *TagService {
	return &TagService{
		repository: repo,
	}
}

func (s *TagService) GetTag(ctx context.Context, id uuid.UUID) (model.Tag, error) {
	return s.repository.GetTag(ctx, id)
}

func (s *TagService) ListTags(ctx context.Context) ([]model.Tag, error) {
	return s.repository.ListTags(ctx)
}

func (s *TagService) CreateTag(ctx context.Context, tag model.Tag) (model.Tag, error) {
	return s.repository.CreateTag(ctx, tag)
}

func (s *TagService) UpdateTag(ctx context.Context, tag model.Tag) (model.Tag, error) {
	return s.repository.UpdateTag(ctx, tag)
}

func (s *TagService) DeleteTag(ctx context.Context, id uuid.UUID) error {
	return s.repository.DeleteTag(ctx, id)
}
