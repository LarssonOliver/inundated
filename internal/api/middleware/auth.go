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

// PublicAPIPaths are the API routes reachable without an authenticated session:
// the endpoints that begin and complete the OIDC login flow.
var PublicAPIPaths = []string{"/api/auth/login", "/api/auth/callback"}

// maxSessionLifetime is the absolute age past which a session must be
// re-established, regardless of how recently it was used. It bounds the window
// in which a stolen session cookie is useful.
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
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			// GetSessionByToken never returns the token; keep the presented one
			// so a renewal can re-issue the same cookie.
			session.Token = cookie.Value

			absoluteExpiry := session.CreatedAt.Add(maxSessionLifetime)
			if time.Now().After(session.ExpiresAt) || time.Now().After(absoluteExpiry) {
				_ = sessionRepository.DeleteSession(r.Context(), session.Id)

				http.SetCookie(w, auth.ClearSessionCookie(secure))

				next.ServeHTTP(w, r)
				return
			}

			if session.ExpiresAt.Before(time.Now().Add(6 * time.Hour)) {
				newExpiry := time.Now().Add(24 * time.Hour)
				// Never slide past the absolute cap.
				if newExpiry.After(absoluteExpiry) {
					newExpiry = absoluteExpiry
				}
				// Skip a renewal that would not actually extend the session
				// (near the cap) to avoid a pointless write and cookie rewrite.
				if newExpiry.After(session.ExpiresAt) {
					// A failed renewal is not fatal: the current session is still
					// valid, so keep using it and let the next request retry.
					if renewed, err := sessionRepository.TouchSession(r.Context(), session.Id, newExpiry); err == nil {
						renewed.Token = session.Token
						session = renewed
						http.SetCookie(w, auth.NewSessionCookie(session, secure))
					}
				}
			}

			user, err := userService.GetUserBySub(r.Context(), session.Sub)
			if err != nil {
				// Only tear the session down when the user is genuinely gone. A
				// transient lookup failure (DB blip, timeout) must leave the
				// still-valid session alone so the next request can retry,
				// rather than forcing a full re-login.
				if errors.Is(err, model.ErrNotFound) {
					_ = sessionRepository.DeleteSession(r.Context(), session.Id)
					http.SetCookie(w, auth.ClearSessionCookie(secure))
				}

				next.ServeHTTP(w, r)
				return
			}

			// Downstream handlers only need the session's identity, not its
			// secret; keep the raw token out of the request context.
			session.Token = ""
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
