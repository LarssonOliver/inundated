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
	if err := s.colorTaskTags(ctx, scope, tags); err != nil {
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
	if err := s.colorTaskTags(ctx, scope, page.Data); err != nil {
		return model.Page[model.Tag]{}, err
	}
	return page, nil
}

// colorTaskTags sets the color of each task tag in tags from its task's
// regular tags (see model.TaskTagColor), in one lookup for all of them.
func (s *ServiceImpl) colorTaskTags(ctx context.Context, scope model.OwnerScope, tags []model.Tag) error {
	var taskIds []uuid.UUID
	for _, tag := range tags {
		if tag.TaskId != nil {
			taskIds = append(taskIds, *tag.TaskId)
		}
	}
	if len(taskIds) == 0 {
		return nil
	}

	labels, err := s.repository.ListTaskLabels(ctx, scope, taskIds)
	if err != nil {
		return err
	}
	for i, tag := range tags {
		if tag.TaskId != nil {
			tags[i].Color = model.TaskTagColor(labels[*tag.TaskId])
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
