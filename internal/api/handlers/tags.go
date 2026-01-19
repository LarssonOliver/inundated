package handlers

import (
	"context"

	"github.com/larssonoliver/inundated/internal/api"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/service"
)

type TagHandler struct {
	svc service.TagService
}

var _ api.TagHandler = (*TagHandler)(nil)

func NewTagHandler(svc service.TagService) *TagHandler {
	return &TagHandler{
		svc,
	}
}

// CreateTag implements [api.TagHandler].
func (t *TagHandler) CreateTag(ctx context.Context, request api.CreateTagRequestObject) (api.CreateTagResponseObject, error) {
	tag := model.Tag{
		Name:  request.Body.Name,
		Color: request.Body.Color,
	}

	reply, err := t.svc.CreateTag(ctx, tag)

	if err != nil {
		return nil, err
	}

	apiTag := api.Tag{
		Id:    reply.Id,
		Name:  reply.Name,
		Color: reply.Color,
	}

	return api.CreateTag201JSONResponse(apiTag), nil
}

// DeleteTag implements [api.TagHandler].
func (t *TagHandler) DeleteTag(ctx context.Context, request api.DeleteTagRequestObject) (api.DeleteTagResponseObject, error) {
	panic("unimplemented")
}

// GetTag implements [api.TagHandler].
func (t *TagHandler) GetTag(ctx context.Context, request api.GetTagRequestObject) (api.GetTagResponseObject, error) {
	panic("unimplemented")
}

// ListTags implements [api.TagHandler].
func (t *TagHandler) ListTags(ctx context.Context, request api.ListTagsRequestObject) (api.ListTagsResponseObject, error) {
	panic("unimplemented")
}

// UpdateTag implements [api.TagHandler].
func (t *TagHandler) UpdateTag(ctx context.Context, request api.UpdateTagRequestObject) (api.UpdateTagResponseObject, error) {
	panic("unimplemented")
}
