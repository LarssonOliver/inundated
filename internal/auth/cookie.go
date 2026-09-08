package auth

import (
	"net/http"

	"github.com/larssonoliver/inundated/internal/model"
)

func BaseCookie(name string, secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Path:     "/",
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
}

func NewSessionCookie(session model.Session, secure bool) *http.Cookie {
	c := BaseCookie(model.SessionCookieName, secure)
	c.Value = session.Token
	c.HttpOnly = true
	c.Expires = session.ExpiresAt
	return c
}

func ClearSessionCookie(secure bool) *http.Cookie {
	c := BaseCookie(model.SessionCookieName, secure)
	c.HttpOnly = true
	c.MaxAge = -1
	return c
}
