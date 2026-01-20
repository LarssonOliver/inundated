package handlers

import (
	"context"

	"github.com/larssonoliver/inundated/internal/api"
)

type TimeSpanHandler struct {
	// svc service.TimeSpanService
}

var _ api.TimeSpanHandler = (*TimeSpanHandler)(nil)

func NewTimeSpanHandler() *TimeSpanHandler {
	return &TimeSpanHandler{}
}

// func NewTimeSpanHandler(svc service.TimeSpanService) *TimeSpanHandler {
// 	return &TimeSpanHandler{
// 		svc,
// 	}
// }

// CreateTimeSpan implements [api.TimeSpanHandler].
func (p *TimeSpanHandler) CreateTimeSpan(ctx context.Context, request api.CreateTimeSpanRequestObject) (api.CreateTimeSpanResponseObject, error) {
	panic("unimplemented")
}

// DeleteTimeSpan implements [api.TimeSpanHandler].
func (p *TimeSpanHandler) DeleteTimeSpan(ctx context.Context, request api.DeleteTimeSpanRequestObject) (api.DeleteTimeSpanResponseObject, error) {
	panic("unimplemented")
}

// GetTimeSpan implements [api.TimeSpanHandler].
func (p *TimeSpanHandler) GetTimeSpan(ctx context.Context, request api.GetTimeSpanRequestObject) (api.GetTimeSpanResponseObject, error) {
	panic("unimplemented")
}

// ListTimeSpans implements [api.TimeSpanHandler].
func (p *TimeSpanHandler) ListTimeSpans(ctx context.Context, request api.ListTimeSpansRequestObject) (api.ListTimeSpansResponseObject, error) {
	panic("unimplemented")
}

// UpdateTimeSpan implements [api.TimeSpanHandler].
func (p *TimeSpanHandler) UpdateTimeSpan(ctx context.Context, request api.UpdateTimeSpanRequestObject) (api.UpdateTimeSpanResponseObject, error) {
	panic("unimplemented")
}
