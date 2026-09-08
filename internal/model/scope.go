package model

import "github.com/google/uuid"

type OwnerScope struct {
	userID *uuid.UUID
}

func UserScope(id uuid.UUID) OwnerScope {
	return OwnerScope{userID: &id}
}

func UnownedScope() OwnerScope {
	return OwnerScope{}
}

func (s OwnerScope) UserID() *uuid.UUID {
	if s.userID == nil {
		return nil
	}
	id := *s.userID
	return &id
}
