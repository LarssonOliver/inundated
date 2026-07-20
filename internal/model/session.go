package model

import (
	"time"

	"github.com/google/uuid"
)

var SessionCookieName = "inundated_session"

type Session struct {
	Id        uuid.UUID
	UserId    uuid.UUID
	Sub       string
	ExpiresAt time.Time
}
