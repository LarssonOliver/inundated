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

// GetOrCreateUserBySub implements [Service].
func (s *ServiceImpl) GetOrCreateUserBySub(ctx context.Context, subject string) (model.User, error) {
	user, err := s.repository.GetUserBySub(ctx, subject)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, model.ErrNotFound) {
		return model.User{}, err
	}

	// A brand new user. When it is the very first user in the system, it also
	// takes ownership of every resource that predates user support.
	created, adoption, err := s.repository.CreateUserAdoptingOrphans(ctx, model.User{Sub: subject})
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
