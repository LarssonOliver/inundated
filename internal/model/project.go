package model

import (
	"slices"
	"time"

	"github.com/google/uuid"
)

type Project struct {
	Id         uuid.UUID
	Name       string
	Color      string
	TimeBudget *time.Duration
	// TagId is the project's own project tag (see TagOwnerProject).
	// Timespans and tasks carrying it count toward the project, as do
	// those reaching any of TagIds, which never include it.
	TagId     uuid.UUID
	TagIds    []uuid.UUID
	TotalTime *time.Duration
	// TaskTime is the part of TotalTime logged on the project's tasks.
	TaskTime *time.Duration
	UserId   *uuid.UUID
	Archived bool
}

// EffectiveTagIds returns every tag that counts toward the project: its
// linked TagIds and its own project tag.
func (p Project) EffectiveTagIds() []uuid.UUID {
	if p.TagId == uuid.Nil {
		return p.TagIds
	}
	return append(slices.Clone(p.TagIds), p.TagId)
}
