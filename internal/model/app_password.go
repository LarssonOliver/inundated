package model

import (
	"time"

	"github.com/google/uuid"
)

// AppPassword is a long-lived credential for clients that cannot sign in
// through the browser. The raw token is never a field here: the store keeps
// only a hash of it (see [HashToken]), and callers that need the token pass it
// alongside.
type AppPassword struct {
	Id         uuid.UUID
	Name       string
	CreatedAt  time.Time
	LastUsedAt *time.Time
	UserId     *uuid.UUID
}
