package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

// ownerScope derives the resource ownership scope for the current request:
// the authenticated user, or the unowned pool when the server runs without auth.
func ownerScope(ctx context.Context) model.OwnerScope {
	if u, ok := model.GetCurrentUserFromContext(ctx); ok && u.Id != uuid.Nil {
		return model.UserScope(u.Id)
	}
	return model.UnownedScope()
}
