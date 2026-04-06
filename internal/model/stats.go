package model

// TimeSeriesPoint represents a single data point in a time series
type TimeSeriesPoint struct {
	// Interval is the time bucket for this data point as an ISO 8601 interval
	// in {start}/{duration} form (e.g. "2024-01-01T00:00:00Z/P1W")
	Interval string
	// Value is the aggregated metric value for this bucket
	Value int64
}

// ProjectStats represents aggregated time-series data for a project metric
type ProjectStats struct {
	// ProjectID is the ID of the project this data belongs to
	ProjectID string
	// Metric is the metric that was aggregated (e.g., "time_spent")
	Metric string
	// Interval is the effective time range of the response as an ISO 8601 interval,
	// always resolved to {start}/{end} form
	Interval string
	// Granularity is the ISO 8601 duration used to bucket each series point
	Granularity string
	// Unit is the unit of the value field in each series point
	// (e.g., "seconds" for time_spent)
	Unit string
	// Series is the ordered list of aggregated data points
	Series []TimeSeriesPoint
}
