package contract_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/repository/memory"
	"github.com/larssonoliver/inundated/internal/repository/postgres"
	"github.com/larssonoliver/inundated/test/testutils"
	"github.com/stretchr/testify/require"
)

func TestLoginStateRepositoryContract(t *testing.T) {
	ctx := context.Background()

	run := func(t *testing.T, repoName string, newRepo func(t *testing.T) repository.LoginStateRepository) {
		t.Run(repoName+"CreateAndGetByID", func(t *testing.T) {
			repo := newRepo(t)
			loginState := model.LoginState{
				Id:           uuid.New(),
				RedirectUri:  "https://example.com/callback",
				CodeVerifier: "some-code",
				ExpiresAt:    time.Now().Add(time.Hour).UTC(),
			}

			got, err := repo.CreateLoginState(ctx, loginState)
			require.NoError(t, err)
			require.Equal(t, loginState.Id, got.Id)
			require.Equal(t, loginState.RedirectUri, got.RedirectUri)
			require.Equal(t, loginState.CodeVerifier, got.CodeVerifier)
			require.WithinDuration(t, loginState.ExpiresAt, got.ExpiresAt, time.Second)

			got, err = repo.GetLoginState(ctx, loginState.Id)
			require.NoError(t, err)
			require.Equal(t, loginState.Id, got.Id)
			require.Equal(t, loginState.RedirectUri, got.RedirectUri)
			require.Equal(t, loginState.CodeVerifier, got.CodeVerifier)
			require.WithinDuration(t, loginState.ExpiresAt, got.ExpiresAt, time.Second)
		})

		t.Run(repoName+"GetByIDMissing", func(t *testing.T) {
			repo := newRepo(t)

			_, err := repo.GetLoginState(ctx, uuid.New())
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"CreateDuplicateID", func(t *testing.T) {
			repo := newRepo(t)
			loginState := model.LoginState{
				Id:           uuid.New(),
				RedirectUri:  "https://example.com/dup",
				CodeVerifier: "some-code",
				ExpiresAt:    time.Now().Add(time.Hour).UTC(),
			}

			_, err := repo.CreateLoginState(ctx, loginState)
			require.NoError(t, err)

			_, err = repo.CreateLoginState(ctx, loginState)
			require.ErrorIs(t, err, model.ErrAlreadyExists)
		})

		t.Run(repoName+"CreateNilID", func(t *testing.T) {
			repo := newRepo(t)
			loginState := model.LoginState{
				Id:           uuid.Nil,
				RedirectUri:  "https://example.com/nilid",
				CodeVerifier: "some-code",
				ExpiresAt:    time.Now().Add(time.Hour).UTC(),
			}

			got, err := repo.CreateLoginState(ctx, loginState)
			require.NoError(t, err)
			require.NotEqual(t, uuid.Nil, got.Id)
		})

		t.Run(repoName+"Delete", func(t *testing.T) {
			repo := newRepo(t)
			loginState := model.LoginState{
				Id:           uuid.New(),
				RedirectUri:  "https://example.com/deletetest",
				CodeVerifier: "some-code",
				ExpiresAt:    time.Now().Add(time.Hour).UTC(),
			}

			_, err := repo.CreateLoginState(ctx, loginState)
			require.NoError(t, err)

			err = repo.DeleteLoginState(ctx, loginState.Id)
			require.NoError(t, err)

			_, err = repo.GetLoginState(ctx, loginState.Id)
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"DeleteMissing", func(t *testing.T) {
			repo := newRepo(t)

			err := repo.DeleteLoginState(ctx, uuid.New())
			require.ErrorIs(t, err, model.ErrNotFound)
		})
	}

	// Memory
	run(t, "memory", func(t *testing.T) repository.LoginStateRepository {
		return memory.NewMemoryStore()
	})

	// Postgres
	run(t, "postgres", func(t *testing.T) repository.LoginStateRepository {
		t.Parallel()
		pool := testutils.StartPostgresContainerWithMigrationsApplied(ctx, t)
		return postgres.NewPostgresStoreFromPool(pool)
	})

	// If/when a valkey-backed implementation exists, add it here following
	// the same pattern, e.g.:
	//
	// run(t, "valkey", func(t *testing.T) repository.LoginStateRepository {
	// 	t.Parallel()
	// 	client := testutils.StartValkeyContainer(ctx, t)
	// 	return valkey.NewValkeyLoginStateStore(client)
	// })
}
