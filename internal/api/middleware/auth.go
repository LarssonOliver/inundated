package middleware

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/service"
)

func OIDCAuth(userService service.UserService, sessionRepository repository.SessionRepository) func(http.Handler) http.Handler {
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

				http.SetCookie(w, &http.Cookie{
					Name:   model.SessionCookieName,
					Value:  "",
					Path:   "/",
					MaxAge: -1,
				})

				next.ServeHTTP(w, r)
				return
			}

			if session.ExpiresAt.Before(time.Now().Add(6 * time.Hour)) {
				newExpiry := time.Now().Add(24 * time.Hour)
				session, _ = sessionRepository.TouchSession(r.Context(), sessionId, newExpiry)
			}

			user, err := userService.GetOrCreateUserBySub(r.Context(), session.Sub)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := model.SetUserInContext(r.Context(), user)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := model.GetCurrentUserFromContext(r.Context())

			if !ok || user.Id == uuid.Nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
