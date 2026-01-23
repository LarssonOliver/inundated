package handlers

import (
	"context"

	"github.com/larssonoliver/inundated/internal/api"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/service"
)

type TimeSpanHandler struct {
	svc service.TimeSpanService
}

var _ api.TimeSpanHandler = (*TimeSpanHandler)(nil)

func NewTimeSpanHandler(svc service.TimeSpanService) *TimeSpanHandler {
	return &TimeSpanHandler{
		svc,
	}
}

// CreateTimeSpan implements [api.TimeSpanHandler].
func (p *TimeSpanHandler) CreateTimeSpan(ctx context.Context, request api.CreateTimeSpanRequestObject) (api.CreateTimeSpanResponseObject, error) {
	timespan := model.TimeSpan{
		Name:      request.Body.Name,
		StartTime: request.Body.StartTime,
		EndTime:   request.Body.EndTime,
	}

	if request.Body.TagIds != nil {
		timespan.TagIds = *request.Body.TagIds
	}

	reply, err := p.svc.CreateTimeSpan(ctx, timespan)

	if err == model.ErrInvalidArgument {
		return api.CreateTimeSpan400Response{}, nil
	} else if err != nil {
		return nil, err
	}

	apiTimeSpan := api.TimeSpan{
		Id:        reply.Id,
		Name:      reply.Name,
		StartTime: reply.StartTime,
		EndTime:   reply.EndTime,
	}

	if len(reply.TagIds) > 0 {
		apiTimeSpan.TagIds = &reply.TagIds
	}

	return api.CreateTimeSpan201JSONResponse(apiTimeSpan), nil
}

// DeleteTimeSpan implements [api.TimeSpanHandler].
func (p *TimeSpanHandler) DeleteTimeSpan(ctx context.Context, request api.DeleteTimeSpanRequestObject) (api.DeleteTimeSpanResponseObject, error) {
	err := p.svc.DeleteTimeSpan(ctx, request.TimeSpanId)

	if err == model.ErrNotFound {
		return api.DeleteTimeSpan404Response{}, nil
	} else if err != nil {
		return nil, err
	}

	return api.DeleteTimeSpan204Response{}, nil
}

// GetTimeSpan implements [api.TimeSpanHandler].
func (p *TimeSpanHandler) GetTimeSpan(ctx context.Context, request api.GetTimeSpanRequestObject) (api.GetTimeSpanResponseObject, error) {
	reply, err := p.svc.GetTimeSpan(ctx, request.TimeSpanId)

	if err == model.ErrNotFound {
		return api.GetTimeSpan404Response{}, nil
	} else if err != nil {
		return nil, err
	}

	apiTimeSpan := api.TimeSpan{
		Id:        reply.Id,
		Name:      reply.Name,
		StartTime: reply.StartTime,
		EndTime:   reply.EndTime,
	}

	if len(reply.TagIds) > 0 {
		apiTimeSpan.TagIds = &reply.TagIds
	}

	return api.GetTimeSpan200JSONResponse(apiTimeSpan), nil
}

// ListTimeSpans implements [api.TimeSpanHandler].
func (p *TimeSpanHandler) ListTimeSpans(ctx context.Context, request api.ListTimeSpansRequestObject) (api.ListTimeSpansResponseObject, error) {
	reply, err := p.svc.ListTimeSpans(ctx)

	if err != nil {
		return nil, err
	}

	apiTimeSpans := make([]api.TimeSpan, 0, len(reply))
	for _, ts := range reply {
		apiTimeSpan := api.TimeSpan{
			Id:        ts.Id,
			Name:      ts.Name,
			StartTime: ts.StartTime,
			EndTime:   ts.EndTime,
		}
		if len(ts.TagIds) > 0 {
			apiTimeSpan.TagIds = &ts.TagIds
		}
		apiTimeSpans = append(apiTimeSpans, apiTimeSpan)
	}

	return api.ListTimeSpans200JSONResponse(apiTimeSpans), nil
}

// UpdateTimeSpan implements [api.TimeSpanHandler].
func (p *TimeSpanHandler) UpdateTimeSpan(ctx context.Context, request api.UpdateTimeSpanRequestObject) (api.UpdateTimeSpanResponseObject, error) {
	timespan, err := p.svc.GetTimeSpan(ctx, request.TimeSpanId)

	if err == model.ErrNotFound {
		return api.UpdateTimeSpan404Response{}, nil
	} else if err != nil {
		return nil, err
	}

	if request.Body.Name != nil {
		timespan.Name = *request.Body.Name
	}

	if request.Body.StartTime != nil {
		timespan.StartTime = *request.Body.StartTime
	}

	if request.Body.EndTime != nil {
		timespan.EndTime = *request.Body.EndTime
	}

	if request.Body.TagIds != nil && len(*request.Body.TagIds) >= 0 {
		timespan.TagIds = *request.Body.TagIds
	}

	reply, err := p.svc.UpdateTimeSpan(ctx, timespan)

	if err == model.ErrInvalidArgument {
		return api.UpdateTimeSpan400Response{}, nil
	} else if err != nil {
		return nil, err
	}

	apiTimeSpan := api.TimeSpan{
		Id:        reply.Id,
		Name:      reply.Name,
		StartTime: reply.StartTime,
		EndTime:   reply.EndTime,
	}

	if len(reply.TagIds) > 0 {
		apiTimeSpan.TagIds = &reply.TagIds
	}

	return api.UpdateTimeSpan200JSONResponse(apiTimeSpan), nil
}
