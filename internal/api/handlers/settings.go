package handlers

import (
	"context"
	"errors"

	"github.com/larssonoliver/inundated/internal/api"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/service"
)

type SettingsHandler struct {
	svc service.SettingsService
}

var _ api.SettingsHandler = (*SettingsHandler)(nil)

func NewSettingsHandler(svc service.SettingsService) *SettingsHandler {
	return &SettingsHandler{svc}
}

// GetSettings implements [api.SettingsHandler].
func (h *SettingsHandler) GetSettings(ctx context.Context, request api.GetSettingsRequestObject) (api.GetSettingsResponseObject, error) {
	reply, err := h.svc.GetSettings(ctx)
	if err != nil {
		return nil, err
	}

	return api.GetSettings200JSONResponse(toAPISettings(reply)), nil
}

// UpdateSettings implements [api.SettingsHandler].
func (h *SettingsHandler) UpdateSettings(ctx context.Context, request api.UpdateSettingsRequestObject) (api.UpdateSettingsResponseObject, error) {
	current, err := h.svc.GetSettings(ctx)
	if err != nil {
		return nil, err
	}

	if request.Body.WeekStartDay != nil {
		current.WeekStartDay = string(*request.Body.WeekStartDay)
	}
	if request.Body.Timezone != nil {
		current.Timezone = *request.Body.Timezone
	}
	if request.Body.DurationFormat != nil {
		current.DurationFormat = string(*request.Body.DurationFormat)
	}
	if request.Body.TimeFormat != nil {
		current.TimeFormat = string(*request.Body.TimeFormat)
	}

	reply, err := h.svc.UpdateSettings(ctx, current)
	if errors.Is(err, model.ErrInvalidArgument) {
		return api.UpdateSettings400Response{}, nil
	} else if err != nil {
		return nil, err
	}

	return api.UpdateSettings200JSONResponse(toAPISettings(reply)), nil
}

func toAPISettings(s model.Settings) api.Settings {
	return api.Settings{
		WeekStartDay:   api.WeekStartDay(s.WeekStartDay),
		Timezone:       s.Timezone,
		DurationFormat: api.DurationFormat(s.DurationFormat),
		TimeFormat:     api.TimeFormat(s.TimeFormat),
	}
}
