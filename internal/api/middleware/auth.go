package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/auth"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/service"
)

// PublicAPIPaths are the API routes reachable without an authenticated session:
// the endpoints that begin and complete the OIDC login flow.
var PublicAPIPaths = []string{"/api/auth/login", "/api/auth/callback"}

func OIDCAuth(userService service.UserService, sessionRepository repository.SessionRepository, secure bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			cookie, err := r.Cookie(model.SessionCookieName)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			sessionId, err := uuid.Parse(cookie.Value)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			session, err := sessionRepository.GetSession(r.Context(), sessionId)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			if time.Now().After(session.ExpiresAt) {
				_ = sessionRepository.DeleteSession(r.Context(), sessionId)

				http.SetCookie(w, auth.ClearSessionCookie(secure))

				next.ServeHTTP(w, r)
				return
			}

			if session.ExpiresAt.Before(time.Now().Add(6 * time.Hour)) {
				newExpiry := time.Now().Add(24 * time.Hour)
				// A failed renewal is not fatal: the current session is still
				// valid, so keep using it and let the next request retry.
				if renewed, err := sessionRepository.TouchSession(r.Context(), sessionId, newExpiry); err == nil {
					session = renewed
					http.SetCookie(w, auth.NewSessionCookie(session, secure))
				}
			}

			user, err := userService.GetUserBySub(r.Context(), session.Sub)
			if err != nil {
				_ = sessionRepository.DeleteSession(r.Context(), sessionId)

				http.SetCookie(w, auth.ClearSessionCookie(secure))

				next.ServeHTTP(w, r)
				return
			}

			ctx := model.SetSessionInContext(r.Context(), session)
			ctx = model.SetUserInContext(ctx, user)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuth rejects requests that have no authenticated user in context with
// 401, except for requests whose path exactly matches one of publicPaths.
func RequireAuth(publicPaths ...string) func(http.Handler) http.Handler {
	public := make(map[string]bool, len(publicPaths))
	for _, p := range publicPaths {
		public[p] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if public[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			user, ok := model.GetCurrentUserFromContext(r.Context())

			if !ok || user.Id == uuid.Nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RejectPathPrefixes returns 404 for any request whose path is under one of the
// given prefixes. It is mounted in userless mode to hide the OIDC routes, which
// have no provider to talk to.
func RejectPathPrefixes(prefixes ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, prefix := range prefixes {
				if strings.HasPrefix(r.URL.Path, prefix) {
					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`{"message":"not found"}`))
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
