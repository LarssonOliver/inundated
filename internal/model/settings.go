package model

import "github.com/google/uuid"

type Settings struct {
	Id             uuid.UUID
	UserId         *uuid.UUID
	WeekStartDay   string
	Timezone       string
	DurationFormat string
	TimeFormat     string
	DateFormat     string
}

const (
	WeekStartMonday = "monday"
	WeekStartSunday = "sunday"

	DurationFormatLong    = "long"
	DurationFormatDecimal = "decimal"
	DurationFormatClock   = "clock"

	TimeFormat12h = "12h"
	TimeFormat24h = "24h"

	DateFormatISO  = "iso"  // 2024-01-15
	DateFormatUS   = "us"   // 01/15/2024
	DateFormatEU   = "eu"   // 15/01/2024
	DateFormatText = "text" // 15 Jan 2024

	// TimezoneBrowser is a sentinel Timezone value meaning "use whatever
	// timezone the client's browser is currently in" rather than a fixed
	// IANA zone. The backend only stores and passes this through - it never
	// resolves it, since only the client knows its own local zone.
	TimezoneBrowser = "browser"
)

// DefaultSettings is the row a scope gets the first time it's asked for its
// settings.
func DefaultSettings() Settings {
	return Settings{
		WeekStartDay:   WeekStartMonday,
		Timezone:       "UTC",
		DurationFormat: DurationFormatLong,
		TimeFormat:     TimeFormat24h,
		DateFormat:     DateFormatISO,
	}
}
