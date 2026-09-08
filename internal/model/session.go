package model

import (
	"crypto/sha256"
	"time"

	"github.com/google/uuid"
)

var SessionCookieName = "inundated_session"

type Session struct {
	Id     uuid.UUID
	UserId uuid.UUID
	Sub    string
	// Token is the opaque secret the client carries in its session cookie.
	// Only HashSessionToken(Token) is persisted, so Token is set when a session
	// is created and when a presented cookie is looked up, but never populated
	// from storage.
	Token string
	// CreatedAt is when the session was first established. Sliding renewal
	// extends ExpiresAt but never past CreatedAt + the absolute cap, so a
	// stolen cookie cannot be kept alive indefinitely.
	CreatedAt time.Time
	ExpiresAt time.Time
}

// HashSessionToken returns the digest stored in place of a raw session token,
// so a read of the sessions table yields no usable credentials.
func HashSessionToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}
