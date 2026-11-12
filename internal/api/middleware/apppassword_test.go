package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/api/middleware"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const challenge = `Basic realm="inundated"`

type appPasswordHarness struct {
	handler           http.Handler
	nextCalls         int
	lastContext       context.Context
	authenticateCalls int
}

func newAppPasswordHarness(requireOwner bool, authenticate func(token string) (model.AppPassword, error)) *appPasswordHarness {
	h := &appPasswordHarness{}
	svc := &service.AppPasswordServiceMock{
		AuthenticateFn: func(ctx context.Context, token string) (model.AppPassword, error) {
			h.authenticateCalls++
			return authenticate(token)
		},
	}
	h.handler = middleware.AppPasswordAuth(svc, requireOwner)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.nextCalls++
		h.lastContext = r.Context()
		w.WriteHeader(http.StatusOK)
	}))
	return h
}

func (h *appPasswordHarness) do(remoteAddr, password string, withAuth bool) *http.Response {
	req := httptest.NewRequest("PROPFIND", "/dav/", nil)
	req.RemoteAddr = remoteAddr
	if withAuth {
		req.SetBasicAuth("ignored", password)
	}
	rec := httptest.NewRecorder()
	h.handler.ServeHTTP(rec, req)
	return rec.Result()
}

func TestAppPasswordAuth(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name           string
		requireOwner   bool
		withAuth       bool
		authenticate   func(token string) (model.AppPassword, error)
		wantStatus     int
		wantNextCalled bool
		wantUserID     *uuid.UUID
	}{
		{
			name:     "no Authorization header",
			withAuth: false,
			authenticate: func(string) (model.AppPassword, error) {
				t.Fatal("must not authenticate")
				return model.AppPassword{}, nil
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "empty password",
			withAuth: true,
			authenticate: func(string) (model.AppPassword, error) {
				t.Fatal("must not authenticate")
				return model.AppPassword{}, nil
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:         "unknown token",
			withAuth:     true,
			authenticate: func(string) (model.AppPassword, error) { return model.AppPassword{}, model.ErrNotFound },
			wantStatus:   http.StatusUnauthorized,
		},
		{
			name:           "owned password",
			withAuth:       true,
			authenticate:   func(string) (model.AppPassword, error) { return model.AppPassword{UserId: &userID}, nil },
			wantStatus:     http.StatusOK,
			wantNextCalled: true,
			wantUserID:     &userID,
		},
		{
			name:           "unowned password, owner not required",
			withAuth:       true,
			authenticate:   func(string) (model.AppPassword, error) { return model.AppPassword{}, nil },
			wantStatus:     http.StatusOK,
			wantNextCalled: true,
		},
		{
			name:         "unowned password, owner required",
			requireOwner: true,
			withAuth:     true,
			authenticate: func(string) (model.AppPassword, error) { return model.AppPassword{}, nil },
			wantStatus:   http.StatusUnauthorized,
		},
		{
			name:         "service error",
			withAuth:     true,
			authenticate: func(string) (model.AppPassword, error) { return model.AppPassword{}, errors.New("database down") },
			wantStatus:   http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newAppPasswordHarness(tt.requireOwner, tt.authenticate)
			password := "secret"
			if tt.name == "empty password" {
				password = ""
			}

			res := h.do("192.0.2.1:1234", password, tt.withAuth)

			assert.Equal(t, tt.wantStatus, res.StatusCode)
			assert.Equal(t, tt.wantNextCalled, h.nextCalls == 1)
			if tt.wantStatus == http.StatusUnauthorized {
				assert.Equal(t, challenge, res.Header.Get("WWW-Authenticate"))
			}
			if tt.wantNextCalled {
				user, ok := model.GetCurrentUserFromContext(h.lastContext)
				if tt.wantUserID != nil {
					require.True(t, ok)
					assert.Equal(t, *tt.wantUserID, user.Id)
				} else {
					assert.False(t, ok)
				}
			}
		})
	}
}

func TestAppPasswordAuthSendsPasswordAsToken(t *testing.T) {
	var received string
	h := newAppPasswordHarness(false, func(token string) (model.AppPassword, error) {
		received = token
		return model.AppPassword{}, nil
	})

	h.do("192.0.2.1:1234", "the-token", true)

	assert.Equal(t, "the-token", received)
}

func TestAppPasswordAuthServiceErrorsDoNotCountAsFailures(t *testing.T) {
	failing := true
	h := newAppPasswordHarness(false, func(string) (model.AppPassword, error) {
		if failing {
			return model.AppPassword{}, errors.New("database down")
		}
		return model.AppPassword{}, nil
	})

	for range middleware.AppPasswordFailureLimit + 5 {
		res := h.do("192.0.2.1:1234", "secret", true)
		require.Equal(t, http.StatusServiceUnavailable, res.StatusCode)
	}
	failing = false

	assert.Equal(t, http.StatusOK, h.do("192.0.2.1:1234", "secret", true).StatusCode)
}

func TestAppPasswordAuthFailureLimit(t *testing.T) {
	valid := false
	h := newAppPasswordHarness(false, func(string) (model.AppPassword, error) {
		if valid {
			return model.AppPassword{}, nil
		}
		return model.AppPassword{}, model.ErrNotFound
	})

	for range middleware.AppPasswordFailureLimit {
		res := h.do("192.0.2.1:1234", "wrong", true)
		require.Equal(t, http.StatusUnauthorized, res.StatusCode)
	}
	valid = true
	callsBefore := h.authenticateCalls

	res := h.do("192.0.2.1:5678", "right", true)

	assert.Equal(t, http.StatusTooManyRequests, res.StatusCode)
	assert.Equal(t, callsBefore, h.authenticateCalls)
	assert.Zero(t, h.nextCalls)

	other := h.do("192.0.2.2:1234", "right", true)
	assert.Equal(t, http.StatusOK, other.StatusCode)
}

func TestAppPasswordAuthMissingHeaderDoesNotCountAsFailure(t *testing.T) {
	h := newAppPasswordHarness(false, func(token string) (model.AppPassword, error) {
		if token == "right" {
			return model.AppPassword{}, nil
		}
		return model.AppPassword{}, model.ErrNotFound
	})

	for range middleware.AppPasswordFailureLimit + 5 {
		res := h.do("192.0.2.1:1234", "", false)
		require.Equal(t, http.StatusUnauthorized, res.StatusCode)
		require.NotEmpty(t, res.Header.Get("WWW-Authenticate"))
	}

	assert.Equal(t, http.StatusOK, h.do("192.0.2.1:1234", "right", true).StatusCode)
}
