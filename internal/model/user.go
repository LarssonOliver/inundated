package model

import (
	"github.com/google/uuid"
)

type User struct {
	Id    uuid.UUID
	Sub   string // OIDC subject claim - unique per provider
	Email string
	Name  string // Non-nullable - empty string if not provided
}

// OrphanAdoption reports how many previously unowned resources were assigned to
// a newly created first user. It is the zero value for every subsequent user.
type OrphanAdoption struct {
	Projects  int
	Tags      int
	Timespans int
}

// Total is the number of resources adopted across all resource types.
func (a OrphanAdoption) Total() int {
	return a.Projects + a.Tags + a.Timespans
}
