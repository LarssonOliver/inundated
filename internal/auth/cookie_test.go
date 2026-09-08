package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestBaseCookie(t *testing.T) {
	got := BaseCookie("whatever", true)
	assert.Equal(t, "whatever", got.Name)
	assert.Equal(t, "/", got.Path)
	assert.Equal(t, http.SameSiteLaxMode, got.SameSite)
	assert.True(t, got.Secure)
	assert.False(t, BaseCookie("whatever", false).Secure)
}

func TestNewSessionCookie(t *testing.T) {
	id := uuid.New()
	token := "opaque-session-token"
	expiresAt := time.Now().Add(24 * time.Hour)

	t.Run("secure origin sets the Secure attribute and carries the token, not the id", func(t *testing.T) {
		got := NewSessionCookie(token, expiresAt, true)
		assert.Equal(t, model.SessionCookieName, got.Name)
		assert.Equal(t, token, got.Value)
		assert.NotContains(t, got.Value, id.String())
		assert.True(t, got.HttpOnly)
		assert.True(t, got.Secure)
		assert.Equal(t, "/", got.Path)
		assert.Equal(t, http.SameSiteLaxMode, got.SameSite)
		assert.Equal(t, expiresAt, got.Expires)
	})

	t.Run("insecure origin omits the Secure attribute so the browser keeps the cookie", func(t *testing.T) {
		got := NewSessionCookie(token, expiresAt, false)
		assert.False(t, got.Secure)
	})
}

func TestClearSessionCookie(t *testing.T) {
	got := ClearSessionCookie(true)

	assert.Equal(t, model.SessionCookieName, got.Name)
	assert.Empty(t, got.Value)
	assert.Negative(t, got.MaxAge, "a negative MaxAge is what tells the browser to delete the cookie")
	assert.Equal(t, "/", got.Path)
	assert.True(t, got.HttpOnly)
	assert.True(t, got.Secure)
	assert.Equal(t, http.SameSiteLaxMode, got.SameSite)

	assert.False(t, ClearSessionCookie(false).Secure, "an insecure origin must be able to clear its cookie too")
}
