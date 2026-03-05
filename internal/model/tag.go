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
}
