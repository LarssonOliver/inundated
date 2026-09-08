package auth

import (
	"net/http"

	"github.com/larssonoliver/inundated/internal/model"
)

// BaseCookie returns a cookie pre-filled with the attributes every cookie this
// app sets shares: rooted at "/", Lax same-site, and Secure everywhere except a
// plain-HTTP local origin (where a Secure cookie would never reach the browser).
// Callers set Value, HttpOnly and any expiry. Centralising this keeps a new
// cookie from silently omitting SameSite or Path.
func BaseCookie(name string, secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Path:     "/",
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
}

// NewSessionCookie builds the session cookie. Its value is the session's opaque
// token (only the hash of which is stored server-side), so session.Token must be
// populated. secure controls the Secure attribute and must be off for a
// plain-HTTP origin, or the browser drops the cookie and the user can never stay
// signed in.
func NewSessionCookie(session model.Session, secure bool) *http.Cookie {
	c := BaseCookie(model.SessionCookieName, secure)
	c.Value = session.Token
	c.HttpOnly = true
	c.Expires = session.ExpiresAt
	return c
}

// ClearSessionCookie returns a cookie that instructs the browser to drop the
// session cookie. It is emitted on logout and whenever the middleware finds a
// cookie pointing at a session that no longer exists. secure must match the
// value used for [NewSessionCookie] so the browser recognises the same cookie.
func ClearSessionCookie(secure bool) *http.Cookie {
	c := BaseCookie(model.SessionCookieName, secure)
	c.HttpOnly = true
	c.MaxAge = -1
	return c
}
