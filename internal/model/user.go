package model

import (
	"github.com/google/uuid"
)

type User struct {
	Id    uuid.UUID
	Sub   string
	Email string
	Name  string
}

type UserIdentity struct {
	Sub   string
	Email string
	Name  string
}

type OrphanAdoption struct {
	Projects  int
	Tags      int
	Timespans int
	Tasks     int
	Settings  int
	// An orphan app password may be held by anyone who could reach the
	// instance while it had no login, so they are deleted, not adopted.
	DeletedAppPasswordCount int
}

func (a OrphanAdoption) Total() int {
	return a.Projects + a.Tags + a.Timespans + a.Tasks + a.Settings + a.DeletedAppPasswordCount
}
