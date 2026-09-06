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
	// UserId is the owner of the tag. Nil means unowned - a resource that
	// predates user support and has not yet been adopted by a first user.
	UserId *uuid.UUID
}
