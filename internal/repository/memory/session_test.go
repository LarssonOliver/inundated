package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository/memory"
	"github.com/stretchr/testify/require"
)

func TestMemoryStore_CreateSession(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		store := memory.NewMemoryStore()
		session := model.Session{
			Id:        uuid.New(),
			UserId:    uuid.New(),
			Sub:       "auth0|user123",
			ExpiresAt: time.Now().Add(time.Hour).UTC(),
		}

		got, err := store.CreateSession(ctx, session)
		require.NoError(t, err)
		require.Equal(t, session, got)
	})

	t.Run("DuplicateID", func(t *testing.T) {
		store := memory.NewMemoryStore()
		session := model.Session{
			Id:        uuid.New(),
			UserId:    uuid.New(),
			Sub:       "auth0|dup",
			ExpiresAt: time.Now().Add(time.Hour).UTC(),
		}

		_, err := store.CreateSession(ctx, session)
		require.NoError(t, err)

		_, err = store.CreateSession(ctx, session)
		require.ErrorIs(t, err, model.ErrAlreadyExists)
	})

	t.Run("NilID", func(t *testing.T) {
		store := memory.NewMemoryStore()
		session := model.Session{
			Id:        uuid.Nil,
			UserId:    uuid.New(),
			Sub:       "auth0|dup",
			ExpiresAt: time.Now().Add(time.Hour).UTC(),
		}

		got, err := store.CreateSession(ctx, session)
		require.NoError(t, err)
		require.NotEqual(t, got.Id, uuid.Nil)
	})
}

func TestMemoryStore_GetSession(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		store := memory.NewMemoryStore()
		session := model.Session{
			Id:        uuid.New(),
			UserId:    uuid.New(),
			Sub:       "auth0|user123",
			ExpiresAt: time.Now().Add(time.Hour).UTC(),
		}
		_, err := store.CreateSession(ctx, session)
		require.NoError(t, err)

		got, err := store.GetSession(ctx, session.Id)
		require.NoError(t, err)
		require.Equal(t, session, got)
	})

	t.Run("NotFound", func(t *testing.T) {
		store := memory.NewMemoryStore()

		_, err := store.GetSession(ctx, uuid.New())
		require.ErrorIs(t, err, model.ErrNotFound)
	})
}

func TestMemoryStore_TouchSession(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		store := memory.NewMemoryStore()
		session := model.Session{
			Id:        uuid.New(),
			UserId:    uuid.New(),
			Sub:       "auth0|updatetest",
			ExpiresAt: time.Now().Add(time.Hour).UTC(),
		}
		_, err := store.CreateSession(ctx, session)
		require.NoError(t, err)

		newExpiresAt := time.Now().Add(2 * time.Hour).UTC()

		got, err := store.TouchSession(ctx, session.Id, newExpiresAt)
		require.NoError(t, err)
		require.Equal(t, newExpiresAt, got.ExpiresAt)
		require.Equal(t, session.Id, got.Id)
		require.Equal(t, session.UserId, got.UserId)
		require.Equal(t, session.Sub, got.Sub)

		got, err = store.GetSession(ctx, session.Id)
		require.NoError(t, err)
		require.Equal(t, newExpiresAt, got.ExpiresAt)
		require.Equal(t, session.Id, got.Id)
		require.Equal(t, session.UserId, got.UserId)
		require.Equal(t, session.Sub, got.Sub)
	})

	t.Run("NotFound", func(t *testing.T) {
		store := memory.NewMemoryStore()
		session := model.Session{
			Id:        uuid.New(),
			UserId:    uuid.New(),
			Sub:       "auth0|ghost",
			ExpiresAt: time.Now().Add(time.Hour).UTC(),
		}

		_, err := store.TouchSession(ctx, session.Id, time.Now().Add(2*time.Hour).UTC())
		require.ErrorIs(t, err, model.ErrNotFound)
	})

	t.Run("DoesNotAffectOtherSessions", func(t *testing.T) {
		store := memory.NewMemoryStore()
		sessionA := model.Session{
			Id:        uuid.New(),
			UserId:    uuid.New(),
			Sub:       "auth0|a",
			ExpiresAt: time.Now().Add(time.Hour).UTC(),
		}
		sessionB := model.Session{
			Id:        uuid.New(),
			UserId:    uuid.New(),
			Sub:       "auth0|b",
			ExpiresAt: time.Now().Add(time.Hour).UTC(),
		}
		_, err := store.CreateSession(ctx, sessionA)
		require.NoError(t, err)
		_, err = store.CreateSession(ctx, sessionB)
		require.NoError(t, err)

		updatedA := sessionA
		updatedA.ExpiresAt = time.Now().Add(2 * time.Hour).UTC()
		_, err = store.TouchSession(ctx, updatedA.Id, updatedA.ExpiresAt)
		require.NoError(t, err)

		got, err := store.GetSession(ctx, sessionB.Id)
		require.NoError(t, err)
		require.Equal(t, sessionB, got)
	})
}

func TestMemoryStore_DeleteSession(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		store := memory.NewMemoryStore()
		session := model.Session{
			Id:        uuid.New(),
			UserId:    uuid.New(),
			Sub:       "auth0|deletetest",
			ExpiresAt: time.Now().Add(time.Hour).UTC(),
		}
		_, err := store.CreateSession(ctx, session)
		require.NoError(t, err)

		err = store.DeleteSession(ctx, session.Id)
		require.NoError(t, err)

		_, err = store.GetSession(ctx, session.Id)
		require.ErrorIs(t, err, model.ErrNotFound)
	})

	t.Run("NotFound", func(t *testing.T) {
		store := memory.NewMemoryStore()

		err := store.DeleteSession(ctx, uuid.New())
		require.ErrorIs(t, err, model.ErrNotFound)
	})

	t.Run("DoesNotAffectOtherSessions", func(t *testing.T) {
		store := memory.NewMemoryStore()
		sessionA := model.Session{
			Id:        uuid.New(),
			UserId:    uuid.New(),
			Sub:       "auth0|a",
			ExpiresAt: time.Now().Add(time.Hour).UTC(),
		}
		sessionB := model.Session{
			Id:        uuid.New(),
			UserId:    uuid.New(),
			Sub:       "auth0|b",
			ExpiresAt: time.Now().Add(time.Hour).UTC(),
		}
		_, err := store.CreateSession(ctx, sessionA)
		require.NoError(t, err)
		_, err = store.CreateSession(ctx, sessionB)
		require.NoError(t, err)

		err = store.DeleteSession(ctx, sessionA.Id)
		require.NoError(t, err)

		got, err := store.GetSession(ctx, sessionB.Id)
		require.NoError(t, err)
		require.Equal(t, sessionB, got)
	})

	t.Run("DeleteAllExpiredSessions", func(t *testing.T) {
		store := memory.NewMemoryStore()

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

		_, err := store.CreateSession(ctx, sessions[0])
		require.NoError(t, err)
		_, err = store.CreateSession(ctx, sessions[1])
		require.NoError(t, err)

		err = store.DeleteAllExpiredSessions(ctx)
		require.NoError(t, err)

		_, err = store.GetSession(ctx, sessions[0].Id)
		require.ErrorIs(t, err, model.ErrNotFound)
		_, err = store.GetSession(ctx, sessions[1].Id)
		require.NoError(t, err)
	})
}
