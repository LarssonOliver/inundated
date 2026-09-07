package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/larssonoliver/inundated/internal/api"
	"github.com/larssonoliver/inundated/internal/api/handlers"
	"github.com/larssonoliver/inundated/internal/api/middleware"
	"github.com/larssonoliver/inundated/internal/auth"
	"github.com/larssonoliver/inundated/internal/config"
	"github.com/larssonoliver/inundated/internal/repository/memory"
	"github.com/larssonoliver/inundated/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testCSRFKey = []byte("0123456789abcdef0123456789abcdef")

func buildTestServer(oidc auth.OIDCClient) (api.StrictServerInterface, service.Service, *memory.MemoryStore) {
	repo := memory.NewMemoryStore()
	svc := service.NewService(repo)
	authSvc := service.NewAuthService(svc, repo, repo, oidc)
	return api.NewServer(handlers.NewHandler(authSvc, svc)), svc, repo
}

func get(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

// The derived OIDC redirect URI must point at a route the API actually serves
// without a session, or login breaks. config can't import middleware, so this
// guards the two definitions against drift.
func TestDerivedRedirectURIIsAPublicRoute(t *testing.T) {
	cfg, err := config.Load(config.WithArgs(nil), config.WithEnvLookup(func(k string) (string, bool) {
		return map[string]string{
			"OIDC_ISSUER_URL":    "https://issuer.example.com",
			"OIDC_CLIENT_ID":     "id",
			"OIDC_CLIENT_SECRET": "secret",
			"PUBLIC_BASE_URL":    "https://app.example.com",
		}[k], true
	}))
	require.NoError(t, err)

	path := strings.TrimPrefix(cfg.OIDC.RedirectURL, "https://app.example.com")
	assert.Contains(t, middleware.PublicAPIPaths, path)
}

func TestNewRouter_UserlessMode(t *testing.T) {
	cfg := &config.Config{} // OIDC unset
	server, svc, repo := buildTestServer(auth.NewOIDCClient())
	r := newRouter(cfg, svc, repo, server, testCSRFKey)

	t.Run("health is public", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, get(t, r, "/health").Code)
	})

	t.Run("OIDC routes are hidden", func(t *testing.T) {
		assert.Equal(t, http.StatusNotFound, get(t, r, "/api/auth/login").Code)
		assert.Equal(t, http.StatusNotFound, get(t, r, "/api/auth/callback?code=x&state=y").Code)
	})

	t.Run("resource routes are reachable without a session", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, get(t, r, "/api/projects").Code)
	})
}

func TestNewRouter_OIDCMode(t *testing.T) {
	cfg := &config.Config{
		PublicBaseURL: "https://app.example.com",
		OIDC: config.OIDCConfig{
			IssuerURL:    "https://issuer.example.com",
			ClientID:     "id",
			ClientSecret: "secret",
			HTTPTimeout:  time.Second,
		},
	}

	oidcMock := auth.NewOIDCClientMock()
	oidcMock.BeginAuthorizationFn = func(state string) (auth.OIDCAuthorizationRequest, error) {
		return auth.OIDCAuthorizationRequest{Uri: "https://issuer.example.com/authorize?state=" + state, CodeVerifier: "verifier"}, nil
	}

	server, svc, repo := buildTestServer(oidcMock)
	r := newRouter(cfg, svc, repo, server, testCSRFKey)

	t.Run("resource routes require a session", func(t *testing.T) {
		assert.Equal(t, http.StatusUnauthorized, get(t, r, "/api/projects").Code)
	})

	t.Run("login is reachable without a session", func(t *testing.T) {
		rec := get(t, r, "/api/auth/login")
		require.Equal(t, http.StatusFound, rec.Code)
		assert.Contains(t, rec.Header().Get("Location"), "issuer.example.com/authorize")
	})

	t.Run("health is still public", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, get(t, r, "/health").Code)
	})
}
