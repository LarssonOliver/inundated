package utils

import (
	"testing"
	"time"
)

func TestParseISO8601Duration(t *testing.T) {
	tests := []struct {
		name     string
		duration string
		want     time.Duration
		wantErr  bool
	}{
		{
			name:     "1 day",
			duration: "P1D",
			want:     24 * time.Hour,
		},
		{
			name:     "1 week",
			duration: "P1W",
			want:     7 * 24 * time.Hour,
		},
		{
			name:     "1 hour",
			duration: "PT1H",
			want:     time.Hour,
		},
		{
			name:     "1 minute",
			duration: "PT1M",
			want:     time.Minute,
		},
		{
			name:     "30 seconds",
			duration: "PT30S",
			want:     30 * time.Second,
		},
		{
			name:     "1 month (30 days)",
			duration: "P1M",
			want:     30 * 24 * time.Hour,
		},
		{
			name:     "3 months",
			duration: "P3M",
			want:     90 * 24 * time.Hour,
		},
		{
			name:     "1 year (365 days)",
			duration: "P1Y",
			want:     365 * 24 * time.Hour,
		},
		{
			name:     "30 days",
			duration: "P30D",
			want:     30 * 24 * time.Hour,
		},
		{
			name:     "complex: 1 day 2 hours 30 minutes",
			duration: "P1DT2H30M",
			want:     24*time.Hour + 2*time.Hour + 30*time.Minute,
		},
		{
			name:     "1 week 2 days",
			duration: "P1W2D",
			want:     9 * 24 * time.Hour,
		},
		{
			name:     "decimal seconds",
			duration: "PT1.5S",
			want:     1500 * time.Millisecond,
		},
		{
			name:     "invalid: missing P",
			duration: "1D",
			wantErr:  true,
		},
		{
			name:     "invalid: malformed",
			duration: "P1X",
			wantErr:  true,
		},
		{
			name:     "invalid: empty duration",
			duration: "P",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseISO8601Duration(tt.duration)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseISO8601Duration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseISO8601Duration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseISO8601Interval(t *testing.T) {
	now := time.Date(2024, 4, 1, 12, 0, 0, 0, time.UTC)
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 3, 31, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name     string
		interval string
		now      time.Time
		want     Interval
		wantErr  bool
	}{
		{
			name:     "start/end form",
			interval: "2024-01-01T00:00:00Z/2024-03-31T23:59:59Z",
			now:      now,
			want: Interval{
				Start: start,
				End:   end,
			},
		},
		{
			name:     "start/duration form",
			interval: "2024-01-01T00:00:00Z/P30D",
			now:      now,
			want: Interval{
				Start: start,
				End:   start.Add(30 * 24 * time.Hour),
			},
		},
		{
			name:     "duration/end form",
			interval: "P30D/2024-03-31T23:59:59Z",
			now:      now,
			want: Interval{
				Start: end.Add(-30 * 24 * time.Hour),
				End:   end,
			},
		},
		{
			name:     "now token in end",
			interval: "P30D/{now}",
			now:      now,
			want: Interval{
				Start: now.Add(-30 * 24 * time.Hour),
				End:   now,
			},
		},
		{
			name:     "start/duration with complex duration",
			interval: "2024-01-01T00:00:00Z/P3M",
			now:      now,
			want: Interval{
				Start: start,
				End:   start.Add(90 * 24 * time.Hour),
			},
		},
		{
			name:     "invalid: empty string",
			interval: "",
			now:      now,
			wantErr:  true,
		},
		{
			name:     "invalid: no separator",
			interval: "2024-01-01T00:00:00Z",
			now:      now,
			wantErr:  true,
		},
		{
			name:     "invalid: both parts are durations",
			interval: "P1D/P2D",
			now:      now,
			wantErr:  true,
		},
		{
			name:     "invalid: malformed datetime",
			interval: "2024-13-01T00:00:00Z/2024-03-31T23:59:59Z",
			now:      now,
			wantErr:  true,
		},
		{
			name:     "invalid: malformed duration",
			interval: "2024-01-01T00:00:00Z/P1X",
			now:      now,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseISO8601Interval(tt.interval, tt.now)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseISO8601Interval() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !got.Start.Equal(tt.want.Start) || !got.End.Equal(tt.want.End) {
					t.Errorf("ParseISO8601Interval() = {%v, %v}, want {%v, %v}",
						got.Start, got.End, tt.want.Start, tt.want.End)
				}
			}
		})
	}
}

func TestIntervalDuration(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	interval := Interval{Start: start, End: end}
	got := interval.Duration()
	want := 24 * time.Hour

	if got != want {
		t.Errorf("Interval.Duration() = %v, want %v", got, want)
	}
}

func TestFormatIntervalAsISO8601(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 3, 31, 23, 59, 59, 0, time.UTC)

	interval := Interval{Start: start, End: end}
	got := FormatIntervalAsISO8601(interval)
	want := "2024-01-01T00:00:00Z/2024-03-31T23:59:59Z"

	if got != want {
		t.Errorf("FormatIntervalAsISO8601() = %v, want %v", got, want)
	}
}

func TestFormatDurationAsISO8601(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "zero",
			duration: 0,
			want:     "PT0S",
		},
		{
			name:     "1 hour",
			duration: time.Hour,
			want:     "PT1H",
		},
		{
			name:     "1 day",
			duration: 24 * time.Hour,
			want:     "P1D",
		},
		{
			name:     "1 week",
			duration: 7 * 24 * time.Hour,
			want:     "P1W",
		},
		{
			name:     "30 minutes",
			duration: 30 * time.Minute,
			want:     "PT30M",
		},
		{
			name:     "1 hour 30 minutes",
			duration: time.Hour + 30*time.Minute,
			want:     "PT1H30M",
		},
		{
			name:     "1 day 2 hours",
			duration: 24*time.Hour + 2*time.Hour,
			want:     "P1DT2H",
		},
		{
			name:     "9 days",
			duration: 9 * 24 * time.Hour,
			want:     "P9D",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDurationAsISO8601(tt.duration)
			if got != tt.want {
				t.Errorf("FormatDurationAsISO8601() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatIntervalWithDuration(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	duration := 7 * 24 * time.Hour

	got := FormatIntervalWithDuration(start, duration)
	want := "2024-01-01T00:00:00Z/P1W"

	if got != want {
		t.Errorf("FormatIntervalWithDuration() = %v, want %v", got, want)
	}
}
