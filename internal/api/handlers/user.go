package handlers

import (
	"context"
	"github.com/larssonoliver/inundated/internal/api"
	"github.com/larssonoliver/inundated/internal/service"
)

type UserHandler struct {
	svc service.UserService
}

var _ api.UserHandler = (*UserHandler)(nil)

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{
		svc,
	}
}

// GetCurrentUser implements [api.UserHandler].
func (u *UserHandler) GetCurrentUser(ctx context.Context, request api.GetCurrentUserRequestObject) (api.GetCurrentUserResponseObject, error) {
	panic("unimplemented")
}

// UpdateCurrentUser implements [api.UserHandler].
func (u *UserHandler) UpdateCurrentUser(ctx context.Context, request api.UpdateCurrentUserRequestObject) (api.UpdateCurrentUserResponseObject, error) {
	panic("unimplemented")
}
