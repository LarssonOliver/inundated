package utils

import (
	"fmt"
	"time"
)

// Bucket represents a time bucket with start and end times
type Bucket struct {
	Start time.Time
	End   time.Time
}

// Duration returns the duration of the bucket
func (b Bucket) Duration() time.Duration {
	return b.End.Sub(b.Start)
}

// Overlaps checks if the bucket overlaps with a given time range
func (b Bucket) Overlaps(start, end time.Time) bool {
	return b.Start.Before(end) && b.End.After(start)
}

// OverlapDuration returns the duration of overlap between the bucket and a given time range
func (b Bucket) OverlapDuration(start, end time.Time) time.Duration {
	if !b.Overlaps(start, end) {
		return 0
	}

	overlapStart := b.Start
	if start.After(overlapStart) {
		overlapStart = start
	}

	overlapEnd := b.End
	if end.Before(overlapEnd) {
		overlapEnd = end
	}

	return overlapEnd.Sub(overlapStart)
}

// GenerateBuckets generates time buckets for the given interval and granularity in the specified timezone.
// The granularity is an ISO 8601 duration string (e.g., "P1D", "PT1H", "P1W").
// Buckets are aligned to natural boundaries in the given timezone (e.g., day boundaries for P1D).
func GenerateBuckets(interval Interval, granularity string, tz *time.Location) ([]Bucket, error) {
	if tz == nil {
		tz = time.UTC
	}

	duration, err := ParseISO8601Duration(granularity)
	if err != nil {
		return nil, fmt.Errorf("invalid granularity: %w", err)
	}

	// Convert interval times to the target timezone
	start := interval.Start.In(tz)
	end := interval.End.In(tz)

	// Determine if we need calendar-aware bucketing (day, week, month)
	// For simplicity, we'll detect this based on the granularity string
	isCalendarBased := isCalendarBasedGranularity(granularity)

	var buckets []Bucket

	if isCalendarBased {
		buckets, err = generateCalendarBuckets(start, end, granularity, tz)
		if err != nil {
			return nil, err
		}
	} else {
		buckets = generateFixedBuckets(start, end, duration)
	}

	return buckets, nil
}

// isCalendarBasedGranularity checks if the granularity requires calendar-aware bucketing
func isCalendarBasedGranularity(granularity string) bool {
	// Day, week, month, and year granularities need calendar awareness
	return granularity == "P1D" || granularity == "P1W" || granularity == "P1M" || granularity == "P1Y" ||
		// Also handle multiples
		len(granularity) >= 3 && (granularity[len(granularity)-1] == 'D' || granularity[len(granularity)-1] == 'W' || granularity[len(granularity)-1] == 'M' || granularity[len(granularity)-1] == 'Y') && granularity[0] == 'P' && granularity[1] != 'T'
}

// generateFixedBuckets generates buckets with fixed duration (e.g., hours, minutes)
func generateFixedBuckets(start, end time.Time, duration time.Duration) []Bucket {
	var buckets []Bucket
	current := start

	for current.Before(end) {
		bucketEnd := current.Add(duration)
		if bucketEnd.After(end) {
			bucketEnd = end
		}

		buckets = append(buckets, Bucket{
			Start: current,
			End:   bucketEnd,
		})

		current = bucketEnd
	}

	return buckets
}

// generateCalendarBuckets generates buckets aligned to calendar boundaries (day, week, month)
func generateCalendarBuckets(start, end time.Time, granularity string, tz *time.Location) ([]Bucket, error) {
	var buckets []Bucket

	switch granularity {
	case "P1D":
		// Align to day boundaries in the given timezone
		current := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, tz)
		if current.Before(start) {
			current = current.AddDate(0, 0, 1)
		}

		for current.Before(end) {
			bucketEnd := current.AddDate(0, 0, 1)
			if bucketEnd.After(end) {
				bucketEnd = end
			}

			// Include bucket if it overlaps with the interval
			if bucketEnd.After(start) {
				bucketStart := current
				if bucketStart.Before(start) {
					bucketStart = start
				}
				buckets = append(buckets, Bucket{
					Start: bucketStart,
					End:   bucketEnd,
				})
			}

			current = current.AddDate(0, 0, 1)
		}

	case "P1W":
		// Align to week boundaries (Monday 00:00 in the given timezone)
		current := start
		// Find the Monday of the week containing start
		weekday := int(current.Weekday())
		if weekday == 0 { // Sunday
			weekday = 7
		}
		daysFromMonday := weekday - 1
		current = time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, tz).AddDate(0, 0, -daysFromMonday)

		if current.Before(start) {
			current = current.AddDate(0, 0, 7)
		}

		for current.Before(end) {
			bucketEnd := current.AddDate(0, 0, 7)
			if bucketEnd.After(end) {
				bucketEnd = end
			}

			// Include bucket if it overlaps with the interval
			if bucketEnd.After(start) {
				bucketStart := current
				if bucketStart.Before(start) {
					bucketStart = start
				}
				buckets = append(buckets, Bucket{
					Start: bucketStart,
					End:   bucketEnd,
				})
			}

			current = current.AddDate(0, 0, 7)
		}

	case "P1M":
		// Align to month boundaries in the given timezone
		current := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, tz)
		if current.Before(start) {
			current = current.AddDate(0, 1, 0)
		}

		for current.Before(end) {
			bucketEnd := current.AddDate(0, 1, 0)
			if bucketEnd.After(end) {
				bucketEnd = end
			}

			// Include bucket if it overlaps with the interval
			if bucketEnd.After(start) {
				bucketStart := current
				if bucketStart.Before(start) {
					bucketStart = start
				}
				buckets = append(buckets, Bucket{
					Start: bucketStart,
					End:   bucketEnd,
				})
			}

			current = current.AddDate(0, 1, 0)
		}

	default:
		return nil, fmt.Errorf("unsupported calendar-based granularity: %s", granularity)
	}

	return buckets, nil
}

// SplitDurationAcrossBuckets splits a duration proportionally across buckets that overlap with the given time range.
// Returns a map from bucket index to the duration portion that falls within that bucket.
func SplitDurationAcrossBuckets(buckets []Bucket, spanStart, spanEnd time.Time, totalDuration time.Duration) map[int]time.Duration {
	result := make(map[int]time.Duration)

	spanDuration := spanEnd.Sub(spanStart)
	if spanDuration <= 0 {
		return result
	}

	for i, bucket := range buckets {
		overlapDuration := bucket.OverlapDuration(spanStart, spanEnd)
		if overlapDuration > 0 {
			// Calculate proportional duration
			proportion := float64(overlapDuration) / float64(spanDuration)
			bucketDuration := time.Duration(proportion * float64(totalDuration))
			result[i] = bucketDuration
		}
	}

	return result
}
