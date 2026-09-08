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
	UserId    *uuid.UUID // owner; nil for resources predating user support
}
