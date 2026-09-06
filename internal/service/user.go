package service

import (
	"context"
	"errors"
	"log"

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

// GetUserBySub implements [UserService].
func (s *ServiceImpl) GetUserBySub(ctx context.Context, sub string) (model.User, error) {
	return s.repository.GetUserBySub(ctx, sub)
}

// GetOrCreateUserByIdentity reconciles the local user record with the OIDC
// claims: unknown subjects are created, and a drifted email or name is updated.
func (s *ServiceImpl) GetOrCreateUserByIdentity(ctx context.Context, identity model.UserIdentity) (model.User, error) {
	user, err := s.repository.GetUserBySub(ctx, identity.Sub)
	if err != nil {
		if !errors.Is(err, model.ErrNotFound) {
			return model.User{}, err
		}
		return s.createUserFromIdentity(ctx, identity)
	}

	if user.Email == identity.Email && user.Name == identity.Name {
		return user, nil
	}

	user.Email = identity.Email
	user.Name = identity.Name
	return s.repository.UpdateUser(ctx, user)
}

func (s *ServiceImpl) createUserFromIdentity(ctx context.Context, identity model.UserIdentity) (model.User, error) {
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

// UpdateCurrentUser implements [Service].
func (s *ServiceImpl) UpdateCurrentUser(ctx context.Context, user model.User) (model.User, error) {
	current, err := s.GetCurrentUser(ctx)
	if err != nil {
		return model.User{}, err
	}
	user.Id = current.Id
	return s.repository.UpdateUser(ctx, user)
}
