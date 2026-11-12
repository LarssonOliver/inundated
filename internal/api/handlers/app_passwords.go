package handlers

import (
	"context"
	"errors"

	"github.com/larssonoliver/inundated/internal/api"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/service"
)

type AppPasswordHandler struct {
	svc service.AppPasswordService
}

var _ api.AppPasswordHandler = (*AppPasswordHandler)(nil)

func NewAppPasswordHandler(svc service.AppPasswordService) *AppPasswordHandler {
	return &AppPasswordHandler{svc}
}

// ListAppPasswords implements [api.AppPasswordHandler].
func (h *AppPasswordHandler) ListAppPasswords(ctx context.Context, request api.ListAppPasswordsRequestObject) (api.ListAppPasswordsResponseObject, error) {
	reply, err := h.svc.ListAppPasswords(ctx)
	if err != nil {
		return nil, err
	}

	items := make(api.ListAppPasswords200JSONResponse, len(reply))
	for i, appPassword := range reply {
		items[i] = toAPIAppPassword(appPassword)
	}

	return items, nil
}

// CreateAppPassword implements [api.AppPasswordHandler].
func (h *AppPasswordHandler) CreateAppPassword(ctx context.Context, request api.CreateAppPasswordRequestObject) (api.CreateAppPasswordResponseObject, error) {
	reply, token, err := h.svc.CreateAppPassword(ctx, request.Body.Name)
	if errors.Is(err, model.ErrInvalidArgument) {
		return api.CreateAppPassword400Response{}, nil
	} else if err != nil {
		return nil, err
	}

	appPassword := toAPIAppPassword(reply)

	return api.CreateAppPassword201JSONResponse{
		Id:         appPassword.Id,
		Name:       appPassword.Name,
		CreatedAt:  appPassword.CreatedAt,
		LastUsedAt: appPassword.LastUsedAt,
		Password:   token,
	}, nil
}

// DeleteAppPassword implements [api.AppPasswordHandler].
func (h *AppPasswordHandler) DeleteAppPassword(ctx context.Context, request api.DeleteAppPasswordRequestObject) (api.DeleteAppPasswordResponseObject, error) {
	err := h.svc.DeleteAppPassword(ctx, request.AppPasswordId)
	if errors.Is(err, model.ErrNotFound) {
		return api.DeleteAppPassword404Response{}, nil
	} else if err != nil {
		return nil, err
	}

	return api.DeleteAppPassword204Response{}, nil
}

func toAPIAppPassword(a model.AppPassword) api.AppPassword {
	return api.AppPassword{
		Id:         a.Id,
		Name:       a.Name,
		CreatedAt:  a.CreatedAt,
		LastUsedAt: a.LastUsedAt,
	}
}
