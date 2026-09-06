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

// UserIdentity holds the claims from the OIDC provider. Email is required.
type UserIdentity struct {
	Sub   string
	Email string
	Name  string
}

// OrphanAdoption counts the resources a first user claimed on creation.
type OrphanAdoption struct {
	Projects  int
	Tags      int
	Timespans int
}

func (a OrphanAdoption) Total() int {
	return a.Projects + a.Tags + a.Timespans
}
