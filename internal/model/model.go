package model

import (
	"time"

	"github.com/google/uuid"
)

type Tag struct {
	Id    uuid.UUID
	Name  string
	Color string
}

type Project struct {
	Id         uuid.UUID
	Name       string
	Color      string
	TimeBudget time.Duration
	TagIds     []uuid.UUID
}

type TimeSpan struct {
	Id        uuid.UUID
	Name      string
	StartTime time.Time
	EndTime   time.Time
	TagIds    []uuid.UUID
}
