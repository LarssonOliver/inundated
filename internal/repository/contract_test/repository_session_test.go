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

	aSession := func(sub, token string) model.Session {
		return model.Session{
			Id:        uuid.New(),
			UserId:    uuid.New(),
			Sub:       sub,
			Token:     token,
			CreatedAt: time.Now().Add(-time.Minute).UTC(),
			ExpiresAt: time.Now().Add(time.Hour).UTC(),
		}
	}

	run := func(t *testing.T, repoName string, newRepo func(t *testing.T) repository.SessionRepository) {
		t.Run(repoName+"CreateAndGetByToken", func(t *testing.T) {
			repo := newRepo(t)
			session := aSession("auth0|user123", "tok-user123")

			got, err := repo.CreateSession(ctx, session)
			require.NoError(t, err)
			require.Equal(t, session.Id, got.Id)
			require.Equal(t, session.UserId, got.UserId)
			require.Equal(t, session.Sub, got.Sub)
			require.WithinDuration(t, session.CreatedAt, got.CreatedAt, time.Second)
			require.WithinDuration(t, session.ExpiresAt, got.ExpiresAt, time.Second)

			got, err = repo.GetSessionByToken(ctx, session.Token)
			require.NoError(t, err)
			require.Equal(t, session.Id, got.Id)
			require.Equal(t, session.UserId, got.UserId)
			require.Equal(t, session.Sub, got.Sub)
			require.Empty(t, got.Token, "a looked-up session must not carry the raw token back")
			require.WithinDuration(t, session.CreatedAt, got.CreatedAt, time.Second)
			require.WithinDuration(t, session.ExpiresAt, got.ExpiresAt, time.Second)
		})

		t.Run(repoName+"TokenIsNotTheID", func(t *testing.T) {
			repo := newRepo(t)
			session := aSession("auth0|opaque", "tok-opaque")
			_, err := repo.CreateSession(ctx, session)
			require.NoError(t, err)

			// The internal id is not a credential: presenting it must not
			// resolve the session.
			_, err = repo.GetSessionByToken(ctx, session.Id.String())
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"GetByTokenMissing", func(t *testing.T) {
			repo := newRepo(t)

			_, err := repo.GetSessionByToken(ctx, "nope-not-a-real-token")
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"CreateEmptyToken", func(t *testing.T) {
			repo := newRepo(t)
			session := aSession("auth0|notoken", "")

			_, err := repo.CreateSession(ctx, session)
			require.ErrorIs(t, err, model.ErrInvalidArgument)
		})

		t.Run(repoName+"CreateDuplicateID", func(t *testing.T) {
			repo := newRepo(t)
			session := aSession("auth0|dup", "tok-dup-a")

			_, err := repo.CreateSession(ctx, session)
			require.NoError(t, err)

			dup := session
			dup.Token = "tok-dup-b"
			_, err = repo.CreateSession(ctx, dup)
			require.ErrorIs(t, err, model.ErrAlreadyExists)
		})

		t.Run(repoName+"CreateNilID", func(t *testing.T) {
			repo := newRepo(t)
			session := aSession("auth0|nilid", "tok-nilid")
			session.Id = uuid.Nil

			got, err := repo.CreateSession(ctx, session)
			require.NoError(t, err)
			require.NotEqual(t, uuid.Nil, got.Id)
		})

		t.Run(repoName+"Touch", func(t *testing.T) {
			repo := newRepo(t)
			session := aSession("auth0|updatetest", "tok-touch")

			_, err := repo.CreateSession(ctx, session)
			require.NoError(t, err)

			newExpiresAt := time.Now().Add(2 * time.Hour).UTC()
			updated, err := repo.TouchSession(ctx, session.Id, newExpiresAt)
			require.NoError(t, err)
			require.WithinDuration(t, newExpiresAt, updated.ExpiresAt, time.Second)

			got, err := repo.GetSessionByToken(ctx, session.Token)
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
			session := aSession("auth0|deletetest", "tok-delete")

			_, err := repo.CreateSession(ctx, session)
			require.NoError(t, err)

			err = repo.DeleteSession(ctx, session.Id)
			require.NoError(t, err)

			_, err = repo.GetSessionByToken(ctx, session.Token)
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"DeleteMissing", func(t *testing.T) {
			repo := newRepo(t)

			err := repo.DeleteSession(ctx, uuid.New())
			require.ErrorIs(t, err, model.ErrNotFound)
		})

		t.Run(repoName+"DeleteAllExpiredSessions", func(t *testing.T) {
			repo := newRepo(t)

			// Several consecutive expired entries plus a live one: every
			// expired session must go, regardless of ordering.
			expired := make([]model.Session, 3)
			for i := range expired {
				expired[i] = aSession("auth0|expired", "tok-expired-"+uuid.NewString())
				expired[i].ExpiresAt = time.Now().Add(-1 * time.Hour).UTC()
				_, err := repo.CreateSession(ctx, expired[i])
				require.NoError(t, err)
			}
			live := aSession("auth0|live", "tok-live")
			_, err := repo.CreateSession(ctx, live)
			require.NoError(t, err)

			err = repo.DeleteAllExpiredSessions(ctx)
			require.NoError(t, err)

			for _, s := range expired {
				_, err = repo.GetSessionByToken(ctx, s.Token)
				require.ErrorIs(t, err, model.ErrNotFound)
			}
			_, err = repo.GetSessionByToken(ctx, live.Token)
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
}
