package model

import "github.com/google/uuid"

type Settings struct {
	Id             uuid.UUID
	UserId         *uuid.UUID
	WeekStartDay   string
	Timezone       string
	DurationFormat string
	TimeFormat     string
}

const (
	WeekStartMonday = "monday"
	WeekStartSunday = "sunday"

	DurationFormatLong    = "long"
	DurationFormatDecimal = "decimal"
	DurationFormatClock   = "clock"

	TimeFormat12h = "12h"
	TimeFormat24h = "24h"
)

// DefaultSettings is the row a scope gets the first time it's asked for its
// settings.
func DefaultSettings() Settings {
	return Settings{
		WeekStartDay:   WeekStartMonday,
		Timezone:       "UTC",
		DurationFormat: DurationFormatLong,
		TimeFormat:     TimeFormat24h,
	}
}
