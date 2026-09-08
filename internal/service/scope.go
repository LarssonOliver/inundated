package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

func ownerScope(ctx context.Context) (model.OwnerScope, error) {
	u, ok := model.GetCurrentUserFromContext(ctx)
	if !ok {
		return model.UnownedScope(), nil
	}
	if u.Id == uuid.Nil {
		return model.OwnerScope{}, model.ErrAmbiguousOwnerScope
	}
	return model.UserScope(u.Id), nil
}
