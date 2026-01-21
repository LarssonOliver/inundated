package utils

import "time"

func FloatHoursToDuration(hours *float64) *time.Duration {
	if hours == nil {
		return nil
	}

	duration := time.Duration(*hours * float64(time.Hour))
	return &duration
}

func DurationToFloatHours(duration *time.Duration) *float64 {
	if duration == nil {
		return nil
	}

	hours := duration.Hours()
	return &hours
}
