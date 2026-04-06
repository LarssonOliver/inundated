package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Interval represents a resolved time interval with concrete start and end times
type Interval struct {
	Start time.Time
	End   time.Time
}

// Duration returns the duration of the interval
func (i Interval) Duration() time.Duration {
	return i.End.Sub(i.Start)
}

// ParseISO8601Interval parses an ISO 8601 interval string in any of the three forms:
// - {start}/{end}: "2024-01-01T00:00:00Z/2024-03-31T23:59:59Z"
// - {start}/{duration}: "2024-01-01T00:00:00Z/P3M"
// - {duration}/{end}: "P30D/2024-03-31T23:59:59Z"
//
// The special token "{now}" can be used in place of an end datetime.
// Example: "P30D/{now}" means the last 30 days from now.
func ParseISO8601Interval(interval string, now time.Time) (Interval, error) {
	if interval == "" {
		return Interval{}, fmt.Errorf("interval string is empty")
	}

	parts := strings.Split(interval, "/")
	if len(parts) != 2 {
		return Interval{}, fmt.Errorf("invalid interval format: expected exactly one '/' separator")
	}

	left, right := parts[0], parts[1]

	// Handle {now} substitution
	if right == "{now}" {
		right = now.Format(time.RFC3339)
	}

	// Determine which form we have
	leftIsDuration := strings.HasPrefix(left, "P")
	rightIsDuration := strings.HasPrefix(right, "P")

	switch {
	case !leftIsDuration && !rightIsDuration:
		// Form 1: {start}/{end}
		start, err := time.Parse(time.RFC3339, left)
		if err != nil {
			return Interval{}, fmt.Errorf("failed to parse start datetime: %w", err)
		}
		end, err := time.Parse(time.RFC3339, right)
		if err != nil {
			return Interval{}, fmt.Errorf("failed to parse end datetime: %w", err)
		}
		return Interval{Start: start, End: end}, nil

	case !leftIsDuration && rightIsDuration:
		// Form 2: {start}/{duration}
		start, err := time.Parse(time.RFC3339, left)
		if err != nil {
			return Interval{}, fmt.Errorf("failed to parse start datetime: %w", err)
		}
		duration, err := ParseISO8601Duration(right)
		if err != nil {
			return Interval{}, fmt.Errorf("failed to parse duration: %w", err)
		}
		end := start.Add(duration)
		return Interval{Start: start, End: end}, nil

	case leftIsDuration && !rightIsDuration:
		// Form 3: {duration}/{end}
		duration, err := ParseISO8601Duration(left)
		if err != nil {
			return Interval{}, fmt.Errorf("failed to parse duration: %w", err)
		}
		end, err := time.Parse(time.RFC3339, right)
		if err != nil {
			return Interval{}, fmt.Errorf("failed to parse end datetime: %w", err)
		}
		start := end.Add(-duration)
		return Interval{Start: start, End: end}, nil

	default:
		return Interval{}, fmt.Errorf("invalid interval format: both parts cannot be durations")
	}
}

// ParseISO8601Duration parses an ISO 8601 duration string like "P1D", "PT1H", "P1W", "P3M", etc.
// Returns a time.Duration. Note: For simplicity, months are treated as 30 days and years as 365 days.
func ParseISO8601Duration(duration string) (time.Duration, error) {
	if !strings.HasPrefix(duration, "P") {
		return 0, fmt.Errorf("duration must start with 'P'")
	}

	// Pattern: P[n]Y[n]M[n]W[n]DT[n]H[n]M[n]S
	// We'll use a regex to extract the parts
	pattern := `^P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)W)?(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:\.\d+)?)S)?)?$`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(duration)

	if matches == nil {
		return 0, fmt.Errorf("invalid ISO 8601 duration format: %s", duration)
	}

	var total time.Duration

	// Years (approximate as 365 days)
	if matches[1] != "" {
		years, _ := strconv.Atoi(matches[1])
		total += time.Duration(years) * 365 * 24 * time.Hour
	}

	// Months (approximate as 30 days)
	if matches[2] != "" {
		months, _ := strconv.Atoi(matches[2])
		total += time.Duration(months) * 30 * 24 * time.Hour
	}

	// Weeks
	if matches[3] != "" {
		weeks, _ := strconv.Atoi(matches[3])
		total += time.Duration(weeks) * 7 * 24 * time.Hour
	}

	// Days
	if matches[4] != "" {
		days, _ := strconv.Atoi(matches[4])
		total += time.Duration(days) * 24 * time.Hour
	}

	// Hours
	if matches[5] != "" {
		hours, _ := strconv.Atoi(matches[5])
		total += time.Duration(hours) * time.Hour
	}

	// Minutes
	if matches[6] != "" {
		minutes, _ := strconv.Atoi(matches[6])
		total += time.Duration(minutes) * time.Minute
	}

	// Seconds (can be decimal)
	if matches[7] != "" {
		seconds, _ := strconv.ParseFloat(matches[7], 64)
		total += time.Duration(seconds * float64(time.Second))
	}

	if total == 0 {
		return 0, fmt.Errorf("duration must have at least one non-zero component")
	}

	return total, nil
}

// FormatIntervalAsISO8601 formats an Interval as an ISO 8601 interval string in {start}/{end} form
func FormatIntervalAsISO8601(interval Interval) string {
	return fmt.Sprintf("%s/%s", interval.Start.Format(time.RFC3339), interval.End.Format(time.RFC3339))
}

// FormatIntervalWithDuration formats an Interval as an ISO 8601 interval string in {start}/{duration} form
func FormatIntervalWithDuration(start time.Time, duration time.Duration) string {
	// Format duration as ISO 8601 (simplified for common cases)
	durationStr := FormatDurationAsISO8601(duration)
	return fmt.Sprintf("%s/%s", start.Format(time.RFC3339), durationStr)
}

// FormatDurationAsISO8601 formats a time.Duration as an ISO 8601 duration string
// This is a simplified implementation for common cases
func FormatDurationAsISO8601(d time.Duration) string {
	if d == 0 {
		return "PT0S"
	}

	// Extract components
	hours := int(d.Hours())
	d -= time.Duration(hours) * time.Hour
	minutes := int(d.Minutes())
	d -= time.Duration(minutes) * time.Minute
	seconds := int(d.Seconds())

	// Build string
	result := "P"

	// Check if we have whole days/weeks
	if hours >= 24 {
		days := hours / 24
		hours = hours % 24

		if days >= 7 && days%7 == 0 {
			weeks := days / 7
			result += fmt.Sprintf("%dW", weeks)
		} else {
			result += fmt.Sprintf("%dD", days)
		}
	}

	// Add time component if needed
	if hours > 0 || minutes > 0 || seconds > 0 {
		result += "T"
		if hours > 0 {
			result += fmt.Sprintf("%dH", hours)
		}
		if minutes > 0 {
			result += fmt.Sprintf("%dM", minutes)
		}
		if seconds > 0 {
			result += fmt.Sprintf("%dS", seconds)
		}
	}

	// If we only have date components and they're all zero, we need PT0S
	if result == "P" {
		result = "PT0S"
	}

	return result
}
