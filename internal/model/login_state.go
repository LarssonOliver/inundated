package model

import (
	"time"

	"github.com/google/uuid"
)

type LoginState struct {
	Id           uuid.UUID
	RedirectUri  string
	CodeVerifier string
	ExpiresAt    time.Time
}
