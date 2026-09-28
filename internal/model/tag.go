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
	UserId    *uuid.UUID
	Archived  bool
	// TaskId is set on task tags. A task tag's name, color and archived
	// state follow its task.
	TaskId *uuid.UUID
}
