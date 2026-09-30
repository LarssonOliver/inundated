package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

func (s *ServiceImpl) GetTag(ctx context.Context, id uuid.UUID, includes *TagServiceGetIncludes) (model.Tag, error) {
	scope, err := ownerScope(ctx)
	if err != nil {
		return model.Tag{}, err
	}

	tag, err := s.repository.GetTag(ctx, scope, id)
	if err != nil {
		return model.Tag{}, err
	}
	tags := []model.Tag{tag}
	if err := s.deriveTagColors(ctx, scope, tags); err != nil {
		return model.Tag{}, err
	}
	tag = tags[0]

	if includes != nil {
		if includes.TotalTime {
			totalTime, err := s.repository.GetTotalDurationByTags(ctx, scope, []uuid.UUID{tag.Id})
			if err != nil {
				return model.Tag{}, err
			}
			tag.TotalTime = &totalTime
		}
	}

	return tag, nil
}

func (s *ServiceImpl) ListTags(ctx context.Context, params model.TagListParams) (model.Page[model.Tag], error) {
	scope, err := ownerScope(ctx)
	if err != nil {
		return model.Page[model.Tag]{}, err
	}
	page, err := s.repository.ListTags(ctx, scope, params)
	if err != nil {
		return model.Page[model.Tag]{}, err
	}
	if err := s.deriveTagColors(ctx, scope, page.Data); err != nil {
		return model.Page[model.Tag]{}, err
	}
	return page, nil
}

// deriveTagColors sets the color of each derived tag in tags from its
// source tags (see model.DerivedTagColor), in one lookup for all of them.
func (s *ServiceImpl) deriveTagColors(ctx context.Context, scope model.OwnerScope, tags []model.Tag) error {
	var derivedIds []uuid.UUID
	for _, tag := range tags {
		if tag.IsDerived() {
			derivedIds = append(derivedIds, tag.Id)
		}
	}
	if len(derivedIds) == 0 {
		return nil
	}

	sources, err := s.repository.ListDerivedTagSources(ctx, scope, derivedIds)
	if err != nil {
		return err
	}
	for i, tag := range tags {
		if tag.IsDerived() {
			tags[i].Color = model.DerivedTagColor(sources[tag.Id])
		}
	}
	return nil
}

func (s *ServiceImpl) CreateTag(ctx context.Context, tag model.Tag) (model.Tag, error) {
	scope, err := ownerScope(ctx)
	if err != nil {
		return model.Tag{}, err
	}
	tag.Id = uuid.New()
	return s.repository.CreateTag(ctx, scope, tag)
}

func (s *ServiceImpl) UpdateTag(ctx context.Context, tag model.Tag) (model.Tag, error) {
	scope, err := ownerScope(ctx)
	if err != nil {
		return model.Tag{}, err
	}
	return s.repository.UpdateTag(ctx, scope, tag)
}

func (s *ServiceImpl) DeleteTag(ctx context.Context, id uuid.UUID) error {
	scope, err := ownerScope(ctx)
	if err != nil {
		return err
	}
	return s.repository.DeleteTag(ctx, scope, id)
}
