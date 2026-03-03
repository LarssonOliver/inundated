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
}
