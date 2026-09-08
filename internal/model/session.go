package model

import (
	"time"

	"github.com/google/uuid"
)

var SessionCookieName = "inundated_session"

type Session struct {
	Id     uuid.UUID
	UserId uuid.UUID
	Sub    string
	// CreatedAt is when the session was first established. Sliding renewal
	// extends ExpiresAt but never past CreatedAt + the absolute cap, so a
	// stolen cookie cannot be kept alive indefinitely.
	CreatedAt time.Time
	ExpiresAt time.Time
}
