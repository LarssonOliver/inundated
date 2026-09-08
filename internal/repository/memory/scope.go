package memory

import (
	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

// matchesScope reports whether a row with the given owner belongs to scope.
func matchesScope(userID *uuid.UUID, scope model.OwnerScope) bool {
	want := scope.UserID()
	if want == nil {
		return userID == nil
	}
	return userID != nil && *userID == *want
}
