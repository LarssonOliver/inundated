package handlers

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/api"
	"github.com/larssonoliver/inundated/internal/auth"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/service"
)

type AuthHandler struct {
	svc service.AuthService
	// secureCookies controls the Secure attribute on the session cookie; it is
	// off for a plain-HTTP origin so the browser keeps the cookie.
	secureCookies bool
}

var _ api.AuthHandler = (*AuthHandler)(nil)

func NewAuthHandler(svc service.AuthService, secureCookies bool) *AuthHandler {
	return &AuthHandler{
		svc:           svc,
		secureCookies: secureCookies,
	}
}

// safeRedirectPath confines the post-login redirect to a path on this site.
// Anything that could send the browser to another origin -- an absolute URL, a
// protocol-relative "//host", a "/\host", a non-http scheme, or a value with
// control characters -- collapses to "/". This keeps the login flow from being
// used as an open redirector.
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
			// Re-checked here even though BeginLogin already sanitizes it: this
			// value ends up verbatim in a Location header.
			Location:  safeRedirectPath(redirectUrl),
			SetCookie: auth.NewSessionCookie(session, a.secureCookies).String(),
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
	if err != nil {
		return nil, errors.New("failed to logout session")
	}

	return api.AuthLogout204Response{
		Headers: api.AuthLogout204ResponseHeaders{
			SetCookie: auth.ClearSessionCookie(a.secureCookies).String(),
		},
	}, nil
}
