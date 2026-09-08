package model

import (
	"time"

	"github.com/google/uuid"
)

type LoginState struct {
	Id           uuid.UUID
	RedirectUri  string
	CodeVerifier string
	// Nonce is the per-login value embedded in the authorization request and
	// echoed back in the ID token's nonce claim. Verifying it on callback binds
	// the returned token to this authentication request (OIDC replay defense).
	Nonce     string
	ExpiresAt time.Time
}
