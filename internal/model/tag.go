package model

import (
	"time"

	"github.com/google/uuid"
)

type Tag struct {
	Id        uuid.UUID
	Name      string
	Color     string
	TotalTime *time.Duration
	UserId    *uuid.UUID // owner; nil for resources predating user support
}
