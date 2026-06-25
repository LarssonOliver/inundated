package model

import (
	"github.com/google/uuid"
)

const (
	ContextKeyCurrentUserId = "current_user_id"
)

type User struct {
	Id    uuid.UUID
	Sub   string // OIDC subject claim - unique per provider
	Email string
	Name  string // Non-nullable - empty string if not provided
}
