package auth

import (
	"net/http"
	"time"

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

func NewSessionCookie(token string, expiresAt time.Time, secure bool) *http.Cookie {
	c := BaseCookie(model.SessionCookieName, secure)
	c.Value = token
	c.HttpOnly = true
	c.Expires = expiresAt
	return c
}

func ClearSessionCookie(secure bool) *http.Cookie {
	c := BaseCookie(model.SessionCookieName, secure)
	c.HttpOnly = true
	c.MaxAge = -1
	return c
}
