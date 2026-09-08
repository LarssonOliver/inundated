package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/larssonoliver/inundated/internal/model"
)

// GetCurrentUser implements [Service].
func (s *ServiceImpl) GetCurrentUser(ctx context.Context) (model.User, error) {
	user, ok := model.GetCurrentUserFromContext(ctx)
	if !ok {
		return model.User{}, model.ErrNotFound
	}
	return user, nil
}

// GetUserBySub implements [UserService].
func (s *ServiceImpl) GetUserBySub(ctx context.Context, sub string) (model.User, error) {
	return s.repository.GetUserBySub(ctx, sub)
}

func (s *ServiceImpl) GetOrCreateUserByIdentity(ctx context.Context, identity model.UserIdentity) (model.User, error) {
	user, err := s.repository.GetUserBySub(ctx, identity.Sub)
	if err != nil {
		if !errors.Is(err, model.ErrNotFound) {
			return model.User{}, err
		}
		return s.createUserFromIdentity(ctx, identity)
	}

	changed := false
	if identity.Email != "" && identity.Email != user.Email {
		user.Email = identity.Email
		changed = true
	}
	if identity.Name != "" && identity.Name != user.Name {
		user.Name = identity.Name
		changed = true
	}
	if !changed {
		return user, nil
	}
	return s.repository.UpdateUser(ctx, user)
}

func (s *ServiceImpl) createUserFromIdentity(ctx context.Context, identity model.UserIdentity) (model.User, error) {
	if identity.Email == "" {
		return model.User{}, fmt.Errorf(
			"OIDC identity %q has no email claim; check that the provider returns email for the configured scopes: %w",
			identity.Sub, model.ErrInvalidArgument,
		)
	}

	created, adoption, err := s.repository.CreateUserAdoptingOrphans(ctx, model.User{
		Sub:   identity.Sub,
		Email: identity.Email,
		Name:  identity.Name,
	})
	if err != nil {
		return model.User{}, err
	}
	if adoption.Total() > 0 {
		log.Printf(
			"first user %s adopted %d orphaned resources (%d projects, %d tags, %d timespans)",
			created.Id, adoption.Total(), adoption.Projects, adoption.Tags, adoption.Timespans,
		)
	}
	return created, nil
}
