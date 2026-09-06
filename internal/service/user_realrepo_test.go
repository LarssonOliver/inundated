package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository/memory"
	"github.com/larssonoliver/inundated/internal/service"
	"github.com/stretchr/testify/require"
)

// Regression: a brand new user logging in used to fail because the service
// created the user with only a sub, and the repository rejects an empty email.
// GetOrCreateUserByIdentity must persist the full identity.
func TestUserService_GetOrCreateUserByIdentity_MemoryBacked(t *testing.T) {
	ctx := context.Background()
	store := memory.NewMemoryStore()
	svc := service.NewService(store)

	// Seed a resource that predates user support.
	start := time.Now().UTC()
	_, err := store.CreateTimespan(ctx, model.Timespan{Name: "old", StartTime: start, EndTime: start.Add(time.Hour)})
	require.NoError(t, err)

	identity := model.UserIdentity{Sub: "oidc|abc", Email: "abc@example.com", Name: "Abc"}

	created, err := svc.GetOrCreateUserByIdentity(ctx, identity)
	require.NoError(t, err)
	require.NotEqual(t, "", created.Id.String())
	require.Equal(t, identity.Email, created.Email)
	require.Equal(t, identity.Name, created.Name)

	// The user is persisted and the first user adopted the orphan timespan.
	persisted, err := store.GetUserBySub(ctx, identity.Sub)
	require.NoError(t, err)
	require.Equal(t, created.Id, persisted.Id)

	page, err := store.ListTimespans(ctx, model.PaginationParams{Limit: 10})
	require.NoError(t, err)
	require.Len(t, page.Data, 1)
	require.NotNil(t, page.Data[0].UserId)
	require.Equal(t, created.Id, *page.Data[0].UserId)

	// Logging in again with a drifted email updates the record in place.
	updated, err := svc.GetOrCreateUserByIdentity(ctx, model.UserIdentity{Sub: identity.Sub, Email: "new@example.com", Name: "Abc"})
	require.NoError(t, err)
	require.Equal(t, created.Id, updated.Id)
	require.Equal(t, "new@example.com", updated.Email)

	persisted, err = store.GetUserBySub(ctx, identity.Sub)
	require.NoError(t, err)
	require.Equal(t, "new@example.com", persisted.Email)
}
