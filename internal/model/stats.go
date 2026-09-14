package model

import (
	"time"

	"github.com/google/uuid"
)

type StatsMetric string

const (
	StatsMetricTimeSpent StatsMetric = "time_spent"
)

type BucketRange struct {
	Start time.Time
	End   time.Time
}

type BucketValue struct {
	Bucket BucketRange
	Value  float64
}

type ProjectStats struct {
	ProjectID   uuid.UUID
	Metric      StatsMetric
	Interval    BucketRange
	Granularity string
	Unit        string
	Series      []BucketValue
}

type TagStats struct {
	TagID       uuid.UUID
	Metric      StatsMetric
	Interval    BucketRange
	Granularity string
	Unit        string
	Series      []BucketValue
}
