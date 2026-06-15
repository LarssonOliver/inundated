package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Sub       string    // OIDC subject claim - unique per provider
	Email     string
	Name      string    // Non-nullable - empty string if not provided
	CreatedAt time.Time // Tracked internally, not returned in API
	UpdatedAt time.Time
}

type UpdateUser struct {
	Email *string
	Name  *string
}
