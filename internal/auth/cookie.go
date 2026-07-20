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
