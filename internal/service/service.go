package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
)

type ServiceImpl struct {
	TagServiceImpl
}

var _ Service = (*ServiceImpl)(nil)

func NewService(repository repository.Repository) *ServiceImpl {
	return &ServiceImpl{
		*NewTagService(repository),
	}
}

type Service interface {
	TagService
}

type TagService interface {
	GetTag(ctx context.Context, id uuid.UUID) (model.Tag, error)
	ListTags(ctx context.Context) ([]model.Tag, error)
	CreateTag(ctx context.Context, tag model.Tag) (model.Tag, error)
	UpdateTag(ctx context.Context, tag model.Tag) (model.Tag, error)
	DeleteTag(ctx context.Context, id uuid.UUID) error
}
