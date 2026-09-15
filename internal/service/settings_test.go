package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/service"
	"github.com/stretchr/testify/require"
)

func TestSettingsService_GetSettings_ReturnsExisting(t *testing.T) {
	existing := model.Settings{Id: uuid.New(), WeekStartDay: model.WeekStartSunday,
		Timezone: "UTC", DurationFormat: model.DurationFormatLong, TimeFormat: model.TimeFormat24h}

	repo := &repository.RepoMock{
		GetSettingsFn: func(ctx context.Context, scope model.OwnerScope) (model.Settings, error) {
			return existing, nil
		},
	}

	got, err := service.NewService(repo).GetSettings(context.Background())
	require.NoError(t, err)
	require.Equal(t, existing, got)
}

func TestSettingsService_GetSettings_CreatesDefaultsWhenMissing(t *testing.T) {
	var created model.Settings
	var createCalled bool

	repo := &repository.RepoMock{
		GetSettingsFn: func(ctx context.Context, scope model.OwnerScope) (model.Settings, error) {
			return model.Settings{}, model.ErrNotFound
		},
		CreateSettingsFn: func(ctx context.Context, scope model.OwnerScope, settings model.Settings) (model.Settings, error) {
			createCalled = true
			created = settings
			return settings, nil
		},
	}

	got, err := service.NewService(repo).GetSettings(context.Background())
	require.NoError(t, err)
	require.True(t, createCalled)
	require.Equal(t, model.DefaultSettings(), created)
	require.Equal(t, model.DefaultSettings(), got)
}

func TestSettingsService_GetSettings_ConcurrentCreateLosesRace(t *testing.T) {
	winner := model.Settings{WeekStartDay: model.WeekStartSunday, Timezone: "UTC",
		DurationFormat: model.DurationFormatLong, TimeFormat: model.TimeFormat24h}

	getCalls := 0
	repo := &repository.RepoMock{
		GetSettingsFn: func(ctx context.Context, scope model.OwnerScope) (model.Settings, error) {
			getCalls++
			if getCalls == 1 {
				return model.Settings{}, model.ErrNotFound
			}
			return winner, nil
		},
		CreateSettingsFn: func(ctx context.Context, scope model.OwnerScope, settings model.Settings) (model.Settings, error) {
			return model.Settings{}, model.ErrAlreadyExists
		},
	}

	got, err := service.NewService(repo).GetSettings(context.Background())
	require.NoError(t, err)
	require.Equal(t, winner, got)
}

func TestSettingsService_GetSettings_RepositoryError(t *testing.T) {
	repo := &repository.RepoMock{
		GetSettingsFn: func(ctx context.Context, scope model.OwnerScope) (model.Settings, error) {
			return model.Settings{}, errors.New("database error")
		},
	}

	_, err := service.NewService(repo).GetSettings(context.Background())
	require.Error(t, err)
}

func TestSettingsService_UpdateSettings_Success(t *testing.T) {
	want := model.Settings{WeekStartDay: model.WeekStartSunday, Timezone: "Europe/Stockholm",
		DurationFormat: model.DurationFormatDecimal, TimeFormat: model.TimeFormat12h}

	repo := &repository.RepoMock{
		UpdateSettingsFn: func(ctx context.Context, scope model.OwnerScope, settings model.Settings) (model.Settings, error) {
			return settings, nil
		},
	}

	got, err := service.NewService(repo).UpdateSettings(context.Background(), want)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestSettingsService_UpdateSettings_RejectsInvalidFields(t *testing.T) {
	tests := []struct {
		name     string
		settings model.Settings
	}{
		{"bad week start day", model.Settings{WeekStartDay: "tuesday", Timezone: "UTC", DurationFormat: model.DurationFormatLong, TimeFormat: model.TimeFormat24h}},
		{"bad duration format", model.Settings{WeekStartDay: model.WeekStartMonday, Timezone: "UTC", DurationFormat: "fancy", TimeFormat: model.TimeFormat24h}},
		{"bad time format", model.Settings{WeekStartDay: model.WeekStartMonday, Timezone: "UTC", DurationFormat: model.DurationFormatLong, TimeFormat: "30h"}},
		{"bad timezone", model.Settings{WeekStartDay: model.WeekStartMonday, Timezone: "Nowhere/Fake", DurationFormat: model.DurationFormatLong, TimeFormat: model.TimeFormat24h}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repository.RepoMock{
				UpdateSettingsFn: func(ctx context.Context, scope model.OwnerScope, settings model.Settings) (model.Settings, error) {
					t.Fatal("repository must not be called with invalid settings")
					return model.Settings{}, nil
				},
			}

			_, err := service.NewService(repo).UpdateSettings(context.Background(), tt.settings)
			require.ErrorIs(t, err, model.ErrInvalidArgument)
		})
	}
}
