package handlers_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/larssonoliver/inundated/internal/api"
	"github.com/larssonoliver/inundated/internal/api/handlers"
	"github.com/larssonoliver/inundated/internal/auth"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/service"
)

// --- AuthLogin -------------------------------------------------------------

func TestAuthHandler_AuthLogin(t *testing.T) {
	t.Run("no redirect param defaults to root and returns 302 with auth URL", func(t *testing.T) {
		const wantAuthURL = "https://provider.example/oauth/authorize?state=abc"

		var gotRedirect string
		mock := &service.AuthServiceMock{
			BeginLoginFn: func(ctx context.Context, redirectURI string) (string, error) {
				gotRedirect = redirectURI
				return wantAuthURL, nil
			},
		}
		h := handlers.NewAuthHandler(mock, true)

		resp, err := h.AuthLogin(context.Background(), api.AuthLoginRequestObject{
			Params: api.AuthLoginParams{Redirect: nil},
		})

		require.NoError(t, err)
		assert.Equal(t, "/", gotRedirect)
		got, ok := resp.(api.AuthLogin302Response)
		require.True(t, ok, "expected AuthLogin302Response, got %T", resp)
		assert.Equal(t, wantAuthURL, got.Headers.Location)
	})

	t.Run("empty string redirect param defaults to root", func(t *testing.T) {
		empty := ""
		var gotRedirect string
		mock := &service.AuthServiceMock{
			BeginLoginFn: func(ctx context.Context, redirectURI string) (string, error) {
				gotRedirect = redirectURI
				return "https://provider.example/authorize?state=xyz", nil
			},
		}
		h := handlers.NewAuthHandler(mock, true)

		_, err := h.AuthLogin(context.Background(), api.AuthLoginRequestObject{
			Params: api.AuthLoginParams{Redirect: &empty},
		})

		require.NoError(t, err)
		assert.Equal(t, "/", gotRedirect)
	})

	t.Run("explicit redirect param is passed through unchanged", func(t *testing.T) {
		redirect := "/dashboard"
		var gotRedirect string
		mock := &service.AuthServiceMock{
			BeginLoginFn: func(ctx context.Context, redirectURI string) (string, error) {
				gotRedirect = redirectURI
				return "https://provider.example/authorize?state=xyz", nil
			},
		}
		h := handlers.NewAuthHandler(mock, true)

		_, err := h.AuthLogin(context.Background(), api.AuthLoginRequestObject{
			Params: api.AuthLoginParams{Redirect: &redirect},
		})

		require.NoError(t, err)
		assert.Equal(t, redirect, gotRedirect)
	})

	t.Run("redirect param is confined to a site-relative path", func(t *testing.T) {
		cases := map[string]string{
			"/dashboard":                     "/dashboard",
			"/projects/1?tab=notes":          "/projects/1?tab=notes",
			"":                               "/",
			"https://evil.example.com/phish": "/",
			"//evil.example.com":             "/",
			"/\\evil.example.com":            "/",
			"javascript:alert(1)":            "/",
			"relative/no/slash":              "/",
			"/ok\nSet-Cookie: x=y":           "/",
		}

		for in, want := range cases {
			t.Run(in, func(t *testing.T) {
				var got string
				mock := &service.AuthServiceMock{
					BeginLoginFn: func(ctx context.Context, redirectURI string) (string, error) {
						got = redirectURI
						return "https://provider.example/authorize?state=xyz", nil
					},
				}
				h := handlers.NewAuthHandler(mock, true)

				redirect := in
				_, err := h.AuthLogin(context.Background(), api.AuthLoginRequestObject{
					Params: api.AuthLoginParams{Redirect: &redirect},
				})

				require.NoError(t, err)
				assert.Equal(t, want, got)
			})
		}
	})

	t.Run("service error results in generic error and nil response", func(t *testing.T) {
		mock := &service.AuthServiceMock{
			BeginLoginFn: func(ctx context.Context, redirectURI string) (string, error) {
				return "", errors.New("provider unreachable")
			},
		}
		h := handlers.NewAuthHandler(mock, true)

		resp, err := h.AuthLogin(context.Background(), api.AuthLoginRequestObject{
			Params: api.AuthLoginParams{Redirect: nil},
		})

		require.Error(t, err)
		assert.EqualError(t, err, "failed to initiate login")
		assert.Nil(t, resp)
	})
}

// --- AuthCallback ------------------------------------------------------------

func TestAuthHandler_AuthCallback(t *testing.T) {
	t.Run("missing code returns 400", func(t *testing.T) {
		mock := &service.AuthServiceMock{}
		h := handlers.NewAuthHandler(mock, true)

		resp, err := h.AuthCallback(context.Background(), api.AuthCallbackRequestObject{
			Params: api.AuthCallbackParams{Code: "", State: uuid.NewString()},
		})

		require.NoError(t, err)
		assert.IsType(t, api.AuthCallback400Response{}, resp)
	})

	t.Run("missing state returns 400", func(t *testing.T) {
		mock := &service.AuthServiceMock{}
		h := handlers.NewAuthHandler(mock, true)

		resp, err := h.AuthCallback(context.Background(), api.AuthCallbackRequestObject{
			Params: api.AuthCallbackParams{Code: "authcode", State: ""},
		})

		require.NoError(t, err)
		assert.IsType(t, api.AuthCallback400Response{}, resp)
	})

	t.Run("state that is not a valid UUID returns 400", func(t *testing.T) {
		mock := &service.AuthServiceMock{
			// Should never be called since parsing fails first.
			HandleCallbackFn: func(ctx context.Context, stateID uuid.UUID, code string) (model.Session, string, string, error) {
				t.Fatal("HandleCallback should not be called for an invalid state")
				return model.Session{}, "", "", nil
			},
		}
		h := handlers.NewAuthHandler(mock, true)

		resp, err := h.AuthCallback(context.Background(), api.AuthCallbackRequestObject{
			Params: api.AuthCallbackParams{Code: "authcode", State: "not-a-uuid"},
		})

		require.NoError(t, err)
		assert.IsType(t, api.AuthCallback400Response{}, resp)
	})

	t.Run("service error returns 401", func(t *testing.T) {
		state := uuid.New()
		binding := state.String()
		mock := &service.AuthServiceMock{
			HandleCallbackFn: func(ctx context.Context, stateID uuid.UUID, code string) (model.Session, string, string, error) {
				assert.Equal(t, state, stateID)
				assert.Equal(t, "authcode", code)
				return model.Session{}, "", "", errors.New("invalid code")
			},
		}
		h := handlers.NewAuthHandler(mock, true)

		resp, err := h.AuthCallback(context.Background(), api.AuthCallbackRequestObject{
			Params: api.AuthCallbackParams{Code: "authcode", State: state.String(), InundatedLogin: &binding},
		})

		require.NoError(t, err)
		assert.IsType(t, api.AuthCallback401Response{}, resp)
	})

	t.Run("success returns 302 with redirect location and session cookie", func(t *testing.T) {
		state := uuid.New()
		binding := state.String()
		sessionID := uuid.New()
		expiresAt := time.Now().Add(24 * time.Hour).UTC()
		const wantRedirect = "/dashboard"
		const wantToken = "the-opaque-session-token"

		session := model.Session{
			Id:        sessionID,
			ExpiresAt: expiresAt,
		}

		mock := &service.AuthServiceMock{
			HandleCallbackFn: func(ctx context.Context, stateID uuid.UUID, code string) (model.Session, string, string, error) {
				assert.Equal(t, state, stateID)
				assert.Equal(t, "authcode", code)
				return session, wantToken, wantRedirect, nil
			},
		}
		h := handlers.NewAuthHandler(mock, true)

		resp, err := h.AuthCallback(context.Background(), api.AuthCallbackRequestObject{
			Params: api.AuthCallbackParams{Code: "authcode", State: state.String(), InundatedLogin: &binding},
		})

		require.NoError(t, err)
		got, ok := resp.(api.AuthCallback302Response)
		require.True(t, ok, "expected AuthCallback302Response, got %T", resp)

		assert.Equal(t, wantRedirect, got.Headers.Location)

		wantCookie := http.Cookie{
			Name:     model.SessionCookieName,
			Value:    wantToken,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			Expires:  expiresAt,
		}
		assert.Equal(t, wantCookie.String(), got.Headers.SetCookie)
		assert.NotContains(t, got.Headers.SetCookie, sessionID.String(), "the cookie must carry the token, not the session id")
	})

	t.Run("plain-HTTP origin issues the session cookie without the Secure attribute", func(t *testing.T) {
		state := uuid.New()
		binding := state.String()
		session := model.Session{Id: uuid.New(), ExpiresAt: time.Now().Add(24 * time.Hour).UTC()}
		mock := &service.AuthServiceMock{
			HandleCallbackFn: func(ctx context.Context, stateID uuid.UUID, code string) (model.Session, string, string, error) {
				return session, "tok", "/", nil
			},
		}
		h := handlers.NewAuthHandler(mock, false)

		resp, err := h.AuthCallback(context.Background(), api.AuthCallbackRequestObject{
			Params: api.AuthCallbackParams{Code: "authcode", State: state.String(), InundatedLogin: &binding},
		})

		require.NoError(t, err)
		got := resp.(api.AuthCallback302Response)
		assert.NotContains(t, got.Headers.SetCookie, "Secure")
	})
}

// --- login-binding cookie (login CSRF / session fixation defense) ----------

// parseSetCookie extracts the single cookie encoded in a Set-Cookie header value.
func parseSetCookie(t *testing.T, value string) *http.Cookie {
	t.Helper()
	header := http.Header{}
	header.Add("Set-Cookie", value)
	cookies := (&http.Response{Header: header}).Cookies()
	require.Len(t, cookies, 1, "expected exactly one Set-Cookie")
	return cookies[0]
}

func TestAuthHandler_LoginBinding(t *testing.T) {
	authURL := "https://provider.example/authorize?client_id=x&state=abc123&code_challenge=y"

	t.Run("AuthLogin plants an HttpOnly binding cookie carrying the state", func(t *testing.T) {
		mock := &service.AuthServiceMock{
			BeginLoginFn: func(ctx context.Context, redirectURI string) (string, error) {
				return authURL, nil
			},
		}
		h := handlers.NewAuthHandler(mock, true)

		resp, err := h.AuthLogin(context.Background(), api.AuthLoginRequestObject{})
		require.NoError(t, err)
		got := resp.(api.AuthLogin302Response)

		cookie := parseSetCookie(t, got.Headers.SetCookie)
		assert.Equal(t, "inundated_login", cookie.Name)
		assert.Equal(t, "abc123", cookie.Value)
		assert.True(t, cookie.HttpOnly)
		assert.True(t, cookie.Secure)
		assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
	})

	t.Run("plain-HTTP origin plants the binding cookie without Secure", func(t *testing.T) {
		mock := &service.AuthServiceMock{
			BeginLoginFn: func(ctx context.Context, redirectURI string) (string, error) {
				return authURL, nil
			},
		}
		h := handlers.NewAuthHandler(mock, false)

		resp, err := h.AuthLogin(context.Background(), api.AuthLoginRequestObject{})
		require.NoError(t, err)
		got := resp.(api.AuthLogin302Response)
		assert.NotContains(t, got.Headers.SetCookie, "Secure")
	})

	t.Run("AuthCallback with no binding cookie is rejected with 401", func(t *testing.T) {
		state := uuid.New()
		mock := &service.AuthServiceMock{
			HandleCallbackFn: func(ctx context.Context, stateID uuid.UUID, code string) (model.Session, string, string, error) {
				t.Fatal("HandleCallback must not run without a matching binding cookie")
				return model.Session{}, "", "", nil
			},
		}
		h := handlers.NewAuthHandler(mock, true)

		resp, err := h.AuthCallback(context.Background(), api.AuthCallbackRequestObject{
			Params: api.AuthCallbackParams{Code: "authcode", State: state.String()},
		})
		require.NoError(t, err)
		assert.IsType(t, api.AuthCallback401Response{}, resp)
	})

	t.Run("AuthCallback with a binding cookie that does not match state is rejected with 401", func(t *testing.T) {
		state := uuid.New()
		other := uuid.NewString()
		mock := &service.AuthServiceMock{
			HandleCallbackFn: func(ctx context.Context, stateID uuid.UUID, code string) (model.Session, string, string, error) {
				t.Fatal("HandleCallback must not run for a mismatched binding cookie")
				return model.Session{}, "", "", nil
			},
		}
		h := handlers.NewAuthHandler(mock, true)

		resp, err := h.AuthCallback(context.Background(), api.AuthCallbackRequestObject{
			Params: api.AuthCallbackParams{Code: "authcode", State: state.String(), InundatedLogin: &other},
		})
		require.NoError(t, err)
		assert.IsType(t, api.AuthCallback401Response{}, resp)
	})

	t.Run("AuthCallback proceeds when the binding cookie matches state", func(t *testing.T) {
		state := uuid.New()
		binding := state.String()
		session := model.Session{Id: uuid.New(), ExpiresAt: time.Now().Add(time.Hour)}
		mock := &service.AuthServiceMock{
			HandleCallbackFn: func(ctx context.Context, stateID uuid.UUID, code string) (model.Session, string, string, error) {
				return session, "tok", "/", nil
			},
		}
		h := handlers.NewAuthHandler(mock, true)

		resp, err := h.AuthCallback(context.Background(), api.AuthCallbackRequestObject{
			Params: api.AuthCallbackParams{Code: "authcode", State: binding, InundatedLogin: &binding},
		})
		require.NoError(t, err)
		assert.IsType(t, api.AuthCallback302Response{}, resp)
	})
}

func TestAuthHandler_AuthLogout(t *testing.T) {
	t.Run("missing session returns 401", func(t *testing.T) {
		mock := &service.AuthServiceMock{
			LogoutSessionFn: func(ctx context.Context, sessionID uuid.UUID) error {
				t.Fatal("LogoutSession should not be called without a session")
				return nil
			},
		}
		h := handlers.NewAuthHandler(mock, true)

		resp, err := h.AuthLogout(
			context.Background(),
			api.AuthLogoutRequestObject{},
		)

		require.NoError(t, err)
		assert.IsType(t, api.AuthLogout401Response{}, resp)
	})

	t.Run("service error returns error", func(t *testing.T) {
		sessionID := uuid.New()
		session := model.Session{
			Id: sessionID,
		}

		mock := &service.AuthServiceMock{
			LogoutSessionFn: func(ctx context.Context, gotSessionID uuid.UUID) error {
				assert.Equal(t, sessionID, gotSessionID)
				return errors.New("database error")
			},
		}
		h := handlers.NewAuthHandler(mock, true)

		ctx := model.SetSessionInContext(context.Background(), session)

		resp, err := h.AuthLogout(
			ctx,
			api.AuthLogoutRequestObject{},
		)

		require.Error(t, err)
		assert.EqualError(t, err, "failed to logout session")
		assert.Nil(t, resp)
	})

	t.Run("success returns 204 and clears the session cookie", func(t *testing.T) {
		sessionID := uuid.New()
		session := model.Session{
			Id: sessionID,
		}

		mock := &service.AuthServiceMock{
			LogoutSessionFn: func(ctx context.Context, gotSessionID uuid.UUID) error {
				assert.Equal(t, sessionID, gotSessionID)
				return nil
			},
		}
		h := handlers.NewAuthHandler(mock, true)

		ctx := model.SetSessionInContext(context.Background(), session)

		resp, err := h.AuthLogout(
			ctx,
			api.AuthLogoutRequestObject{},
		)

		require.NoError(t, err)
		got, ok := resp.(api.AuthLogout204Response)
		require.True(t, ok, "expected AuthLogout204Response, got %T", resp)
		assert.Equal(t, auth.ClearSessionCookie(true).String(), got.Headers.SetCookie)
	})

	t.Run("an already-deleted session still returns 204 and clears the cookie", func(t *testing.T) {
		// Double-click logout, or the cleanup goroutine reaping the row mid
		// request: the session is gone by the time LogoutSession runs. That is
		// the desired end state, not a server error -- the client must still get
		// 204 and the cookie-clearing Set-Cookie.
		session := model.Session{Id: uuid.New()}
		mock := &service.AuthServiceMock{
			LogoutSessionFn: func(ctx context.Context, gotSessionID uuid.UUID) error {
				return fmt.Errorf("LogoutSession: %w", model.ErrNotFound)
			},
		}
		h := handlers.NewAuthHandler(mock, true)

		resp, err := h.AuthLogout(model.SetSessionInContext(context.Background(), session), api.AuthLogoutRequestObject{})

		require.NoError(t, err)
		got, ok := resp.(api.AuthLogout204Response)
		require.True(t, ok, "expected AuthLogout204Response, got %T", resp)
		assert.Equal(t, auth.ClearSessionCookie(true).String(), got.Headers.SetCookie)
	})
}
