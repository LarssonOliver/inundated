package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/larssonoliver/inundated/internal/api"
	"github.com/larssonoliver/inundated/internal/api/handlers"
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
		h := handlers.NewAuthHandler(mock)

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
				return "https://provider.example/authorize", nil
			},
		}
		h := handlers.NewAuthHandler(mock)

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
				return "https://provider.example/authorize", nil
			},
		}
		h := handlers.NewAuthHandler(mock)

		_, err := h.AuthLogin(context.Background(), api.AuthLoginRequestObject{
			Params: api.AuthLoginParams{Redirect: &redirect},
		})

		require.NoError(t, err)
		assert.Equal(t, redirect, gotRedirect)
	})

	t.Run("service error results in generic error and nil response", func(t *testing.T) {
		mock := &service.AuthServiceMock{
			BeginLoginFn: func(ctx context.Context, redirectURI string) (string, error) {
				return "", errors.New("provider unreachable")
			},
		}
		h := handlers.NewAuthHandler(mock)

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
		h := handlers.NewAuthHandler(mock)

		resp, err := h.AuthCallback(context.Background(), api.AuthCallbackRequestObject{
			Params: api.AuthCallbackParams{Code: "", State: uuid.NewString()},
		})

		require.NoError(t, err)
		assert.IsType(t, api.AuthCallback400Response{}, resp)
	})

	t.Run("missing state returns 400", func(t *testing.T) {
		mock := &service.AuthServiceMock{}
		h := handlers.NewAuthHandler(mock)

		resp, err := h.AuthCallback(context.Background(), api.AuthCallbackRequestObject{
			Params: api.AuthCallbackParams{Code: "authcode", State: ""},
		})

		require.NoError(t, err)
		assert.IsType(t, api.AuthCallback400Response{}, resp)
	})

	t.Run("state that is not a valid UUID returns 400", func(t *testing.T) {
		mock := &service.AuthServiceMock{
			// Should never be called since parsing fails first.
			HandleCallbackFn: func(ctx context.Context, stateID uuid.UUID, code string) (model.Session, string, error) {
				t.Fatal("HandleCallback should not be called for an invalid state")
				return model.Session{}, "", nil
			},
		}
		h := handlers.NewAuthHandler(mock)

		resp, err := h.AuthCallback(context.Background(), api.AuthCallbackRequestObject{
			Params: api.AuthCallbackParams{Code: "authcode", State: "not-a-uuid"},
		})

		require.NoError(t, err)
		assert.IsType(t, api.AuthCallback400Response{}, resp)
	})

	t.Run("service error returns 401", func(t *testing.T) {
		state := uuid.New()
		mock := &service.AuthServiceMock{
			HandleCallbackFn: func(ctx context.Context, stateID uuid.UUID, code string) (model.Session, string, error) {
				assert.Equal(t, state, stateID)
				assert.Equal(t, "authcode", code)
				return model.Session{}, "", errors.New("invalid code")
			},
		}
		h := handlers.NewAuthHandler(mock)

		resp, err := h.AuthCallback(context.Background(), api.AuthCallbackRequestObject{
			Params: api.AuthCallbackParams{Code: "authcode", State: state.String()},
		})

		require.NoError(t, err)
		assert.IsType(t, api.AuthCallback401Response{}, resp)
	})

	t.Run("success returns 302 with redirect location and session cookie", func(t *testing.T) {
		state := uuid.New()
		sessionID := uuid.New()
		expiresAt := time.Now().Add(24 * time.Hour).UTC()
		const wantRedirect = "/dashboard"

		session := model.Session{
			Id:        sessionID,
			ExpiresAt: expiresAt,
		}

		mock := &service.AuthServiceMock{
			HandleCallbackFn: func(ctx context.Context, stateID uuid.UUID, code string) (model.Session, string, error) {
				assert.Equal(t, state, stateID)
				assert.Equal(t, "authcode", code)
				return session, wantRedirect, nil
			},
		}
		h := handlers.NewAuthHandler(mock)

		resp, err := h.AuthCallback(context.Background(), api.AuthCallbackRequestObject{
			Params: api.AuthCallbackParams{Code: "authcode", State: state.String()},
		})

		require.NoError(t, err)
		got, ok := resp.(api.AuthCallback302Response)
		require.True(t, ok, "expected AuthCallback302Response, got %T", resp)

		assert.Equal(t, wantRedirect, got.Headers.Location)

		wantCookie := http.Cookie{
			Name:     model.SessionCookieName,
			Value:    sessionID.String(),
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			Expires:  expiresAt,
		}
		assert.Equal(t, wantCookie.String(), got.Headers.SetCookie)
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
		h := handlers.NewAuthHandler(mock)

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
		h := handlers.NewAuthHandler(mock)

		ctx := model.SetSessionInContext(context.Background(), session)

		resp, err := h.AuthLogout(
			ctx,
			api.AuthLogoutRequestObject{},
		)

		require.Error(t, err)
		assert.EqualError(t, err, "failed to logout session")
		assert.Nil(t, resp)
	})

	t.Run("success returns 204", func(t *testing.T) {
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
		h := handlers.NewAuthHandler(mock)

		ctx := model.SetSessionInContext(context.Background(), session)

		resp, err := h.AuthLogout(
			ctx,
			api.AuthLogoutRequestObject{},
		)

		require.NoError(t, err)
		assert.IsType(t, api.AuthLogout204Response{}, resp)
	})
}
