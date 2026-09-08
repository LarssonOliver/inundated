package model

import (
	"crypto/sha256"
	"time"

	"github.com/google/uuid"
)

const SessionCookieName = "inundated_session"

type Session struct {
	Id        uuid.UUID
	UserId    uuid.UUID
	Sub       string
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
}

func HashSessionToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}
