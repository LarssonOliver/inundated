package auth

import (
	"net/http"

	"github.com/larssonoliver/inundated/internal/model"
)

func NewSessionCookie(session model.Session) *http.Cookie {
	return &http.Cookie{
		Name:     model.SessionCookieName,
		Value:    session.Id.String(),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  session.ExpiresAt,
	}
}

// ClearSessionCookie returns a cookie that instructs the browser to drop the
// session cookie. It is emitted on logout and whenever the middleware finds a
// cookie pointing at a session that no longer exists.
func ClearSessionCookie() *http.Cookie {
	return &http.Cookie{
		Name:     model.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}
