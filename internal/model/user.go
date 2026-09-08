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
}

func (a OrphanAdoption) Total() int {
	return a.Projects + a.Tags + a.Timespans
}
