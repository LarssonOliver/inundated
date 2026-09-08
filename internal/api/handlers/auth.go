package handlers

import (
	"context"
	"errors"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/api"
	"github.com/larssonoliver/inundated/internal/auth"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/service"
)

// loginBindingCookieName is the HttpOnly cookie planted by AuthLogin and echoed
// back by AuthCallback. Binding the callback to a cookie the browser that
// started the flow received means an attacker cannot get a victim's browser to
// complete the attacker's authorization (login CSRF / session fixation): the
// victim's browser carries no matching cookie.
const loginBindingCookieName = "inundated_login"

// loginBindingTTL bounds how long a started login may sit before its callback;
// it matches the server-side login-state expiry.
const loginBindingTTL = 5 * time.Minute

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

	state, err := stateFromAuthURL(authUrl)
	if err != nil {
		return nil, errors.New("failed to initiate login")
	}

	binding := auth.BaseCookie(loginBindingCookieName, a.secureCookies)
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

// stateFromAuthURL pulls the OAuth state parameter out of the provider
// authorization URL so it can be planted as the browser-binding cookie.
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

	// The callback must come from the same browser that started the login: its
	// binding cookie has to echo the state. A missing or mismatched cookie means
	// this is a cross-browser/forged callback (login CSRF / session fixation).
	if request.Params.InundatedLogin == nil || *request.Params.InundatedLogin != request.Params.State {
		return api.AuthCallback401Response{}, nil
	}

	session, redirectUrl, err := a.svc.HandleCallback(ctx, stateId, request.Params.Code)
	if err != nil {
		// The browser only ever sees an opaque 401, so this log line is the
		// one place an operator can find out why a login failed (bad nonce,
		// token exchange failure, an identity with no email claim, ...).
		log.Printf("OIDC callback failed: %v", err)
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

	// A session that is already gone (double-clicked logout, or the cleanup
	// goroutine reaped the row mid-request) is the desired end state, not a
	// failure: fall through to 204 and still clear the cookie.
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
