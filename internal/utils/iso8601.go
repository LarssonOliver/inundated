package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var iso8601DurationRe = regexp.MustCompile(`^P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)W)?(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?)?$`)

var (
	ErrInvalidISO8601Format = errors.New("invalid ISO8601 format")
	ErrISO8601Unprocessable = errors.New("unprocessable ISO8601 value")
)

type ISO8601Duration struct {
	Years   int
	Months  int
	Weeks   int
	Days    int
	Hours   int
	Minutes int
	Seconds int
}

type ResolvedInterval struct {
	Start time.Time
	End   time.Time
}

type TimeBucket struct {
	Start time.Time
	End   time.Time
}

func ParseISO8601Duration(raw string) (ISO8601Duration, error) {
	if raw == "" {
		return ISO8601Duration{}, fmt.Errorf("duration is empty: %w", ErrInvalidISO8601Format)
	}

	m := iso8601DurationRe.FindStringSubmatch(raw)
	if m == nil {
		return ISO8601Duration{}, fmt.Errorf("invalid ISO8601 duration: %q: %w", raw, ErrInvalidISO8601Format)
	}

	hasAny := false
	for i := 1; i < len(m); i++ {
		if m[i] != "" {
			hasAny = true
			break
		}
	}
	if !hasAny {
		return ISO8601Duration{}, fmt.Errorf("duration has no units: %q: %w", raw, ErrInvalidISO8601Format)
	}

	parse := func(s string) (int, error) {
		if s == "" {
			return 0, nil
		}

		n, err := strconv.Atoi(s)
		if err != nil || n < 0 {
			return 0, fmt.Errorf("invalid duration part %q: %w", s, ErrInvalidISO8601Format)
		}

		return n, nil
	}

	years, err := parse(m[1])
	if err != nil {
		return ISO8601Duration{}, err
	}

	months, err := parse(m[2])
	if err != nil {
		return ISO8601Duration{}, err
	}

	weeks, err := parse(m[3])
	if err != nil {
		return ISO8601Duration{}, err
	}

	days, err := parse(m[4])
	if err != nil {
		return ISO8601Duration{}, err
	}

	hours, err := parse(m[5])
	if err != nil {
		return ISO8601Duration{}, err
	}

	minutes, err := parse(m[6])
	if err != nil {
		return ISO8601Duration{}, err
	}

	seconds, err := parse(m[7])
	if err != nil {
		return ISO8601Duration{}, err
	}

	return ISO8601Duration{
		Years:   years,
		Months:  months,
		Weeks:   weeks,
		Days:    days,
		Hours:   hours,
		Minutes: minutes,
		Seconds: seconds,
	}, nil
}

func ParseISO8601Interval(raw string, now time.Time, loc *time.Location) (ResolvedInterval, error) {
	_ = now

	if raw == "" {
		return ResolvedInterval{}, fmt.Errorf("interval is empty: %w", ErrInvalidISO8601Format)
	}

	if loc == nil {
		loc = time.UTC
	}

	parts := strings.Split(raw, "/")
	if len(parts) != 2 {
		return ResolvedInterval{}, fmt.Errorf("interval must contain exactly one slash: %q: %w", raw, ErrInvalidISO8601Format)
	}

	left := strings.TrimSpace(parts[0])
	right := strings.TrimSpace(parts[1])
	if left == "" || right == "" {
		return ResolvedInterval{}, fmt.Errorf("interval parts cannot be empty: %q: %w", raw, ErrInvalidISO8601Format)
	}

	parseTime := func(s string) (time.Time, bool) {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return time.Time{}, false
		}

		return t.In(loc), true
	}

	parseDuration := func(s string) (ISO8601Duration, bool) {
		d, err := ParseISO8601Duration(s)
		if err != nil {
			return ISO8601Duration{}, false
		}

		return d, true
	}

	leftTime, leftIsTime := parseTime(left)
	rightTime, rightIsTime := parseTime(right)
	leftDuration, leftIsDuration := parseDuration(left)
	rightDuration, rightIsDuration := parseDuration(right)

	var out ResolvedInterval

	switch {
	case leftIsTime && rightIsTime:
		out = ResolvedInterval{Start: leftTime, End: rightTime}

	case leftIsTime && rightIsDuration:
		end, err := addISO8601Duration(leftTime, rightDuration, loc)
		if err != nil {
			return ResolvedInterval{}, err
		}

		out = ResolvedInterval{Start: leftTime, End: end}

	case leftIsDuration && rightIsTime:
		start, err := addISO8601Duration(rightTime, negateDuration(leftDuration), loc)
		if err != nil {
			return ResolvedInterval{}, err
		}

		out = ResolvedInterval{Start: start, End: rightTime}

	default:
		return ResolvedInterval{}, fmt.Errorf("invalid interval expression: %q: %w", raw, ErrInvalidISO8601Format)
	}

	if !out.End.After(out.Start) {
		return ResolvedInterval{}, fmt.Errorf("interval end must be after start: %w", ErrISO8601Unprocessable)
	}

	return out, nil
}

func ParseTimezone(raw string) (*time.Location, error) {
	if strings.TrimSpace(raw) == "" {
		return time.UTC, nil
	}

	loc, err := time.LoadLocation(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone %q: %w", raw, err)
	}

	return loc, nil
}

func BuildTimeBuckets(interval ResolvedInterval, granularity ISO8601Duration, loc *time.Location, maxBuckets int) ([]TimeBucket, error) {
	if loc == nil {
		loc = time.UTC
	}

	if maxBuckets <= 0 {
		return nil, fmt.Errorf("maxBuckets must be > 0: %w", ErrISO8601Unprocessable)
	}

	if !interval.End.After(interval.Start) {
		return nil, fmt.Errorf("interval end must be after start: %w", ErrISO8601Unprocessable)
	}

	if hasNegativeDurationPart(granularity) {
		return nil, fmt.Errorf("negative granularity is not allowed: %w", ErrISO8601Unprocessable)
	}

	if isZeroDuration(granularity) {
		return nil, fmt.Errorf("zero granularity is not allowed: %w", ErrISO8601Unprocessable)
	}

	start := interval.Start.In(loc)
	end := interval.End.In(loc)

	buckets := make([]TimeBucket, 0, 16)
	cur := start

	for cur.Before(end) {
		next, err := addISO8601Duration(cur, granularity, loc)
		if err != nil {
			return nil, err
		}

		if !next.After(cur) {
			return nil, fmt.Errorf("granularity does not advance time: %w", ErrISO8601Unprocessable)
		}

		if next.After(end) {
			next = end
		}

		buckets = append(buckets, TimeBucket{
			Start: cur,
			End:   next,
		})

		if len(buckets) > maxBuckets {
			return nil, fmt.Errorf("bucket limit exceeded: %w", ErrISO8601Unprocessable)
		}

		cur = next
	}

	return buckets, nil
}

func addISO8601Duration(t time.Time, d ISO8601Duration, loc *time.Location) (time.Time, error) {
	if loc == nil {
		loc = time.UTC
	}

	t = t.In(loc)
	t = t.AddDate(d.Years, d.Months, d.Weeks*7+d.Days)
	t = t.Add(
		time.Duration(d.Hours)*time.Hour +
			time.Duration(d.Minutes)*time.Minute +
			time.Duration(d.Seconds)*time.Second,
	)

	return t, nil
}

func negateDuration(d ISO8601Duration) ISO8601Duration {
	return ISO8601Duration{
		Years:   -d.Years,
		Months:  -d.Months,
		Weeks:   -d.Weeks,
		Days:    -d.Days,
		Hours:   -d.Hours,
		Minutes: -d.Minutes,
		Seconds: -d.Seconds,
	}
}

func isZeroDuration(d ISO8601Duration) bool {
	return d.Years == 0 &&
		d.Months == 0 &&
		d.Weeks == 0 &&
		d.Days == 0 &&
		d.Hours == 0 &&
		d.Minutes == 0 &&
		d.Seconds == 0
}

func hasNegativeDurationPart(d ISO8601Duration) bool {
	return d.Years < 0 ||
		d.Months < 0 ||
		d.Weeks < 0 ||
		d.Days < 0 ||
		d.Hours < 0 ||
		d.Minutes < 0 ||
		d.Seconds < 0
}
