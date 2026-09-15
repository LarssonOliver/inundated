package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/larssonoliver/inundated/internal/model"
)

var validWeekStartDays = map[string]bool{
	model.WeekStartMonday: true,
	model.WeekStartSunday: true,
}

var validDurationFormats = map[string]bool{
	model.DurationFormatLong:    true,
	model.DurationFormatDecimal: true,
	model.DurationFormatClock:   true,
}

var validTimeFormats = map[string]bool{
	model.TimeFormat12h: true,
	model.TimeFormat24h: true,
}

func validateSettings(s model.Settings) error {
	if !validWeekStartDays[s.WeekStartDay] {
		return fmt.Errorf("weekStartDay %q: %w", s.WeekStartDay, model.ErrInvalidArgument)
	}
	if !validDurationFormats[s.DurationFormat] {
		return fmt.Errorf("durationFormat %q: %w", s.DurationFormat, model.ErrInvalidArgument)
	}
	if !validTimeFormats[s.TimeFormat] {
		return fmt.Errorf("timeFormat %q: %w", s.TimeFormat, model.ErrInvalidArgument)
	}
	if _, err := time.LoadLocation(s.Timezone); err != nil {
		return fmt.Errorf("timezone %q: %w", s.Timezone, model.ErrInvalidArgument)
	}
	return nil
}

// GetSettings implements [SettingsService].
func (s *ServiceImpl) GetSettings(ctx context.Context) (model.Settings, error) {
	scope, err := ownerScope(ctx)
	if err != nil {
		return model.Settings{}, err
	}

	settings, err := s.repository.GetSettings(ctx, scope)
	if errors.Is(err, model.ErrNotFound) {
		return s.createDefaultSettings(ctx, scope)
	}
	if err != nil {
		return model.Settings{}, err
	}
	return settings, nil
}

func (s *ServiceImpl) createDefaultSettings(ctx context.Context, scope model.OwnerScope) (model.Settings, error) {
	created, err := s.repository.CreateSettings(ctx, scope, model.DefaultSettings())
	if errors.Is(err, model.ErrAlreadyExists) {
		// A concurrent request (two tabs, a double-fired first load) created
		// this scope's row between our lookup and this insert. Adopt the
		// winner rather than failing this one.
		return s.repository.GetSettings(ctx, scope)
	}
	if err != nil {
		return model.Settings{}, err
	}
	return created, nil
}

// UpdateSettings implements [SettingsService].
func (s *ServiceImpl) UpdateSettings(ctx context.Context, settings model.Settings) (model.Settings, error) {
	scope, err := ownerScope(ctx)
	if err != nil {
		return model.Settings{}, err
	}
	if err := validateSettings(settings); err != nil {
		return model.Settings{}, err
	}
	return s.repository.UpdateSettings(ctx, scope, settings)
}
