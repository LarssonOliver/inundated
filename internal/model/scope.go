package model

import "github.com/google/uuid"

// OwnerScope identifies whose resources an operation may see or modify. The zero
// value is the unowned scope, used when the server runs without authentication
// (rows with user_id IS NULL).
type OwnerScope struct {
	userID *uuid.UUID
}

// UserScope scopes operations to a single authenticated user.
func UserScope(id uuid.UUID) OwnerScope {
	return OwnerScope{userID: &id}
}

// UnownedScope scopes operations to rows with no owner.
func UnownedScope() OwnerScope {
	return OwnerScope{}
}

// UserID returns the scoped user id, or nil for the unowned scope.
func (s OwnerScope) UserID() *uuid.UUID {
	return s.userID
}
