package handlers

import (
	"context"
	"slices"

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

	if err == model.ErrInvalidArgument {
		return api.CreateTag400Response{}, nil
	} else if err != nil {
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
	err := t.svc.DeleteTag(ctx, request.TagId)

	if err == model.ErrNotFound {
		return api.DeleteTag404Response{}, nil
	} else if err != nil {
		return nil, err
	}

	return api.DeleteTag204Response{}, nil
}

// GetTag implements [api.TagHandler].
func (t *TagHandler) GetTag(ctx context.Context, request api.GetTagRequestObject) (api.GetTagResponseObject, error) {
	includes := service.TagServiceGetIncludes{}

	if request.Params.Include != nil {
		includes.TotalTime = slices.Contains(*request.Params.Include, string(api.GetTagParamsIncludeTotalTimeMs))
	}

	reply, err := t.svc.GetTag(ctx, request.TagId, &includes)

	if err == model.ErrNotFound {
		return api.GetTag404Response{}, nil
	} else if err != nil {
		return nil, err
	}

	apiTag := api.Tag{
		Id:    reply.Id,
		Name:  reply.Name,
		Color: reply.Color,
	}

	if includes.TotalTime {
		TotalTimeMs := int(reply.TotalTime.Milliseconds())
		apiTag.TotalTimeMs = &TotalTimeMs
	}

	return api.GetTag200JSONResponse(apiTag), nil
}

// ListTags implements [api.TagHandler].
func (t *TagHandler) ListTags(ctx context.Context, request api.ListTagsRequestObject) (api.ListTagsResponseObject, error) {
	reply, err := t.svc.ListTags(ctx)

	if err != nil {
		return nil, err
	}

	apiTags := make([]api.Tag, 0, len(reply))
	for _, tag := range reply {
		apiTag := api.Tag{
			Id:    tag.Id,
			Name:  tag.Name,
			Color: tag.Color,
		}
		apiTags = append(apiTags, apiTag)
	}

	response := api.PaginatedTags{
		Data: apiTags,
	}

	return api.ListTags200JSONResponse(response), nil
}

// UpdateTag implements [api.TagHandler].
func (t *TagHandler) UpdateTag(ctx context.Context, request api.UpdateTagRequestObject) (api.UpdateTagResponseObject, error) {
	tag, err := t.svc.GetTag(ctx, request.TagId, nil)
	if err == model.ErrNotFound {
		return api.UpdateTag404Response{}, nil
	} else if err != nil {
		return nil, err
	}

	if request.Body.Name != nil {
		tag.Name = *request.Body.Name
	}
	if request.Body.Color != nil {
		tag.Color = *request.Body.Color
	}

	reply, err := t.svc.UpdateTag(ctx, tag)

	if err == model.ErrInvalidArgument {
		return api.UpdateTag400Response{}, nil
	} else if err != nil {
		return nil, err
	}

	apiTag := api.Tag{
		Id:    reply.Id,
		Name:  reply.Name,
		Color: reply.Color,
	}

	return api.UpdateTag200JSONResponse(apiTag), nil
}
