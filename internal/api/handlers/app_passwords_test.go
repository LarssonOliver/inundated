package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/api"
	"github.com/larssonoliver/inundated/internal/api/handlers"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAppPasswordHandler_ListAppPasswords(t *testing.T) {
	t.Run("never contains a password", func(t *testing.T) {
		lastUsed := time.Now()
		svc := &service.AppPasswordServiceMock{
			ListFn: func(ctx context.Context) ([]model.AppPassword, error) {
				return []model.AppPassword{
					{Id: uuid.New(), Name: "Phone", CreatedAt: time.Now(), LastUsedAt: &lastUsed},
					{Id: uuid.New(), Name: "Laptop", CreatedAt: time.Now()},
				}, nil
			},
		}

		h := handlers.NewAppPasswordHandler(svc)

		raw, err := h.ListAppPasswords(context.Background(), api.ListAppPasswordsRequestObject{})
		require.NoError(t, err)

		got := raw.(api.ListAppPasswords200JSONResponse)
		require.Len(t, got, 2)
		require.Equal(t, "Phone", got[0].Name)
		require.NotNil(t, got[0].LastUsedAt)
		require.Nil(t, got[1].LastUsedAt)

		body, err := json.Marshal(got)
		require.NoError(t, err)
		require.NotContains(t, string(body), "password")
	})

	t.Run("service error", func(t *testing.T) {
		svc := &service.AppPasswordServiceMock{
			ListFn: func(ctx context.Context) ([]model.AppPassword, error) {
				return nil, errors.New("service error")
			},
		}

		h := handlers.NewAppPasswordHandler(svc)

		_, err := h.ListAppPasswords(context.Background(), api.ListAppPasswordsRequestObject{})
		require.Error(t, err)
	})
}

func TestAppPasswordHandler_CreateAppPassword(t *testing.T) {
	t.Run("successful create returns the password", func(t *testing.T) {
		id := uuid.New()
		svc := &service.AppPasswordServiceMock{
			CreateFn: func(ctx context.Context, name string) (model.AppPassword, string, error) {
				require.Equal(t, "Phone", name)
				return model.AppPassword{Id: id, Name: name, CreatedAt: time.Now()}, "secret-token", nil
			},
		}

		h := handlers.NewAppPasswordHandler(svc)

		raw, err := h.CreateAppPassword(context.Background(), api.CreateAppPasswordRequestObject{
			Body: &api.CreateAppPassword{Name: "Phone"},
		})
		require.NoError(t, err)

		got := raw.(api.CreateAppPassword201JSONResponse)
		require.Equal(t, id, got.Id)
		require.Equal(t, "Phone", got.Name)
		require.Equal(t, "secret-token", got.Password)
	})

	t.Run("invalid argument", func(t *testing.T) {
		svc := &service.AppPasswordServiceMock{
			CreateFn: func(ctx context.Context, name string) (model.AppPassword, string, error) {
				return model.AppPassword{}, "", model.ErrInvalidArgument
			},
		}

		h := handlers.NewAppPasswordHandler(svc)

		raw, err := h.CreateAppPassword(context.Background(), api.CreateAppPasswordRequestObject{
			Body: &api.CreateAppPassword{Name: ""},
		})
		require.NoError(t, err)
		require.IsType(t, api.CreateAppPassword400Response{}, raw)
	})

	t.Run("service error", func(t *testing.T) {
		svc := &service.AppPasswordServiceMock{
			CreateFn: func(ctx context.Context, name string) (model.AppPassword, string, error) {
				return model.AppPassword{}, "", errors.New("service error")
			},
		}

		h := handlers.NewAppPasswordHandler(svc)

		_, err := h.CreateAppPassword(context.Background(), api.CreateAppPasswordRequestObject{
			Body: &api.CreateAppPassword{Name: "Phone"},
		})
		require.Error(t, err)
	})
}

func TestAppPasswordHandler_DeleteAppPassword(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		id := uuid.New()
		svc := &service.AppPasswordServiceMock{
			DeleteFn: func(ctx context.Context, got uuid.UUID) error {
				require.Equal(t, id, got)
				return nil
			},
		}

		h := handlers.NewAppPasswordHandler(svc)

		raw, err := h.DeleteAppPassword(context.Background(), api.DeleteAppPasswordRequestObject{AppPasswordId: id})
		require.NoError(t, err)
		require.IsType(t, api.DeleteAppPassword204Response{}, raw)
	})

	t.Run("not found", func(t *testing.T) {
		svc := &service.AppPasswordServiceMock{
			DeleteFn: func(ctx context.Context, id uuid.UUID) error { return model.ErrNotFound },
		}

		h := handlers.NewAppPasswordHandler(svc)

		raw, err := h.DeleteAppPassword(context.Background(), api.DeleteAppPasswordRequestObject{AppPasswordId: uuid.New()})
		require.NoError(t, err)
		require.IsType(t, api.DeleteAppPassword404Response{}, raw)
	})

	t.Run("service error", func(t *testing.T) {
		svc := &service.AppPasswordServiceMock{
			DeleteFn: func(ctx context.Context, id uuid.UUID) error { return errors.New("service error") },
		}

		h := handlers.NewAppPasswordHandler(svc)

		_, err := h.DeleteAppPassword(context.Background(), api.DeleteAppPasswordRequestObject{AppPasswordId: uuid.New()})
		require.Error(t, err)
	})
}
