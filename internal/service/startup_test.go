package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/service"
	"github.com/stretchr/testify/require"
)

func TestEnsureAuthConfigConsistent(t *testing.T) {
	ctx := context.Background()

	t.Run("oidc enabled is always allowed", func(t *testing.T) {
		repo := &repository.RepoMock{
			HasUsersFn: func(ctx context.Context) (bool, error) { return true, nil },
		}
		require.NoError(t, service.EnsureAuthConfigConsistent(ctx, repo, true))
	})

	t.Run("userless mode allowed when no users exist", func(t *testing.T) {
		repo := &repository.RepoMock{
			HasUsersFn: func(ctx context.Context) (bool, error) { return false, nil },
		}
		require.NoError(t, service.EnsureAuthConfigConsistent(ctx, repo, false))
	})

	t.Run("userless mode rejected once a user exists", func(t *testing.T) {
		repo := &repository.RepoMock{
			HasUsersFn: func(ctx context.Context) (bool, error) { return true, nil },
		}
		err := service.EnsureAuthConfigConsistent(ctx, repo, false)
		require.Error(t, err)
	})

	t.Run("propagates repository errors", func(t *testing.T) {
		sentinel := errors.New("boom")
		repo := &repository.RepoMock{
			HasUsersFn: func(ctx context.Context) (bool, error) { return false, sentinel },
		}
		err := service.EnsureAuthConfigConsistent(ctx, repo, false)
		require.ErrorIs(t, err, sentinel)
	})
}
