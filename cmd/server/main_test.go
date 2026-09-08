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

func buildTestServer(oidc auth.OIDCClient, secureCookies bool) (api.StrictServerInterface, service.Service, *memory.MemoryStore) {
	repo := memory.NewMemoryStore()
	svc := service.NewService(repo)
	authSvc := service.NewAuthService(svc, repo, repo, oidc)
	return api.NewServer(handlers.NewHandler(authSvc, svc, secureCookies)), svc, repo
}

func get(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func csrfHandshake(t *testing.T, h http.Handler) (cookies []*http.Cookie, token string) {
	t.Helper()
	rec := get(t, h, "/api/projects")
	for _, c := range rec.Result().Cookies() {
		cookies = append(cookies, c)
		if c.Name == middleware.XSRFCookieName {
			token = c.Value
		}
	}
	require.NotEmpty(t, token, "expected an %s cookie", middleware.XSRFCookieName)
	return cookies, token
}

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
	server, svc, repo := buildTestServer(auth.NewOIDCClient(), shouldUseSecureCookies(cfg))
	r := newRouter(cfg, svc, repo, server, testCSRFKey)

	t.Run("health is public", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, get(t, r, "/health").Code)
	})

	t.Run("non-API traffic carries no CSRF cookie", func(t *testing.T) {
		for _, c := range get(t, r, "/health").Result().Cookies() {
			assert.NotEqual(t, middleware.XSRFCookieName, c.Name,
				"CSRF machinery must not run outside the API group")
		}
	})

	t.Run("OIDC routes are hidden", func(t *testing.T) {
		assert.Equal(t, http.StatusNotFound, get(t, r, "/api/auth/login").Code)
		assert.Equal(t, http.StatusNotFound, get(t, r, "/api/auth/callback?code=x&state=y").Code)
	})

	t.Run("resource routes are reachable without a session", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, get(t, r, "/api/projects").Code)
	})

	t.Run("an unknown API route is a JSON 404, not the SPA", func(t *testing.T) {
		rec := get(t, r, "/api/does-not-exist")
		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Header().Get("Content-Type"), "json")
		assert.NotContains(t, rec.Body.String(), "<!DOCTYPE html>")
	})

	t.Run("a mutation without a CSRF token is rejected", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/projects",
			strings.NewReader(`{"name":"p","color":"#00ff00"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("a mutation carrying the exposed CSRF token succeeds", func(t *testing.T) {
		cookies, token := csrfHandshake(t, r)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/projects",
			strings.NewReader(`{"name":"p","color":"#00ff00"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", "http://example.com")
		req.Header.Set(middleware.XSRFHeaderName, token)
		for _, c := range cookies {
			req.AddCookie(c)
		}
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
	})
}

func TestNewHTTPServer_HasTimeouts(t *testing.T) {
	s := newHTTPServer("127.0.0.1:0", http.NotFoundHandler())

	assert.Positive(t, s.ReadHeaderTimeout)
	assert.Positive(t, s.ReadTimeout)
	assert.Positive(t, s.WriteTimeout)
	assert.Positive(t, s.IdleTimeout)
}

func TestNewRouter_RateLimiting(t *testing.T) {
	cfg := &config.Config{}
	server, svc, repo := buildTestServer(auth.NewOIDCClient(), shouldUseSecureCookies(cfg))
	r := newRouter(cfg, svc, repo, server, testCSRFKey)

	callFrom := func(path, ip string) int {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.RemoteAddr = ip + ":40000"
		r.ServeHTTP(rec, req)
		return rec.Code
	}

	t.Run("a per-IP flood on the API is shed with 429", func(t *testing.T) {
		const ip = "203.0.113.10"
		got429 := false
		for i := 0; i < middleware.APIRateLimitRequests+5; i++ {
			if callFrom("/api/projects", ip) == http.StatusTooManyRequests {
				got429 = true
				break
			}
		}
		assert.True(t, got429, "expected the API rate limit to kick in within the window")
	})

	t.Run("another IP is unaffected", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, callFrom("/api/projects", "203.0.113.99"))
	})
}

func TestNewRouter_BodyLimit(t *testing.T) {
	cfg := &config.Config{}
	server, svc, repo := buildTestServer(auth.NewOIDCClient(), shouldUseSecureCookies(cfg))
	r := newRouter(cfg, svc, repo, server, testCSRFKey)

	cookies, token := csrfHandshake(t, r)

	oversized := `{"name":"` + strings.Repeat("x", int(middleware.MaxAPIBodyBytes)+1) + `"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/projects", strings.NewReader(oversized))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set(middleware.XSRFHeaderName, token)
	req.RemoteAddr = "198.51.100.5:40000"
	for _, c := range cookies {
		req.AddCookie(c)
	}
	r.ServeHTTP(rec, req)

	assert.NotEqual(t, http.StatusCreated, rec.Code, "an oversized body must never reach the handler")
	assert.GreaterOrEqual(t, rec.Code, 400)
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
	oidcMock.BeginAuthorizationFn = func(state string, nonce string) (auth.OIDCAuthorizationRequest, error) {
		return auth.OIDCAuthorizationRequest{Uri: "https://issuer.example.com/authorize?state=" + state + "&nonce=" + nonce, CodeVerifier: "verifier"}, nil
	}

	server, svc, repo := buildTestServer(oidcMock, shouldUseSecureCookies(cfg))
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

	t.Run("an unknown API route is a JSON 404, not the SPA", func(t *testing.T) {
		rec := get(t, r, "/api/does-not-exist")
		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Header().Get("Content-Type"), "json")
	})
}
