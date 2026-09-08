package model

import (
	"time"

	"github.com/google/uuid"
)

const LoginBindingCookieName = "inundated_login"

type LoginState struct {
	Id           uuid.UUID
	RedirectUri  string
	CodeVerifier string
	Nonce        string
	ExpiresAt    time.Time
}
