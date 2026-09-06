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

// Regression: a first login used to fail because the service created the user
// with only a sub, which the repository rejects for having no email.
func TestUserService_GetOrCreateUserByIdentity_MemoryBacked(t *testing.T) {
	ctx := context.Background()
	store := memory.NewMemoryStore()
	svc := service.NewService(store)

	start := time.Now().UTC()
	_, err := store.CreateTimespan(ctx, model.Timespan{Name: "old", StartTime: start, EndTime: start.Add(time.Hour)})
	require.NoError(t, err)

	identity := model.UserIdentity{Sub: "oidc|abc", Email: "abc@example.com", Name: "Abc"}

	created, err := svc.GetOrCreateUserByIdentity(ctx, identity)
	require.NoError(t, err)
	require.NotEqual(t, "", created.Id.String())
	require.Equal(t, identity.Email, created.Email)
	require.Equal(t, identity.Name, created.Name)

	persisted, err := store.GetUserBySub(ctx, identity.Sub)
	require.NoError(t, err)
	require.Equal(t, created.Id, persisted.Id)

	page, err := store.ListTimespans(ctx, model.PaginationParams{Limit: 10})
	require.NoError(t, err)
	require.Len(t, page.Data, 1)
	require.NotNil(t, page.Data[0].UserId)
	require.Equal(t, created.Id, *page.Data[0].UserId)

	// Re-login with a drifted email updates the record in place.
	updated, err := svc.GetOrCreateUserByIdentity(ctx, model.UserIdentity{Sub: identity.Sub, Email: "new@example.com", Name: "Abc"})
	require.NoError(t, err)
	require.Equal(t, created.Id, updated.Id)
	require.Equal(t, "new@example.com", updated.Email)

	persisted, err = store.GetUserBySub(ctx, identity.Sub)
	require.NoError(t, err)
	require.Equal(t, "new@example.com", persisted.Email)
}
