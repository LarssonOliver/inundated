package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/api/middleware"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOIDCAuth(t *testing.T) {
	validUUID := uuid.New()
	userID := uuid.New()
	sessionID := uuid.New()

	// Helper to create a base mock request handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name        string
		cookieValue string
		setupMocks  func(s *repository.SessionRepoMock, u *service.UserServiceMock)
		checkResult func(t *testing.T, res *http.Response, nextCalledWithUser bool, lastSeenCtx context.Context)
	}{
		{
			name:        "No cookie present - passes through without context",
			cookieValue: "",
			setupMocks:  func(s *repository.SessionRepoMock, u *service.UserServiceMock) {},
			checkResult: func(t *testing.T, res *http.Response, nextCalledWithUser bool, lastSeenCtx context.Context) {
				assert.False(t, nextCalledWithUser)
			},
		},
		{
			name:        "Invalid UUID cookie - passes through without context",
			cookieValue: "not-a-uuid",
			setupMocks:  func(s *repository.SessionRepoMock, u *service.UserServiceMock) {},
			checkResult: func(t *testing.T, res *http.Response, nextCalledWithUser bool, lastSeenCtx context.Context) {
				assert.False(t, nextCalledWithUser)
			},
		},
		{
			name:        "Session not found in DB - passes through without context",
			cookieValue: validUUID.String(),
			setupMocks: func(s *repository.SessionRepoMock, u *service.UserServiceMock) {
				s.GetSessionFn = func(ctx context.Context, id uuid.UUID) (model.Session, error) {
					return model.Session{}, errors.New("not found")
				}
			},
			checkResult: func(t *testing.T, res *http.Response, nextCalledWithUser bool, lastSeenCtx context.Context) {
				assert.False(t, nextCalledWithUser)
			},
		},
		{
			name:        "Session expired - deletes session and clears cookie",
			cookieValue: validUUID.String(),
			setupMocks: func(s *repository.SessionRepoMock, u *service.UserServiceMock) {
				s.GetSessionFn = func(ctx context.Context, id uuid.UUID) (model.Session, error) {
					return model.Session{ExpiresAt: time.Now().Add(-1 * time.Hour)}, nil
				}
				s.DeleteSessionFn = func(ctx context.Context, id uuid.UUID) error {
					assert.Equal(t, validUUID, id)
					return nil
				}
			},
			checkResult: func(t *testing.T, res *http.Response, nextCalledWithUser bool, lastSeenCtx context.Context) {
				assert.False(t, nextCalledWithUser)
				cookies := res.Cookies()
				require.Len(t, cookies, 1)
				assert.Equal(t, model.SessionCookieName, cookies[0].Name)
				assert.Equal(t, -1, cookies[0].MaxAge)
			},
		},
		{
			name:        "Valid session - attaches context successfully",
			cookieValue: validUUID.String(),
			setupMocks: func(s *repository.SessionRepoMock, u *service.UserServiceMock) {
				s.GetSessionFn = func(ctx context.Context, id uuid.UUID) (model.Session, error) {
					return model.Session{Id: sessionID, Sub: "sub_123", ExpiresAt: time.Now().Add(12 * time.Hour)}, nil
				}
				u.GetUserBySubFn = func(ctx context.Context, sub string) (model.User, error) {
					assert.Equal(t, "sub_123", sub)
					return model.User{Id: userID}, nil
				}
			},
			checkResult: func(t *testing.T, res *http.Response, nextCalledWithUser bool, lastSeenCtx context.Context) {
				assert.True(t, nextCalledWithUser)
				user, ok := model.GetCurrentUserFromContext(lastSeenCtx)
				require.True(t, ok)
				assert.Equal(t, userID, user.Id)
				session, ok := model.GetSessionFromContext(lastSeenCtx)
				require.True(t, ok)
				assert.Equal(t, sessionID, session.Id)
			},
		},
		{
			name:        "Session user no longer exists - deletes session and clears cookie",
			cookieValue: validUUID.String(),
			setupMocks: func(s *repository.SessionRepoMock, u *service.UserServiceMock) {
				s.GetSessionFn = func(ctx context.Context, id uuid.UUID) (model.Session, error) {
					return model.Session{Id: sessionID, Sub: "sub_gone", ExpiresAt: time.Now().Add(12 * time.Hour)}, nil
				}
				s.DeleteSessionFn = func(ctx context.Context, id uuid.UUID) error {
					assert.Equal(t, validUUID, id)
					return nil
				}
				u.GetUserBySubFn = func(ctx context.Context, sub string) (model.User, error) {
					return model.User{}, model.ErrNotFound
				}
			},
			checkResult: func(t *testing.T, res *http.Response, nextCalledWithUser bool, lastSeenCtx context.Context) {
				assert.False(t, nextCalledWithUser)
				cookies := res.Cookies()
				require.Len(t, cookies, 1)
				assert.Equal(t, model.SessionCookieName, cookies[0].Name)
				assert.Equal(t, -1, cookies[0].MaxAge)
			},
		},
		{
			name:        "Valid session closing in on expiration - touches session",
			cookieValue: validUUID.String(),
			setupMocks: func(s *repository.SessionRepoMock, u *service.UserServiceMock) {
				// Expiring in 2 hours triggers the (< 6 hours) condition
				s.GetSessionFn = func(ctx context.Context, id uuid.UUID) (model.Session, error) {
					return model.Session{Sub: "sub_123", ExpiresAt: time.Now().Add(2 * time.Hour)}, nil
				}
				s.TouchSessionFn = func(ctx context.Context, id uuid.UUID, expiresAt time.Time) (model.Session, error) {
					assert.Equal(t, validUUID, id)
					assert.WithinDuration(t, time.Now().Add(24*time.Hour), expiresAt, 2*time.Second)
					return model.Session{Sub: "sub_123", ExpiresAt: expiresAt}, nil
				}
				u.GetUserBySubFn = func(ctx context.Context, sub string) (model.User, error) {
					return model.User{Id: userID}, nil
				}
			},
			checkResult: func(t *testing.T, res *http.Response, nextCalledWithUser bool, lastSeenCtx context.Context) {
				assert.True(t, nextCalledWithUser)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionMock := &repository.SessionRepoMock{}
			userMock := &service.UserServiceMock{}
			tt.setupMocks(sessionMock, userMock)

			var lastSeenCtx context.Context
			var nextCalledWithUser bool

			// Intercept the inner handler execution to see what context passed down
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				lastSeenCtx = r.Context()
				_, nextCalledWithUser = model.GetCurrentUserFromContext(lastSeenCtx)
				nextHandler.ServeHTTP(w, r)
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			if tt.cookieValue != "" {
				req.AddCookie(&http.Cookie{
					Name:  model.SessionCookieName,
					Value: tt.cookieValue,
				})
			}

			mw := middleware.OIDCAuth(userMock, sessionMock)
			mw(testHandler).ServeHTTP(rec, req)

			res := rec.Result()
			defer func() {
				_ = res.Body.Close()
			}()

			tt.checkResult(t, res, nextCalledWithUser, lastSeenCtx)
		})
	}
}

func TestRequireAuth(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(ctx context.Context) context.Context
		expectedStatus int
	}{
		{
			name: "No user context - Returns 401 Unauthorized",
			setupContext: func(ctx context.Context) context.Context {
				return ctx
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Nil User ID in context - Returns 401 Unauthorized",
			setupContext: func(ctx context.Context) context.Context {
				return model.SetUserInContext(ctx, model.User{Id: uuid.Nil})
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Valid user context - Passes through with 200 OK",
			setupContext: func(ctx context.Context) context.Context {
				return model.SetUserInContext(ctx, model.User{Id: uuid.New()})
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = req.WithContext(tt.setupContext(req.Context()))

			mw := middleware.RequireAuth()
			mw(nextHandler).ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}
