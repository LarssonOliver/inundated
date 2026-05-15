package model

import (
	"time"

	"github.com/google/uuid"
)

type ProjectStatsMetric string

const (
	ProjectStatsMetricTimeSpent ProjectStatsMetric = "time_spent"
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
	Metric      ProjectStatsMetric
	Interval    BucketRange
	Granularity string
	Unit        string
	Series      []BucketValue
}
