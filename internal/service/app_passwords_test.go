package service_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/repository/memory"
	"github.com/larssonoliver/inundated/internal/service"
	"github.com/stretchr/testify/require"
)

func TestCreateAppPassword_ReturnsUsableToken(t *testing.T) {
	svc := service.NewService(memory.NewMemoryStore())
	ctx := context.Background()

	created, token, err := svc.CreateAppPassword(ctx, "  Calendar app  ")
	require.NoError(t, err)
	require.Equal(t, "Calendar app", created.Name)

	authenticated, err := svc.AuthenticateAppPassword(ctx, token)
	require.NoError(t, err)
	require.Equal(t, created.Id, authenticated.Id)
}

func TestCreateAppPassword_InvalidName(t *testing.T) {
	svc := service.NewService(memory.NewMemoryStore())

	for _, name := range []string{"", "   ", strings.Repeat("a", 101)} {
		_, _, err := svc.CreateAppPassword(context.Background(), name)
		require.ErrorIs(t, err, model.ErrInvalidArgument, "name %q", name)
	}
}

func TestCreateAppPassword_UnownedInUserlessContext(t *testing.T) {
	svc := service.NewService(memory.NewMemoryStore())

	created, _, err := svc.CreateAppPassword(context.Background(), "name")
	require.NoError(t, err)
	require.Nil(t, created.UserId)
}

func TestAuthenticateAppPassword_Revoked(t *testing.T) {
	svc := service.NewService(memory.NewMemoryStore())
	ctx := context.Background()

	created, token, err := svc.CreateAppPassword(ctx, "name")
	require.NoError(t, err)
	require.NoError(t, svc.DeleteAppPassword(ctx, created.Id))

	_, err = svc.AuthenticateAppPassword(ctx, token)
	require.ErrorIs(t, err, model.ErrNotFound)
}

func TestAuthenticateAppPassword_TouchThrottled(t *testing.T) {
	tests := []struct {
		name        string
		lastUsedAge *time.Duration
		wantTouches int
	}{
		{"used a minute ago", durationPointer(time.Minute), 0},
		{"used six minutes ago", durationPointer(6 * time.Minute), 1},
		{"never used", nil, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			password := model.AppPassword{Id: uuid.New(), Name: "name"}
			if tt.lastUsedAge != nil {
				lastUsedAt := time.Now().Add(-*tt.lastUsedAge)
				password.LastUsedAt = &lastUsedAt
			}
			touches := 0
			repo := &repository.RepoMock{
				GetAppPasswordByTokenFn: func(ctx context.Context, token string) (model.AppPassword, error) {
					return password, nil
				},
				TouchAppPasswordFn: func(ctx context.Context, id uuid.UUID, usedAt time.Time) error {
					require.Equal(t, password.Id, id)
					touches++
					return nil
				},
			}

			got, err := service.NewService(repo).AuthenticateAppPassword(context.Background(), "token")
			require.NoError(t, err)
			require.Equal(t, password.Id, got.Id)
			require.Equal(t, tt.wantTouches, touches)
		})
	}
}

func durationPointer(d time.Duration) *time.Duration { return &d }
