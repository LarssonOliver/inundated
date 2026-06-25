package service

import (
	"context"

	"github.com/larssonoliver/inundated/internal/model"
)

// GetCurrentUser implements [Service].
func (s *ServiceImpl) GetCurrentUser(ctx context.Context) (model.User, error) {
	panic("unimplemented")
}

// GetOrCreateUserBySub implements [Service].
func (s *ServiceImpl) GetOrCreateUserBySub(ctx context.Context, subject string) (model.User, error) {
	panic("unimplemented")
}

// UpdateCurrentUser implements [Service].
func (s *ServiceImpl) UpdateCurrentUser(ctx context.Context, user model.User) (model.User, error) {
	panic("unimplemented")
}
