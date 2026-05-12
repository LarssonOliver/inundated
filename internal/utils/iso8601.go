package utils

import (
	"errors"
	"time"
)

var errISO8601NotImplemented = errors.New("iso8601 utility not implemented")

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
	return ISO8601Duration{}, errISO8601NotImplemented
}

func ParseISO8601Interval(raw string, now time.Time, loc *time.Location) (ResolvedInterval, error) {
	return ResolvedInterval{}, errISO8601NotImplemented
}

func ParseTimezone(raw string) (*time.Location, error) {
	return nil, errISO8601NotImplemented
}

func BuildTimeBuckets(interval ResolvedInterval, granularity ISO8601Duration, loc *time.Location, maxBuckets int) ([]TimeBucket, error) {
	return nil, errISO8601NotImplemented
}
