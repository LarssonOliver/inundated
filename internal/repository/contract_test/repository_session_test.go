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

func TestSessionRepositoryContract(t *testing.T) {
	ctx := context.Background()

	run := func(t *testing.T, repoName string, newRepo func(t *testing.T) repository.SessionRepository) {
		t.Run(repoName+"CreateAndGetByID", func(t *testing.T) {
			repo := newRepo(t)
			session := model.Session{
				Id:        uuid.New(),
				UserId:    uuid.New(),
				Sub:       "auth0|user123",
				ExpiresAt: time.Now().Add(time.Hour).UTC(),
			}

			got, err := repo.CreateSession(ctx, session)
			require.NoError(t, err)
			require.Equal(t, session.Id, got.Id)
			require.Equal(t, session.UserId, got.UserId)
			require.Equal(t, session.Sub, got.Sub)
			require.WithinDuration(t, session.ExpiresAt, got.ExpiresAt, time.Second)

			got, err = repo.GetSession(ctx, session.Id)
			require.NoError(t, err)
			require.Equal(t, session.Id, got.Id)
			require.Equal(t, session.UserId, got.UserId)
			require.Equal(t, session.Sub, got.Sub)
			require.WithinDuration(t, session.ExpiresAt, got.ExpiresAt, time.Second)
		})

		t.Run(repoName+"GetByIDMissing", func(t *testing.T) {
			repo := newRepo(t)

			_, err := repo.GetSession(ctx, uuid.New())
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"CreateDuplicateID", func(t *testing.T) {
			repo := newRepo(t)
			session := model.Session{
				Id:        uuid.New(),
				UserId:    uuid.New(),
				Sub:       "auth0|dup",
				ExpiresAt: time.Now().Add(time.Hour).UTC(),
			}

			_, err := repo.CreateSession(ctx, session)
			require.NoError(t, err)

			_, err = repo.CreateSession(ctx, session)
			require.ErrorIs(t, err, model.ErrAlreadyExists)
		})

		t.Run(repoName+"CreateNilID", func(t *testing.T) {
			repo := newRepo(t)
			session := model.Session{
				Id:        uuid.Nil,
				UserId:    uuid.New(),
				Sub:       "auth0|nilid",
				ExpiresAt: time.Now().Add(time.Hour).UTC(),
			}

			got, err := repo.CreateSession(ctx, session)
			require.NoError(t, err)
			require.NotEqual(t, uuid.Nil, got.Id)
		})

		t.Run(repoName+"Touch", func(t *testing.T) {
			repo := newRepo(t)
			session := model.Session{
				Id:        uuid.New(),
				UserId:    uuid.New(),
				Sub:       "auth0|updatetest",
				ExpiresAt: time.Now().Add(time.Hour).UTC(),
			}

			_, err := repo.CreateSession(ctx, session)
			require.NoError(t, err)

			newExpiresAt := time.Now().Add(2 * time.Hour).UTC()
			updated, err := repo.TouchSession(ctx, session.Id, newExpiresAt)
			require.NoError(t, err)
			require.WithinDuration(t, newExpiresAt, updated.ExpiresAt, time.Second)

			got, err := repo.GetSession(ctx, session.Id)
			require.NoError(t, err)
			require.WithinDuration(t, newExpiresAt, got.ExpiresAt, time.Second)
		})

		t.Run(repoName+"TouchMissing", func(t *testing.T) {
			repo := newRepo(t)

			_, err := repo.TouchSession(ctx, uuid.New(), time.Now().Add(time.Hour).UTC())
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"Delete", func(t *testing.T) {
			repo := newRepo(t)
			session := model.Session{
				Id:        uuid.New(),
				UserId:    uuid.New(),
				Sub:       "auth0|deletetest",
				ExpiresAt: time.Now().Add(time.Hour).UTC(),
			}

			_, err := repo.CreateSession(ctx, session)
			require.NoError(t, err)

			err = repo.DeleteSession(ctx, session.Id)
			require.NoError(t, err)

			_, err = repo.GetSession(ctx, session.Id)
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"DeleteMissing", func(t *testing.T) {
			repo := newRepo(t)

			err := repo.DeleteSession(ctx, uuid.New())
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"DeleteAllExpiredSessions", func(t *testing.T) {
			repo := newRepo(t)

			sessions := []model.Session{
				{
					Id:        uuid.New(),
					UserId:    uuid.New(),
					Sub:       "auth0|expired1",
					ExpiresAt: time.Now().Add(-1 * time.Hour).UTC(),
				},
				{
					Id:        uuid.New(),
					UserId:    uuid.New(),
					Sub:       "auth0|expired1",
					ExpiresAt: time.Now().Add(time.Hour).UTC(),
				},
			}

			_, err := repo.CreateSession(ctx, sessions[0])
			require.NoError(t, err)
			_, err = repo.CreateSession(ctx, sessions[1])
			require.NoError(t, err)

			err = repo.DeleteAllExpiredSessions(ctx)
			require.NoError(t, err)

			_, err = repo.GetSession(ctx, sessions[0].Id)
			require.ErrorIs(t, err, model.ErrNotFound)
			_, err = repo.GetSession(ctx, sessions[1].Id)
			require.NoError(t, err)
		})
	}

	// Memory
	run(t, "memory", func(t *testing.T) repository.SessionRepository {
		return memory.NewMemoryStore()
	})

	// Postgres
	run(t, "postgres", func(t *testing.T) repository.SessionRepository {
		t.Parallel()
		pool := testutils.StartPostgresContainerWithMigrationsApplied(ctx, t)
		return postgres.NewPostgresStoreFromPool(pool)
	})

	// If/when a valkey-backed implementation exists, add it here following
	// the same pattern, e.g.:
	//
	// run(t, "valkey", func(t *testing.T) repository.SessionRepository {
	// 	t.Parallel()
	// 	client := testutils.StartValkeyContainer(ctx, t)
	// 	return valkey.NewValkeySessionStore(client)
	// })
}
