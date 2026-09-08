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

func memSession(sub, token string) (model.Session, string) {
	return model.Session{
		Id:        uuid.New(),
		UserId:    uuid.New(),
		Sub:       sub,
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().Add(time.Hour).UTC(),
	}, token
}

func TestMemoryStore_CreateSession(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		store := memory.NewMemoryStore()
		session, token := memSession("auth0|user123", "tok-1")

		got, err := store.CreateSession(ctx, session, token)
		require.NoError(t, err)
		require.Equal(t, session, got)
	})

	t.Run("EmptyToken", func(t *testing.T) {
		store := memory.NewMemoryStore()
		session, token := memSession("auth0|notoken", "")

		_, err := store.CreateSession(ctx, session, token)
		require.ErrorIs(t, err, model.ErrInvalidArgument)
	})

	t.Run("DuplicateID", func(t *testing.T) {
		store := memory.NewMemoryStore()
		session, token := memSession("auth0|dup", "tok-dup-a")

		_, err := store.CreateSession(ctx, session, token)
		require.NoError(t, err)

		_, err = store.CreateSession(ctx, session, "tok-dup-b")
		require.ErrorIs(t, err, model.ErrAlreadyExists)
	})

	t.Run("NilID", func(t *testing.T) {
		store := memory.NewMemoryStore()
		session, token := memSession("auth0|nilid", "tok-nilid")
		session.Id = uuid.Nil

		got, err := store.CreateSession(ctx, session, token)
		require.NoError(t, err)
		require.NotEqual(t, got.Id, uuid.Nil)
	})
}

func TestMemoryStore_GetSessionByToken(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		store := memory.NewMemoryStore()
		session, token := memSession("auth0|user123", "tok-get")
		_, err := store.CreateSession(ctx, session, token)
		require.NoError(t, err)

		got, err := store.GetSessionByToken(ctx, token)
		require.NoError(t, err)
		require.Equal(t, session.Id, got.Id)
		require.Equal(t, session.Sub, got.Sub)
	})

	t.Run("TheIDIsNotAToken", func(t *testing.T) {
		store := memory.NewMemoryStore()
		session, token := memSession("auth0|user123", "tok-x")
		_, err := store.CreateSession(ctx, session, token)
		require.NoError(t, err)

		_, err = store.GetSessionByToken(ctx, session.Id.String())
		require.ErrorIs(t, err, model.ErrNotFound)
	})

	t.Run("NotFound", func(t *testing.T) {
		store := memory.NewMemoryStore()

		_, err := store.GetSessionByToken(ctx, "no-such-token")
		require.ErrorIs(t, err, model.ErrNotFound)
	})
}

func TestMemoryStore_TouchSession(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		store := memory.NewMemoryStore()
		session, token := memSession("auth0|updatetest", "tok-touch")
		_, err := store.CreateSession(ctx, session, token)
		require.NoError(t, err)

		newExpiresAt := time.Now().Add(2 * time.Hour).UTC()

		got, err := store.TouchSession(ctx, session.Id, newExpiresAt)
		require.NoError(t, err)
		require.Equal(t, newExpiresAt, got.ExpiresAt)
		require.Equal(t, session.Id, got.Id)

		got, err = store.GetSessionByToken(ctx, token)
		require.NoError(t, err)
		require.Equal(t, newExpiresAt, got.ExpiresAt)
	})

	t.Run("NotFound", func(t *testing.T) {
		store := memory.NewMemoryStore()

		_, err := store.TouchSession(ctx, uuid.New(), time.Now().Add(2*time.Hour).UTC())
		require.ErrorIs(t, err, model.ErrNotFound)
	})

	t.Run("DoesNotAffectOtherSessions", func(t *testing.T) {
		store := memory.NewMemoryStore()
		sessionA, tokenA := memSession("auth0|a", "tok-a")
		sessionB, tokenB := memSession("auth0|b", "tok-b")
		_, err := store.CreateSession(ctx, sessionA, tokenA)
		require.NoError(t, err)
		_, err = store.CreateSession(ctx, sessionB, tokenB)
		require.NoError(t, err)

		_, err = store.TouchSession(ctx, sessionA.Id, time.Now().Add(2*time.Hour).UTC())
		require.NoError(t, err)

		got, err := store.GetSessionByToken(ctx, tokenB)
		require.NoError(t, err)
		require.WithinDuration(t, sessionB.ExpiresAt, got.ExpiresAt, time.Second)
	})
}

func TestMemoryStore_DeleteSession(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		store := memory.NewMemoryStore()
		session, token := memSession("auth0|deletetest", "tok-del")
		_, err := store.CreateSession(ctx, session, token)
		require.NoError(t, err)

		err = store.DeleteSession(ctx, session.Id)
		require.NoError(t, err)

		_, err = store.GetSessionByToken(ctx, token)
		require.ErrorIs(t, err, model.ErrNotFound)
	})

	t.Run("NotFound", func(t *testing.T) {
		store := memory.NewMemoryStore()

		err := store.DeleteSession(ctx, uuid.New())
		require.ErrorIs(t, err, model.ErrNotFound)
	})

	t.Run("DoesNotAffectOtherSessions", func(t *testing.T) {
		store := memory.NewMemoryStore()
		sessionA, tokenA := memSession("auth0|a", "tok-a2")
		sessionB, tokenB := memSession("auth0|b", "tok-b2")
		_, err := store.CreateSession(ctx, sessionA, tokenA)
		require.NoError(t, err)
		_, err = store.CreateSession(ctx, sessionB, tokenB)
		require.NoError(t, err)

		err = store.DeleteSession(ctx, sessionA.Id)
		require.NoError(t, err)

		got, err := store.GetSessionByToken(ctx, tokenB)
		require.NoError(t, err)
		require.Equal(t, sessionB.Id, got.Id)
	})

	t.Run("DeleteAllExpiredSessions", func(t *testing.T) {
		store := memory.NewMemoryStore()

		expired, expiredToken := memSession("auth0|expired", "tok-expired")
		expired.ExpiresAt = time.Now().Add(-1 * time.Hour).UTC()
		live, liveToken := memSession("auth0|live", "tok-live")

		_, err := store.CreateSession(ctx, expired, expiredToken)
		require.NoError(t, err)
		_, err = store.CreateSession(ctx, live, liveToken)
		require.NoError(t, err)

		err = store.DeleteAllExpiredSessions(ctx)
		require.NoError(t, err)

		_, err = store.GetSessionByToken(ctx, expiredToken)
		require.ErrorIs(t, err, model.ErrNotFound)
		_, err = store.GetSessionByToken(ctx, liveToken)
		require.NoError(t, err)
	})
}
