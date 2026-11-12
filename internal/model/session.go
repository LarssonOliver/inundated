package model

import (
	"time"

	"github.com/google/uuid"
)

const SessionCookieName = "inundated_session"

// Session is the server-side record of a signed-in browser. The raw session
// token is never a field here: the store keeps only a hash of it (see
// [HashToken]), and callers that need the token pass it alongside.
type Session struct {
	Id        uuid.UUID
	UserId    uuid.UUID
	Sub       string
	CreatedAt time.Time
	ExpiresAt time.Time
}
