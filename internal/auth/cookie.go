package auth

import (
	"net/http"

	"github.com/larssonoliver/inundated/internal/model"
)

// NewSessionCookie builds the session cookie. secure controls the Secure
// attribute and must be off for a plain-HTTP origin, or the browser drops the
// cookie and the user can never stay signed in.
func NewSessionCookie(session model.Session, secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     model.SessionCookieName,
		Value:    session.Id.String(),
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  session.ExpiresAt,
	}
}

// ClearSessionCookie returns a cookie that instructs the browser to drop the
// session cookie. It is emitted on logout and whenever the middleware finds a
// cookie pointing at a session that no longer exists. secure must match the
// value used for [NewSessionCookie] so the browser recognises the same cookie.
func ClearSessionCookie(secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     model.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
}
