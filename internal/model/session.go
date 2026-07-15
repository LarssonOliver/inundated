package model

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	Id        uuid.UUID
	UserId    uuid.UUID
	Sub       string
	ExpiresAt time.Time
}
