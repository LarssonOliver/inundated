package auth

import (
	"net/http"
	"testing"

	"github.com/larssonoliver/inundated/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestClearSessionCookie(t *testing.T) {
	got := ClearSessionCookie()

	assert.Equal(t, model.SessionCookieName, got.Name)
	assert.Empty(t, got.Value)
	assert.Negative(t, got.MaxAge, "a negative MaxAge is what tells the browser to delete the cookie")
	assert.Equal(t, "/", got.Path)
	assert.True(t, got.HttpOnly)
	assert.True(t, got.Secure)
	assert.Equal(t, http.SameSiteLaxMode, got.SameSite)
}
