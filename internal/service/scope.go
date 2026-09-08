package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

// ErrAmbiguousOwnerScope means the request context carries a user whose id is
// the zero UUID. That is never a legitimate state (a real user always has an
// id), so rather than fall back to the unowned pool -- which would hand an
// authenticated-but-malformed request the shared userless data -- scope
// resolution fails closed and the caller returns a server error.
var ErrAmbiguousOwnerScope = errors.New("owner scope: user in context has no id")

// ownerScope derives the resource ownership scope for the current request: the
// authenticated user, or the unowned pool when the server runs without auth.
//
// This is the one place that decides whose data a request may touch, so it
// fails closed: "no user in context" is the legitimate userless scope, but
// "a user in context with no id" is a bug and returns ErrAmbiguousOwnerScope
// instead of silently widening access.
func ownerScope(ctx context.Context) (model.OwnerScope, error) {
	u, ok := model.GetCurrentUserFromContext(ctx)
	if !ok {
		return model.UnownedScope(), nil
	}
	if u.Id == uuid.Nil {
		return model.OwnerScope{}, ErrAmbiguousOwnerScope
	}
	return model.UserScope(u.Id), nil
}
