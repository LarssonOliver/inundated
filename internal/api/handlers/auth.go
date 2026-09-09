package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/api"
	"github.com/larssonoliver/inundated/internal/auth"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/service"
)

const loginBindingTTL = 5 * time.Minute

type AuthHandler struct {
	svc           service.AuthService
	secureCookies bool
}

var _ api.AuthHandler = (*AuthHandler)(nil)

func NewAuthHandler(svc service.AuthService, secureCookies bool) *AuthHandler {
	return &AuthHandler{
		svc:           svc,
		secureCookies: secureCookies,
	}
}

// AuthLogin implements [api.AuthHandler].
func (a *AuthHandler) AuthLogin(ctx context.Context, request api.AuthLoginRequestObject) (api.AuthLoginResponseObject, error) {
	redirectUrl := "/"
	if request.Params.Redirect != nil {
		redirectUrl = safeRedirectPath(*request.Params.Redirect)
	}

	authUrl, err := a.svc.BeginLogin(ctx, redirectUrl)

	if err != nil {
		return nil, errors.New("failed to initiate login")
	}

	state, err := stateFromAuthURL(authUrl)
	if err != nil {
		return nil, errors.New("failed to initiate login")
	}

	binding := auth.BaseCookie(model.LoginBindingCookieName, a.secureCookies)
	binding.Value = state
	binding.HttpOnly = true
	binding.MaxAge = int(loginBindingTTL.Seconds())

	return api.AuthLogin302Response{
		Headers: api.AuthLogin302ResponseHeaders{
			Location:  authUrl,
			SetCookie: binding.String(),
		},
	}, nil
}

func safeRedirectPath(raw string) string {
	if raw == "" || raw[0] != '/' || strings.HasPrefix(raw, "//") || strings.HasPrefix(raw, "/\\") {
		return "/"
	}
	if strings.ContainsAny(raw, "\x00\r\n\t") {
		return "/"
	}
	u, err := url.Parse(raw)
	if err != nil || u.IsAbs() || u.Host != "" {
		return "/"
	}
	return u.String()
}

func stateFromAuthURL(authURL string) (string, error) {
	u, err := url.Parse(authURL)
	if err != nil {
		return "", err
	}
	state := u.Query().Get("state")
	if state == "" {
		return "", errors.New("authorization URL carries no state")
	}
	return state, nil
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

	if request.Params.InundatedLogin == nil || *request.Params.InundatedLogin != request.Params.State {
		return api.AuthCallback401Response{}, nil
	}

	session, token, redirectUrl, err := a.svc.HandleCallback(ctx, stateId, request.Params.Code)
	if err != nil {
		slog.WarnContext(ctx, "OIDC callback failed", "error", err)
		return api.AuthCallback401Response{}, nil
	}

	return api.AuthCallback302Response{
		Headers: api.AuthCallback302ResponseHeaders{
			Location:  safeRedirectPath(redirectUrl),
			SetCookie: auth.NewSessionCookie(token, session.ExpiresAt, a.secureCookies).String(),
		},
	}, nil
}

// AuthLogout implements [api.AuthHandler].
func (a *AuthHandler) AuthLogout(ctx context.Context, request api.AuthLogoutRequestObject) (api.AuthLogoutResponseObject, error) {
	session, ok := model.GetSessionFromContext(ctx)
	if !ok {
		return api.AuthLogout401Response{}, nil
	}

	err := a.svc.LogoutSession(ctx, session.Id)
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		return nil, errors.New("failed to logout session")
	}

	return api.AuthLogout204Response{
		Headers: api.AuthLogout204ResponseHeaders{
			SetCookie: auth.ClearSessionCookie(a.secureCookies).String(),
		},
	}, nil
}
