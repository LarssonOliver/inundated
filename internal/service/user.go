package service

import (
	"context"
	"errors"

	"github.com/larssonoliver/inundated/internal/model"
)

// GetCurrentUser implements [Service].
func (s *ServiceImpl) GetCurrentUser(ctx context.Context) (model.User, error) {
	user, ok := ctx.Value(model.UserContextKey).(model.User)
	if !ok {
		return model.User{}, model.ErrNotFound
	}
	return user, nil
}

// GetOrCreateUserBySub implements [Service].
func (s *ServiceImpl) GetOrCreateUserBySub(ctx context.Context, subject string) (model.User, error) {
	user, err := s.repository.GetUserBySub(ctx, subject)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, model.ErrNotFound) {
		return model.User{}, err
	}

	return s.repository.CreateUser(ctx, model.User{Sub: subject})
}

// UpdateCurrentUser implements [Service].
func (s *ServiceImpl) UpdateCurrentUser(ctx context.Context, user model.User) (model.User, error) {
	current, err := s.GetCurrentUser(ctx)
	if err != nil {
		return model.User{}, err
	}
	user.Id = current.Id
	return s.repository.UpdateUser(ctx, user)
}
