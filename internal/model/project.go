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
	// ArchivedAt is when the project was archived, set by the store while
	// Archived is. Ignored on update.
	ArchivedAt *time.Time
}

// ArchivedAtAfter returns when an item archived since prev (nil if it
// wasn't) is archived once its archived flag is set to archived at now:
// an item that stays archived keeps its original time.
func ArchivedAtAfter(prev *time.Time, archived bool, now time.Time) *time.Time {
	switch {
	case !archived:
		return nil
	case prev != nil:
		return prev
	default:
		return &now
	}
}

// CompareArchived orders active items before archived ones, and archived
// ones most recently archived first. Items it ranks equal are left for the
// caller to order, e.g. by name.
func CompareArchived(a, b *time.Time) int {
	switch {
	case a == nil && b == nil:
		return 0
	case a == nil:
		return -1
	case b == nil:
		return 1
	default:
		return b.Compare(*a)
	}
}

// EffectiveTagIds returns every tag that counts toward the project: its
// linked TagIds and its own project tag.
func (p Project) EffectiveTagIds() []uuid.UUID {
	if p.TagId == uuid.Nil {
		return p.TagIds
	}
	return append(slices.Clone(p.TagIds), p.TagId)
}
