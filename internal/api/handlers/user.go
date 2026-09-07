package handlers

import (
	"context"
	"errors"

	"github.com/larssonoliver/inundated/internal/api"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/service"
	"github.com/oapi-codegen/runtime/types"
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
	user, err := u.svc.GetCurrentUser(ctx)

	if errors.Is(err, model.ErrNotFound) {
		return api.GetCurrentUser401Response{}, nil
	} else if err != nil {
		return nil, errors.New("internal server error")
	}

	apiUser := api.User{
		Id:    user.Id,
		Sub:   user.Sub,
		Name:  &user.Name,
		Email: types.Email(user.Email),
	}

	return api.GetCurrentUser200JSONResponse(apiUser), nil
}
