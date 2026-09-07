package service

import (
	"context"
	"fmt"

	"github.com/larssonoliver/inundated/internal/repository"
)

// EnsureAuthConfigConsistent guards against silently dropping authentication
// after users exist. Userless mode is only permitted while the users table is
// empty; once someone has logged in via OIDC the server must keep OIDC
// configured or it would expose every user's data through the unowned scope.
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
