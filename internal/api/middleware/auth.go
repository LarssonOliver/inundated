package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/auth"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/service"
)

var PublicAPIPaths = []string{"/api/auth/login", "/api/auth/callback"}

const maxSessionLifetime = 7 * 24 * time.Hour

func OIDCAuth(userService service.UserService, sessionRepository repository.SessionRepository, secure bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			cookie, err := r.Cookie(model.SessionCookieName)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			session, err := sessionRepository.GetSessionByToken(r.Context(), cookie.Value)
			if errors.Is(err, model.ErrNotFound) || errors.Is(err, model.ErrInvalidArgument) {
				next.ServeHTTP(w, r)
				return
			}
			if err != nil {
				respondServiceUnavailable(w)
				return
			}

			absoluteExpiry := session.CreatedAt.Add(maxSessionLifetime)
			if time.Now().After(session.ExpiresAt) || time.Now().After(absoluteExpiry) {
				_ = sessionRepository.DeleteSession(r.Context(), session.Id)

				http.SetCookie(w, auth.ClearSessionCookie(secure))

				next.ServeHTTP(w, r)
				return
			}

			if session.ExpiresAt.Before(time.Now().Add(6 * time.Hour)) {
				newExpiry := time.Now().Add(24 * time.Hour)
				if newExpiry.After(absoluteExpiry) {
					newExpiry = absoluteExpiry
				}
				if newExpiry.After(session.ExpiresAt) {
					if renewed, err := sessionRepository.TouchSession(r.Context(), session.Id, newExpiry); err == nil {
						session = renewed
						// Re-issue the token the client presented, with the
						// extended expiry.
						http.SetCookie(w, auth.NewSessionCookie(cookie.Value, session.ExpiresAt, secure))
					}
				}
			}

			user, err := userService.GetUserBySub(r.Context(), session.Sub)
			if errors.Is(err, model.ErrNotFound) {
				_ = sessionRepository.DeleteSession(r.Context(), session.Id)
				http.SetCookie(w, auth.ClearSessionCookie(secure))
				next.ServeHTTP(w, r)
				return
			}
			if err != nil {
				respondServiceUnavailable(w)
				return
			}

			ctx := model.SetSessionInContext(r.Context(), session)
			ctx = model.SetUserInContext(ctx, user)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func respondServiceUnavailable(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusServiceUnavailable)
	_, _ = w.Write([]byte(`{"message":"authentication temporarily unavailable"}`))
}

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
