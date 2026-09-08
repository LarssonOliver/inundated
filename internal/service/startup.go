package service

import (
	"context"
	"fmt"

	"github.com/larssonoliver/inundated/internal/repository"
)

func EnsureAuthConfigConsistent(ctx context.Context, users repository.UserRepository, oidcEnabled bool) error {
	if oidcEnabled {
		return nil
	}

	hasUsers, err := users.HasUsers(ctx)
	if err != nil {
		return fmt.Errorf("checking for existing users: %w", err)
	}
	if hasUsers {
		return fmt.Errorf("OIDC is not configured but users already exist; refusing to start in userless mode (set oidc-issuer-url and its client credentials)")
	}
	return nil
}
