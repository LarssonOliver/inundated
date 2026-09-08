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
			name:        "Token matches no session - passes through without context",
			cookieValue: "some-unknown-token",
			setupMocks: func(s *repository.SessionRepoMock, u *service.UserServiceMock) {
				s.GetSessionByTokenFn = func(ctx context.Context, token string) (model.Session, error) {
					return model.Session{}, model.ErrNotFound
				}
			},
			checkResult: func(t *testing.T, res *http.Response, nextCalledWithUser bool, lastSeenCtx context.Context) {
				assert.False(t, nextCalledWithUser)
			},
		},
		{
			name:        "Session lookup errors transiently - 503, does not fall through to a 401",
			cookieValue: validUUID.String(),
			setupMocks: func(s *repository.SessionRepoMock, u *service.UserServiceMock) {
				s.GetSessionByTokenFn = func(ctx context.Context, token string) (model.Session, error) {
					return model.Session{}, errors.New("connection reset")
				}
			},
			checkResult: func(t *testing.T, res *http.Response, nextCalledWithUser bool, lastSeenCtx context.Context) {
				assert.False(t, nextCalledWithUser)
				assert.Nil(t, lastSeenCtx, "the request must not reach the next handler on a transient error")
				assert.Equal(t, http.StatusServiceUnavailable, res.StatusCode)
			},
		},
		{
			name:        "Session expired - deletes session and clears cookie",
			cookieValue: validUUID.String(),
			setupMocks: func(s *repository.SessionRepoMock, u *service.UserServiceMock) {
				s.GetSessionByTokenFn = func(ctx context.Context, token string) (model.Session, error) {
					return model.Session{Id: sessionID, ExpiresAt: time.Now().Add(-1 * time.Hour)}, nil
				}
				s.DeleteSessionFn = func(ctx context.Context, id uuid.UUID) error {
					assert.Equal(t, sessionID, id)
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
			name:        "Session past the absolute lifetime cap - deletes session and clears cookie even though ExpiresAt is in the future",
			cookieValue: validUUID.String(),
			setupMocks: func(s *repository.SessionRepoMock, u *service.UserServiceMock) {
				s.GetSessionByTokenFn = func(ctx context.Context, token string) (model.Session, error) {
					return model.Session{
						Id:        sessionID,
						Sub:       "sub_123",
						CreatedAt: time.Now().Add(-8 * 24 * time.Hour),
						ExpiresAt: time.Now().Add(12 * time.Hour),
					}, nil
				}
				s.DeleteSessionFn = func(ctx context.Context, id uuid.UUID) error {
					assert.Equal(t, sessionID, id)
					return nil
				}
				s.TouchSessionFn = func(ctx context.Context, id uuid.UUID, expiresAt time.Time) (model.Session, error) {
					t.Fatal("a session past the absolute cap must not be renewed")
					return model.Session{}, nil
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
			name:        "Renewal near the absolute cap is clamped and does not exceed CreatedAt + 7d",
			cookieValue: validUUID.String(),
			setupMocks: func(s *repository.SessionRepoMock, u *service.UserServiceMock) {
				createdAt := time.Now().Add(-6*24*time.Hour - 20*time.Hour) // ~6d20h old
				s.GetSessionByTokenFn = func(ctx context.Context, token string) (model.Session, error) {
					return model.Session{
						Id:        sessionID,
						Sub:       "sub_123",
						CreatedAt: createdAt,
						ExpiresAt: time.Now().Add(1 * time.Hour), // within the 6h renewal window
					}, nil
				}
				s.TouchSessionFn = func(ctx context.Context, id uuid.UUID, expiresAt time.Time) (model.Session, error) {
					// The absolute cap is 7 days from CreatedAt.
					assert.WithinDuration(t, createdAt.Add(7*24*time.Hour), expiresAt, 2*time.Second,
						"renewal must be clamped to the absolute cap")
					return model.Session{Id: sessionID, Sub: "sub_123", CreatedAt: createdAt, ExpiresAt: expiresAt}, nil
				}
				u.GetUserBySubFn = func(ctx context.Context, sub string) (model.User, error) {
					return model.User{Id: userID}, nil
				}
			},
			checkResult: func(t *testing.T, res *http.Response, nextCalledWithUser bool, lastSeenCtx context.Context) {
				assert.True(t, nextCalledWithUser)
			},
		},
		{
			name:        "Valid session - attaches context successfully",
			cookieValue: validUUID.String(),
			setupMocks: func(s *repository.SessionRepoMock, u *service.UserServiceMock) {
				s.GetSessionByTokenFn = func(ctx context.Context, token string) (model.Session, error) {
					return model.Session{Id: sessionID, Sub: "sub_123", CreatedAt: time.Now(), ExpiresAt: time.Now().Add(12 * time.Hour)}, nil
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
				s.GetSessionByTokenFn = func(ctx context.Context, token string) (model.Session, error) {
					return model.Session{Id: sessionID, Sub: "sub_gone", CreatedAt: time.Now(), ExpiresAt: time.Now().Add(12 * time.Hour)}, nil
				}
				s.DeleteSessionFn = func(ctx context.Context, id uuid.UUID) error {
					assert.Equal(t, sessionID, id)
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
			name:        "GetUserBySub fails transiently - 503, keeps the session, does not fall through to a 401",
			cookieValue: validUUID.String(),
			setupMocks: func(s *repository.SessionRepoMock, u *service.UserServiceMock) {
				s.GetSessionByTokenFn = func(ctx context.Context, token string) (model.Session, error) {
					return model.Session{Id: sessionID, Sub: "sub_123", CreatedAt: time.Now(), ExpiresAt: time.Now().Add(12 * time.Hour)}, nil
				}
				s.DeleteSessionFn = func(ctx context.Context, id uuid.UUID) error {
					t.Fatal("a valid session must not be deleted when the user lookup fails transiently")
					return nil
				}
				u.GetUserBySubFn = func(ctx context.Context, sub string) (model.User, error) {
					return model.User{}, errors.New("connection reset by peer")
				}
			},
			checkResult: func(t *testing.T, res *http.Response, nextCalledWithUser bool, lastSeenCtx context.Context) {
				assert.False(t, nextCalledWithUser)
				assert.Nil(t, lastSeenCtx, "the request must not reach the next handler on a transient error")
				assert.Equal(t, http.StatusServiceUnavailable, res.StatusCode)
				assert.Empty(t, res.Cookies(), "the session cookie must not be cleared on a transient lookup error")
			},
		},
		{
			name:        "Valid session closing in on expiration - touches session",
			cookieValue: validUUID.String(),
			setupMocks: func(s *repository.SessionRepoMock, u *service.UserServiceMock) {
				// Expiring in 2 hours triggers the (< 6 hours) condition
				s.GetSessionByTokenFn = func(ctx context.Context, token string) (model.Session, error) {
					return model.Session{Id: sessionID, Sub: "sub_123", CreatedAt: time.Now(), ExpiresAt: time.Now().Add(2 * time.Hour)}, nil
				}
				s.TouchSessionFn = func(ctx context.Context, id uuid.UUID, expiresAt time.Time) (model.Session, error) {
					assert.Equal(t, sessionID, id)
					assert.WithinDuration(t, time.Now().Add(24*time.Hour), expiresAt, 2*time.Second)
					return model.Session{Id: sessionID, Sub: "sub_123", CreatedAt: time.Now(), ExpiresAt: expiresAt}, nil
				}
				u.GetUserBySubFn = func(ctx context.Context, sub string) (model.User, error) {
					return model.User{Id: userID}, nil
				}
			},
			checkResult: func(t *testing.T, res *http.Response, nextCalledWithUser bool, lastSeenCtx context.Context) {
				assert.True(t, nextCalledWithUser)
				cookies := res.Cookies()
				require.Len(t, cookies, 1, "a renewed session must re-issue the cookie")
				assert.Equal(t, validUUID.String(), cookies[0].Value,
					"renewal must re-issue the token the client presented, not lose it")
			},
		},
		{
			name:        "Session renewal fails transiently - keeps the still-valid session and authenticates",
			cookieValue: validUUID.String(),
			setupMocks: func(s *repository.SessionRepoMock, u *service.UserServiceMock) {
				s.GetSessionByTokenFn = func(ctx context.Context, token string) (model.Session, error) {
					return model.Session{Id: sessionID, Sub: "sub_123", CreatedAt: time.Now(), ExpiresAt: time.Now().Add(2 * time.Hour)}, nil
				}
				s.TouchSessionFn = func(ctx context.Context, id uuid.UUID, expiresAt time.Time) (model.Session, error) {
					return model.Session{}, errors.New("transient database error")
				}
				s.DeleteSessionFn = func(ctx context.Context, id uuid.UUID) error {
					t.Fatal("a still-valid session must not be deleted when renewal fails")
					return nil
				}
				u.GetUserBySubFn = func(ctx context.Context, sub string) (model.User, error) {
					assert.Equal(t, "sub_123", sub, "the original session's subject must survive a failed renewal")
					return model.User{Id: userID}, nil
				}
			},
			checkResult: func(t *testing.T, res *http.Response, nextCalledWithUser bool, lastSeenCtx context.Context) {
				assert.True(t, nextCalledWithUser)
				assert.Empty(t, res.Cookies(), "no session cookie should be rewritten when renewal fails")
				session, ok := model.GetSessionFromContext(lastSeenCtx)
				require.True(t, ok)
				assert.Equal(t, sessionID, session.Id)
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

			mw := middleware.OIDCAuth(userMock, sessionMock, true)
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

func TestRequireAuthPublicPaths(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := middleware.RequireAuth("/api/auth/login", "/api/auth/callback")

	t.Run("exempt path passes through without a user", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/auth/login", nil)
		mw(nextHandler).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("non-exempt path still requires a user", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
		mw(nextHandler).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("exempt prefix is not enough - match must be exact", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/auth/login/extra", nil)
		mw(nextHandler).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestRejectPathPrefixes(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := middleware.RejectPathPrefixes("/api/auth/")

	t.Run("path under a rejected prefix returns 404", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/auth/login", nil)
		mw(nextHandler).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("other paths pass through", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
		mw(nextHandler).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
