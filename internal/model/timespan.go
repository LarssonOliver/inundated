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
	// CaldavName is the object name a CalDAV client chose when it created
	// the timespan; nil otherwise.
	CaldavName *string
}
