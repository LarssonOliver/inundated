package utils_test

import (
	"testing"
	"time"

	"github.com/larssonoliver/inundated/internal/utils"
	"github.com/stretchr/testify/require"
)

func mustRFC3339(t *testing.T, raw string) time.Time {
	t.Helper()

	parsed, err := time.Parse(time.RFC3339, raw)
	require.NoError(t, err)

	return parsed
}

func TestParseISO8601Duration(t *testing.T) {
	t.Run("valid durations", func(t *testing.T) {
		tests := []struct {
			name string
			raw  string
			want utils.ISO8601Duration
		}{
			{
				name: "hour granularity",
				raw:  "PT1H",
				want: utils.ISO8601Duration{Hours: 1},
			},
			{
				name: "week granularity",
				raw:  "P1W",
				want: utils.ISO8601Duration{Weeks: 1},
			},
			{
				name: "calendar month",
				raw:  "P1M",
				want: utils.ISO8601Duration{Months: 1},
			},
			{
				name: "mixed date and time parts",
				raw:  "P2Y3M4DT5H6M7S",
				want: utils.ISO8601Duration{
					Years: 2, Months: 3, Days: 4,
					Hours: 5, Minutes: 6, Seconds: 7,
				},
			},
			{
				name: "zero seconds normalizes to zero value",
				raw:  "PT0S",
				want: utils.ISO8601Duration{},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := utils.ParseISO8601Duration(tt.raw)
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			})
		}
	})

	t.Run("invalid durations", func(t *testing.T) {
		tests := []string{
			"",
			"P",
			"PT",
			"P1Q",
			"P-1D",
			"PT1H30",
			"foo",
		}

		for _, raw := range tests {
			t.Run(raw, func(t *testing.T) {
				_, err := utils.ParseISO8601Duration(raw)
				require.Error(t, err)
			})
		}
	})
}

func TestParseISO8601Interval(t *testing.T) {
	loc := time.UTC
	now := mustRFC3339(t, "2024-01-10T12:00:00Z")

	t.Run("explicit start/end", func(t *testing.T) {
		raw := "2024-01-01T00:00:00Z/2024-01-02T00:00:00Z"
		got, err := utils.ParseISO8601Interval(raw, now, loc)
		require.NoError(t, err)

		wantStart := mustRFC3339(t, "2024-01-01T00:00:00Z")
		wantEnd := mustRFC3339(t, "2024-01-02T00:00:00Z")
		require.True(t, got.Start.Equal(wantStart))
		require.True(t, got.End.Equal(wantEnd))
	})

	t.Run("start plus duration", func(t *testing.T) {
		raw := "2024-01-01T00:00:00Z/P1D"
		got, err := utils.ParseISO8601Interval(raw, now, loc)
		require.NoError(t, err)

		wantStart := mustRFC3339(t, "2024-01-01T00:00:00Z")
		wantEnd := mustRFC3339(t, "2024-01-02T00:00:00Z")
		require.True(t, got.Start.Equal(wantStart))
		require.True(t, got.End.Equal(wantEnd))
	})

	t.Run("duration plus end", func(t *testing.T) {
		raw := "PT6H/2024-01-02T00:00:00Z"
		got, err := utils.ParseISO8601Interval(raw, now, loc)
		require.NoError(t, err)

		wantStart := mustRFC3339(t, "2024-01-01T18:00:00Z")
		wantEnd := mustRFC3339(t, "2024-01-02T00:00:00Z")
		require.True(t, got.Start.Equal(wantStart))
		require.True(t, got.End.Equal(wantEnd))
	})

	t.Run("invalid intervals", func(t *testing.T) {
		tests := []string{
			"",
			"2024-01-01T00:00:00Z",
			"foo/bar",
			"P1D/P2D",
			"2024-01-02T00:00:00Z/2024-01-01T00:00:00Z",
			"2024-01-01T00:00:00Z/2024-01-01T00:00:00Z",
		}

		for _, raw := range tests {
			t.Run(raw, func(t *testing.T) {
				_, err := utils.ParseISO8601Interval(raw, now, loc)
				require.Error(t, err)
			})
		}
	})
}

func TestParseTimezone(t *testing.T) {
	t.Run("valid timezones", func(t *testing.T) {
		tests := []struct {
			name string
			raw  string
			want string
		}{
			{
				name: "UTC",
				raw:  "UTC",
				want: "UTC",
			},
			{
				name: "stockholm",
				raw:  "Europe/Stockholm",
				want: "Europe/Stockholm",
			},
			{
				name: "empty defaults to UTC",
				raw:  "",
				want: "UTC",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := utils.ParseTimezone(tt.raw)
				require.NoError(t, err)
				require.NotNil(t, got)
				require.Equal(t, tt.want, got.String())
			})
		}
	})

	t.Run("invalid timezone", func(t *testing.T) {
		_, err := utils.ParseTimezone("Invalid/Timezone")
		require.Error(t, err)
	})
}

func TestBuildTimeBuckets(t *testing.T) {
	t.Run("hourly buckets split interval", func(t *testing.T) {
		interval := utils.ResolvedInterval{
			Start: mustRFC3339(t, "2024-01-01T00:00:00Z"),
			End:   mustRFC3339(t, "2024-01-01T03:00:00Z"),
		}

		buckets, err := utils.BuildTimeBuckets(interval, utils.ISO8601Duration{Hours: 1}, time.UTC, 10)
		require.NoError(t, err)
		require.Len(t, buckets, 3)

		for i := range 3 {
			wantStart := interval.Start.Add(time.Duration(i) * time.Hour)
			wantEnd := interval.Start.Add(time.Duration(i+1) * time.Hour)
			require.True(t, buckets[i].Start.Equal(wantStart))
			require.True(t, buckets[i].End.Equal(wantEnd))
		}
	})

	t.Run("last bucket is clamped to interval end", func(t *testing.T) {
		interval := utils.ResolvedInterval{
			Start: mustRFC3339(t, "2024-01-01T00:00:00Z"),
			End:   mustRFC3339(t, "2024-01-01T02:30:00Z"),
		}

		buckets, err := utils.BuildTimeBuckets(interval, utils.ISO8601Duration{Hours: 1}, time.UTC, 10)
		require.NoError(t, err)
		require.Len(t, buckets, 3)
		require.True(t, buckets[len(buckets)-1].End.Equal(interval.End))
	})

	t.Run("calendar month buckets", func(t *testing.T) {
		interval := utils.ResolvedInterval{
			Start: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
		}

		buckets, err := utils.BuildTimeBuckets(interval, utils.ISO8601Duration{Months: 1}, time.UTC, 10)
		require.NoError(t, err)
		require.Len(t, buckets, 3)

		want := []utils.TimeBucket{
			{
				Start: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
			},
			{
				Start: time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
			},
			{
				Start: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
			},
		}

		for i := range want {
			require.True(t, buckets[i].Start.Equal(want[i].Start))
			require.True(t, buckets[i].End.Equal(want[i].End))
		}
	})

	t.Run("non-UTC timezone", func(t *testing.T) {
		interval := utils.ResolvedInterval{
			Start: mustRFC3339(t, "2024-01-01T12:00:00Z"),
			End:   mustRFC3339(t, "2024-01-05T12:00:00Z"),
		}

		tz := time.FixedZone("UTC+2", 2*60*60)
		buckets, err := utils.BuildTimeBuckets(interval, utils.ISO8601Duration{Days: 1}, tz, 10)
		require.NoError(t, err)
		require.Len(t, buckets, 5)

		want := []utils.TimeBucket{
			{
				Start: time.Date(2024, 1, 1, 14, 0, 0, 0, tz),
				End:   time.Date(2024, 1, 2, 2, 0, 0, 0, tz),
			},
			{
				Start: time.Date(2024, 1, 2, 2, 0, 0, 0, tz),
				End:   time.Date(2024, 1, 3, 2, 0, 0, 0, tz),
			},
			{
				Start: time.Date(2024, 1, 3, 2, 0, 0, 0, tz),
				End:   time.Date(2024, 1, 4, 2, 0, 0, 0, tz),
			},
			{
				Start: time.Date(2024, 1, 4, 2, 0, 0, 0, tz),
				End:   time.Date(2024, 1, 5, 2, 0, 0, 0, tz),
			},
			{
				Start: time.Date(2024, 1, 5, 2, 0, 0, 0, tz),
				End:   time.Date(2024, 1, 5, 14, 0, 0, 0, tz),
			},
		}

		for i := range want {
			require.True(t, buckets[i].Start.Equal(want[i].Start))
			require.True(t, buckets[i].End.Equal(want[i].End))
		}
	})

	t.Run("invalid bucket inputs", func(t *testing.T) {
		tests := []struct {
			name        string
			interval    utils.ResolvedInterval
			granularity utils.ISO8601Duration
			maxBuckets  int
		}{
			{
				name: "interval start after end",
				interval: utils.ResolvedInterval{
					Start: mustRFC3339(t, "2024-01-02T00:00:00Z"),
					End:   mustRFC3339(t, "2024-01-01T00:00:00Z"),
				},
				granularity: utils.ISO8601Duration{Hours: 1},
				maxBuckets:  10,
			},
			{
				name: "zero granularity",
				interval: utils.ResolvedInterval{
					Start: mustRFC3339(t, "2024-01-01T00:00:00Z"),
					End:   mustRFC3339(t, "2024-01-01T03:00:00Z"),
				},
				granularity: utils.ISO8601Duration{},
				maxBuckets:  10,
			},
			{
				name: "negative granularity component",
				interval: utils.ResolvedInterval{
					Start: mustRFC3339(t, "2024-01-01T00:00:00Z"),
					End:   mustRFC3339(t, "2024-01-01T03:00:00Z"),
				},
				granularity: utils.ISO8601Duration{Hours: -1},
				maxBuckets:  10,
			},
			{
				name: "max buckets exceeded",
				interval: utils.ResolvedInterval{
					Start: mustRFC3339(t, "2024-01-01T00:00:00Z"),
					End:   mustRFC3339(t, "2024-01-01T05:00:00Z"),
				},
				granularity: utils.ISO8601Duration{Hours: 1},
				maxBuckets:  4,
			},
			{
				name: "non-positive max buckets",
				interval: utils.ResolvedInterval{
					Start: mustRFC3339(t, "2024-01-01T00:00:00Z"),
					End:   mustRFC3339(t, "2024-01-01T01:00:00Z"),
				},
				granularity: utils.ISO8601Duration{Hours: 1},
				maxBuckets:  0,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := utils.BuildTimeBuckets(tt.interval, tt.granularity, time.UTC, tt.maxBuckets)
				require.Error(t, err)
			})
		}
	})
}
