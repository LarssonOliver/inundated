package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
)

type TagServiceImpl struct {
	repository repository.TagRepository
}

var _ TagService = (*TagServiceImpl)(nil)

func NewTagService(repo repository.TagRepository) *TagServiceImpl {
	return &TagServiceImpl{
		repository: repo,
	}
}

func (s *TagServiceImpl) GetTag(ctx context.Context, id uuid.UUID) (model.Tag, error) {
	return s.repository.GetTag(ctx, id)
}

func (s *TagServiceImpl) ListTags(ctx context.Context) ([]model.Tag, error) {
	return s.repository.ListTags(ctx)
}

func (s *TagServiceImpl) CreateTag(ctx context.Context, tag model.Tag) (model.Tag, error) {
	return s.repository.CreateTag(ctx, tag)
}

func (s *TagServiceImpl) UpdateTag(ctx context.Context, tag model.Tag) (model.Tag, error) {
	return s.repository.UpdateTag(ctx, tag)
}

func (s *TagServiceImpl) DeleteTag(ctx context.Context, id uuid.UUID) error {
	return s.repository.DeleteTag(ctx, id)
}
