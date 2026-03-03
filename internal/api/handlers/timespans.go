package handlers

import (
	"context"

	"github.com/larssonoliver/inundated/internal/api"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/service"
)

type TimespanHandler struct {
	svc service.TimespanService
}

var _ api.TimespanHandler = (*TimespanHandler)(nil)

func NewTimespanHandler(svc service.TimespanService) *TimespanHandler {
	return &TimespanHandler{
		svc,
	}
}

// CreateTimespan implements [api.TimespanHandler].
func (p *TimespanHandler) CreateTimespan(ctx context.Context, request api.CreateTimespanRequestObject) (api.CreateTimespanResponseObject, error) {

	timespan := model.Timespan{
		StartTime: request.Body.StartTime,
		EndTime:   request.Body.EndTime,
	}

	if request.Body.Name != nil {
		timespan.Name = *request.Body.Name
	} else {
		timespan.Name = ""
	}

	if request.Body.TagIds != nil {
		timespan.TagIds = *request.Body.TagIds
	}

	reply, err := p.svc.CreateTimespan(ctx, timespan)

	if err == model.ErrInvalidArgument {
		return api.CreateTimespan400Response{}, nil
	} else if err != nil {
		return nil, err
	}

	apiTimespan := api.Timespan{
		Id:        reply.Id,
		Name:      &reply.Name,
		StartTime: reply.StartTime,
		EndTime:   reply.EndTime,
	}

	if len(reply.TagIds) > 0 {
		apiTimespan.TagIds = &reply.TagIds
	}

	return api.CreateTimespan201JSONResponse(apiTimespan), nil
}

// DeleteTimespan implements [api.TimespanHandler].
func (p *TimespanHandler) DeleteTimespan(ctx context.Context, request api.DeleteTimespanRequestObject) (api.DeleteTimespanResponseObject, error) {
	err := p.svc.DeleteTimespan(ctx, request.TimespanId)

	if err == model.ErrNotFound {
		return api.DeleteTimespan404Response{}, nil
	} else if err != nil {
		return nil, err
	}

	return api.DeleteTimespan204Response{}, nil
}

// GetTimespan implements [api.TimespanHandler].
func (p *TimespanHandler) GetTimespan(ctx context.Context, request api.GetTimespanRequestObject) (api.GetTimespanResponseObject, error) {
	reply, err := p.svc.GetTimespan(ctx, request.TimespanId)

	if err == model.ErrNotFound {
		return api.GetTimespan404Response{}, nil
	} else if err != nil {
		return nil, err
	}

	apiTimespan := api.Timespan{
		Id:        reply.Id,
		Name:      &reply.Name,
		StartTime: reply.StartTime,
		EndTime:   reply.EndTime,
	}

	if len(reply.TagIds) > 0 {
		apiTimespan.TagIds = &reply.TagIds
	}

	return api.GetTimespan200JSONResponse(apiTimespan), nil
}

// ListTimespans implements [api.TimespanHandler].
func (p *TimespanHandler) ListTimespans(ctx context.Context, request api.ListTimespansRequestObject) (api.ListTimespansResponseObject, error) {
	reply, err := p.svc.ListTimespans(ctx)

	if err != nil {
		return nil, err
	}

	apiTimespans := make([]api.Timespan, 0, len(reply))
	for _, ts := range reply {
		apiTimespan := api.Timespan{
			Id:        ts.Id,
			Name:      &ts.Name,
			StartTime: ts.StartTime,
			EndTime:   ts.EndTime,
		}
		if len(ts.TagIds) > 0 {
			apiTimespan.TagIds = &ts.TagIds
		}
		apiTimespans = append(apiTimespans, apiTimespan)
	}

	return api.ListTimespans200JSONResponse(apiTimespans), nil
}

// UpdateTimespan implements [api.TimespanHandler].
func (p *TimespanHandler) UpdateTimespan(ctx context.Context, request api.UpdateTimespanRequestObject) (api.UpdateTimespanResponseObject, error) {
	timespan, err := p.svc.GetTimespan(ctx, request.TimespanId)

	if err == model.ErrNotFound {
		return api.UpdateTimespan404Response{}, nil
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

	reply, err := p.svc.UpdateTimespan(ctx, timespan)

	if err == model.ErrInvalidArgument {
		return api.UpdateTimespan400Response{}, nil
	} else if err != nil {
		return nil, err
	}

	apiTimespan := api.Timespan{
		Id:        reply.Id,
		Name:      &reply.Name,
		StartTime: reply.StartTime,
		EndTime:   reply.EndTime,
	}

	if len(reply.TagIds) > 0 {
		apiTimespan.TagIds = &reply.TagIds
	}

	return api.UpdateTimespan200JSONResponse(apiTimespan), nil
}
