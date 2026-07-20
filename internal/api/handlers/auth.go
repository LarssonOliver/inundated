package handlers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/api"
	"github.com/larssonoliver/inundated/internal/auth"
	"github.com/larssonoliver/inundated/internal/service"
)

type AuthHandler struct {
	svc service.AuthService
}

var _ api.AuthHandler = (*AuthHandler)(nil)

func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{
		svc,
	}
}

// AuthLogin implements [api.AuthHandler].
func (a *AuthHandler) AuthLogin(ctx context.Context, request api.AuthLoginRequestObject) (api.AuthLoginResponseObject, error) {
	redirectUrl := ""

	if request.Params.Redirect == nil || *request.Params.Redirect == "" {
		redirectUrl = "/"
	} else {
		redirectUrl = *request.Params.Redirect
	}

	authUrl, err := a.svc.BeginLogin(ctx, redirectUrl)

	if err != nil {
		return nil, errors.New("failed to initiate login")
	}

	return api.AuthLogin302Response{
		Headers: api.AuthLogin302ResponseHeaders{
			Location: authUrl,
		},
	}, nil
}

// AuthCallback implements [api.AuthHandler].
func (a *AuthHandler) AuthCallback(ctx context.Context, request api.AuthCallbackRequestObject) (api.AuthCallbackResponseObject, error) {
	if request.Params.Code == "" || request.Params.State == "" {
		return api.AuthCallback400Response{}, nil
	}

	stateId, err := uuid.Parse(request.Params.State)
	if err != nil {
		return api.AuthCallback400Response{}, nil
	}

	session, redirectUrl, err := a.svc.HandleCallback(ctx, stateId, request.Params.Code)
	if err != nil {
		return api.AuthCallback401Response{}, nil
	}

	return api.AuthCallback302Response{
		Headers: api.AuthCallback302ResponseHeaders{
			Location:  redirectUrl,
			SetCookie: auth.NewSessionCookie(session).String(),
		},
	}, nil
}

// AuthLogout implements [api.AuthHandler].
func (a *AuthHandler) AuthLogout(ctx context.Context, request api.AuthLogoutRequestObject) (api.AuthLogoutResponseObject, error) {
	panic("unimplemented")
}
