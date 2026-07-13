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

func TestMemoryStore_CreateLoginState(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		store := memory.NewMemoryStore()
		loginState := model.LoginState{
			Id:           uuid.New(),
			RedirectUri:  "https://example.com/callback",
			CodeVerifier: "some-code",
			ExpiresAt:    time.Now().Add(time.Hour).UTC(),
		}

		got, err := store.CreateLoginState(ctx, loginState)
		require.NoError(t, err)
		require.Equal(t, loginState, got)
	})

	t.Run("DuplicateID", func(t *testing.T) {
		store := memory.NewMemoryStore()
		loginState := model.LoginState{
			Id:           uuid.New(),
			RedirectUri:  "https://example.com/dup",
			CodeVerifier: "some-code",
			ExpiresAt:    time.Now().Add(time.Hour).UTC(),
		}

		_, err := store.CreateLoginState(ctx, loginState)
		require.NoError(t, err)

		_, err = store.CreateLoginState(ctx, loginState)
		require.ErrorIs(t, err, model.ErrAlreadyExists)
	})

	t.Run("NilID", func(t *testing.T) {
		store := memory.NewMemoryStore()
		loginState := model.LoginState{
			Id:           uuid.Nil,
			RedirectUri:  "https://example.com/nilid",
			CodeVerifier: "some-code",
			ExpiresAt:    time.Now().Add(time.Hour).UTC(),
		}

		got, err := store.CreateLoginState(ctx, loginState)
		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, got.Id)
	})
}

func TestMemoryStore_GetLoginState(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		store := memory.NewMemoryStore()
		loginState := model.LoginState{
			Id:           uuid.New(),
			RedirectUri:  "https://example.com/callback",
			CodeVerifier: "some-code",
			ExpiresAt:    time.Now().Add(time.Hour).UTC(),
		}

		_, err := store.CreateLoginState(ctx, loginState)
		require.NoError(t, err)

		got, err := store.GetLoginState(ctx, loginState.Id)
		require.NoError(t, err)
		require.Equal(t, loginState, got)
	})

	t.Run("NotFound", func(t *testing.T) {
		store := memory.NewMemoryStore()

		_, err := store.GetLoginState(ctx, uuid.New())
		require.ErrorIs(t, err, model.ErrNotFound)
	})
}

func TestMemoryStore_DeleteLoginState(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		store := memory.NewMemoryStore()
		loginState := model.LoginState{
			Id:           uuid.New(),
			RedirectUri:  "https://example.com/deletetest",
			CodeVerifier: "some-code",
			ExpiresAt:    time.Now().Add(time.Hour).UTC(),
		}

		_, err := store.CreateLoginState(ctx, loginState)
		require.NoError(t, err)

		err = store.DeleteLoginState(ctx, loginState.Id)
		require.NoError(t, err)

		_, err = store.GetLoginState(ctx, loginState.Id)
		require.ErrorIs(t, err, model.ErrNotFound)
	})

	t.Run("NotFound", func(t *testing.T) {
		store := memory.NewMemoryStore()

		err := store.DeleteLoginState(ctx, uuid.New())
		require.ErrorIs(t, err, model.ErrNotFound)
	})

	t.Run("DoesNotAffectOtherLoginStates", func(t *testing.T) {
		store := memory.NewMemoryStore()
		loginStateA := model.LoginState{
			Id:           uuid.New(),
			RedirectUri:  "https://example.com/a",
			CodeVerifier: "some-code",
			ExpiresAt:    time.Now().Add(time.Hour).UTC(),
		}
		loginStateB := model.LoginState{
			Id:           uuid.New(),
			RedirectUri:  "https://example.com/b",
			CodeVerifier: "some-code2",
			ExpiresAt:    time.Now().Add(time.Hour).UTC(),
		}

		_, err := store.CreateLoginState(ctx, loginStateA)
		require.NoError(t, err)
		_, err = store.CreateLoginState(ctx, loginStateB)
		require.NoError(t, err)

		err = store.DeleteLoginState(ctx, loginStateA.Id)
		require.NoError(t, err)

		got, err := store.GetLoginState(ctx, loginStateB.Id)
		require.NoError(t, err)
		require.Equal(t, loginStateB, got)
	})
}
