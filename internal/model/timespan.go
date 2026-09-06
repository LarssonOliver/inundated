package model

import (
	"time"

	"github.com/google/uuid"
)

type Timespan struct {
	Id        uuid.UUID
	Name      string
	StartTime time.Time
	EndTime   time.Time
	TagIds    []uuid.UUID
	// UserId is the owner of the timespan. Nil means unowned - a resource that
	// predates user support and has not yet been adopted by a first user.
	UserId *uuid.UUID
}
