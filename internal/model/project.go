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
	// TaskTime is the part of TotalTime logged on the project's tasks.
	TaskTime *time.Duration
	UserId   *uuid.UUID
	Archived bool
}
