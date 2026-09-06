package model

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	Id         uuid.UUID
	Name       string
	Color      string
	TimeBudget *time.Duration
	TagIds     []uuid.UUID
	TotalTime  *time.Duration
	// UserId is the owner of the project. Nil means unowned - a resource
	// that predates user support and has not yet been adopted by a first user.
	UserId *uuid.UUID
}
