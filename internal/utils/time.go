package utils

import (
	"errors"
	"fmt"
	"math"
	"time"
)

// ErrHoursOutOfRange reports an hour count too large for a time.Duration.
var ErrHoursOutOfRange = errors.New("hours out of range")

// maxHours is the largest whole number of hours a time.Duration holds
// (about 292 years).
const maxHours = float64(math.MaxInt64 / int64(time.Hour))

// FloatHoursToDuration converts hours to a Duration, or fails with
// ErrHoursOutOfRange when the value doesn't fit (or is NaN): converting an
// out-of-range float to an integer gives a different result on different
// CPUs.
func FloatHoursToDuration(hours *float64) (*time.Duration, error) {
	if hours == nil {
		return nil, nil
	}
	if math.IsNaN(*hours) || math.Abs(*hours) > maxHours {
		return nil, fmt.Errorf("%v: %w", *hours, ErrHoursOutOfRange)
	}

	duration := time.Duration(*hours * float64(time.Hour))
	return &duration, nil
}

func DurationToFloatHours(duration *time.Duration) *float64 {
	if duration == nil {
		return nil
	}

	hours := duration.Hours()
	return &hours
}
