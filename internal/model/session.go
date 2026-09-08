package model

import (
	"crypto/sha256"
	"time"

	"github.com/google/uuid"
)

const SessionCookieName = "inundated_session"

// Session is the server-side record of a signed-in browser. The raw session
// token is never a field here: the store keeps only a hash of it (see
// [HashSessionToken]), and callers that need the token pass it alongside.
type Session struct {
	Id        uuid.UUID
	UserId    uuid.UUID
	Sub       string
	CreatedAt time.Time
	ExpiresAt time.Time
}

func HashSessionToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}
