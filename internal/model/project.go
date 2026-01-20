package model

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	Id         uuid.UUID
	Name       string
	Color      string
	TimeBudget time.Duration
	TagIds     []uuid.UUID
}
