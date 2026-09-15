package handlers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/larssonoliver/inundated/internal/api"
	"github.com/larssonoliver/inundated/internal/api/handlers"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/service"
	"github.com/stretchr/testify/require"
)

func TestSettingsHandler_GetSettings(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		svc := &service.SettingsServiceMock{
			GetFn: func(ctx context.Context) (model.Settings, error) {
				return model.Settings{
					WeekStartDay: model.WeekStartSunday, Timezone: "Europe/Stockholm",
					DurationFormat: model.DurationFormatDecimal, TimeFormat: model.TimeFormat12h,
				}, nil
			},
		}

		h := handlers.NewSettingsHandler(svc)

		raw, err := h.GetSettings(context.Background(), api.GetSettingsRequestObject{})
		require.NoError(t, err)

		got := raw.(api.GetSettings200JSONResponse)
		require.Equal(t, api.Sunday, got.WeekStartDay)
		require.Equal(t, "Europe/Stockholm", got.Timezone)
		require.Equal(t, api.Decimal, got.DurationFormat)
		require.Equal(t, api.N12h, got.TimeFormat)
	})

	t.Run("service error", func(t *testing.T) {
		svc := &service.SettingsServiceMock{
			GetFn: func(ctx context.Context) (model.Settings, error) {
				return model.Settings{}, errors.New("service error")
			},
		}

		h := handlers.NewSettingsHandler(svc)

		_, err := h.GetSettings(context.Background(), api.GetSettingsRequestObject{})
		require.Error(t, err)
	})
}

func TestSettingsHandler_UpdateSettings(t *testing.T) {
	current := model.Settings{
		WeekStartDay: model.WeekStartMonday, Timezone: "UTC",
		DurationFormat: model.DurationFormatLong, TimeFormat: model.TimeFormat24h,
	}

	t.Run("partial update merges onto current settings", func(t *testing.T) {
		var updateArg model.Settings
		svc := &service.SettingsServiceMock{
			GetFn: func(ctx context.Context) (model.Settings, error) { return current, nil },
			UpdateFn: func(ctx context.Context, settings model.Settings) (model.Settings, error) {
				updateArg = settings
				return settings, nil
			},
		}

		h := handlers.NewSettingsHandler(svc)

		sunday := api.Sunday
		raw, err := h.UpdateSettings(context.Background(), api.UpdateSettingsRequestObject{
			Body: &api.UpdateSettings{WeekStartDay: &sunday},
		})
		require.NoError(t, err)

		require.Equal(t, model.WeekStartSunday, updateArg.WeekStartDay)
		require.Equal(t, "UTC", updateArg.Timezone, "fields not in the patch must carry the current value forward")
		require.Equal(t, model.DurationFormatLong, updateArg.DurationFormat)
		require.Equal(t, model.TimeFormat24h, updateArg.TimeFormat)

		got := raw.(api.UpdateSettings200JSONResponse)
		require.Equal(t, api.Sunday, got.WeekStartDay)
	})

	t.Run("invalid argument maps to 400", func(t *testing.T) {
		svc := &service.SettingsServiceMock{
			GetFn: func(ctx context.Context) (model.Settings, error) { return current, nil },
			UpdateFn: func(ctx context.Context, settings model.Settings) (model.Settings, error) {
				return model.Settings{}, model.ErrInvalidArgument
			},
		}

		h := handlers.NewSettingsHandler(svc)

		badTz := "Not/AZone"
		raw, err := h.UpdateSettings(context.Background(), api.UpdateSettingsRequestObject{
			Body: &api.UpdateSettings{Timezone: &badTz},
		})
		require.NoError(t, err)
		_, is400 := raw.(api.UpdateSettings400Response)
		require.True(t, is400)
	})

	t.Run("get error propagates", func(t *testing.T) {
		svc := &service.SettingsServiceMock{
			GetFn: func(ctx context.Context) (model.Settings, error) {
				return model.Settings{}, errors.New("service error")
			},
		}

		h := handlers.NewSettingsHandler(svc)

		_, err := h.UpdateSettings(context.Background(), api.UpdateSettingsRequestObject{
			Body: &api.UpdateSettings{},
		})
		require.Error(t, err)
	})
}
