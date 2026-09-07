package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestNewSessionCookie(t *testing.T) {
	session := model.Session{Id: uuid.New(), ExpiresAt: time.Now().Add(24 * time.Hour)}

	t.Run("secure origin sets the Secure attribute", func(t *testing.T) {
		got := NewSessionCookie(session, true)
		assert.Equal(t, model.SessionCookieName, got.Name)
		assert.Equal(t, session.Id.String(), got.Value)
		assert.True(t, got.HttpOnly)
		assert.True(t, got.Secure)
	})

	t.Run("insecure origin omits the Secure attribute so the browser keeps the cookie", func(t *testing.T) {
		got := NewSessionCookie(session, false)
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
