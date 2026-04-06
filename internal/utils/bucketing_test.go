package utils

import (
	"testing"
	"time"
)

func TestBucketDuration(t *testing.T) {
	bucket := Bucket{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	got := bucket.Duration()
	want := 24 * time.Hour

	if got != want {
		t.Errorf("Bucket.Duration() = %v, want %v", got, want)
	}
}

func TestBucketOverlaps(t *testing.T) {
	bucket := Bucket{
		Start: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name  string
		start time.Time
		end   time.Time
		want  bool
	}{
		{
			name:  "fully contains",
			start: time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
			want:  true,
		},
		{
			name:  "overlaps start",
			start: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC),
			want:  true,
		},
		{
			name:  "overlaps end",
			start: time.Date(2024, 1, 1, 16, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 1, 1, 20, 0, 0, 0, time.UTC),
			want:  true,
		},
		{
			name:  "completely before",
			start: time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			want:  false,
		},
		{
			name:  "completely after",
			start: time.Date(2024, 1, 1, 20, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 1, 1, 22, 0, 0, 0, time.UTC),
			want:  false,
		},
		{
			name:  "exact match",
			start: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC),
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bucket.Overlaps(tt.start, tt.end)
			if got != tt.want {
				t.Errorf("Bucket.Overlaps() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBucketOverlapDuration(t *testing.T) {
	bucket := Bucket{
		Start: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name  string
		start time.Time
		end   time.Time
		want  time.Duration
	}{
		{
			name:  "fully within bucket",
			start: time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
			want:  2 * time.Hour,
		},
		{
			name:  "overlaps start",
			start: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC),
			want:  2 * time.Hour, // 12:00-14:00
		},
		{
			name:  "overlaps end",
			start: time.Date(2024, 1, 1, 16, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 1, 1, 20, 0, 0, 0, time.UTC),
			want:  2 * time.Hour, // 16:00-18:00
		},
		{
			name:  "no overlap (before)",
			start: time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			want:  0,
		},
		{
			name:  "no overlap (after)",
			start: time.Date(2024, 1, 1, 20, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 1, 1, 22, 0, 0, 0, time.UTC),
			want:  0,
		},
		{
			name:  "span contains bucket",
			start: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 1, 1, 20, 0, 0, 0, time.UTC),
			want:  6 * time.Hour, // entire bucket
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bucket.OverlapDuration(tt.start, tt.end)
			if got != tt.want {
				t.Errorf("Bucket.OverlapDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateBuckets_FixedDuration(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 1, 6, 0, 0, 0, time.UTC)
	interval := Interval{Start: start, End: end}

	tests := []struct {
		name        string
		granularity string
		wantCount   int
		firstBucket Bucket
	}{
		{
			name:        "hourly buckets",
			granularity: "PT1H",
			wantCount:   6,
			firstBucket: Bucket{
				Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC),
			},
		},
		{
			name:        "30-minute buckets",
			granularity: "PT30M",
			wantCount:   12,
			firstBucket: Bucket{
				Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, 1, 1, 0, 30, 0, 0, time.UTC),
			},
		},
		{
			name:        "2-hour buckets",
			granularity: "PT2H",
			wantCount:   3,
			firstBucket: Bucket{
				Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, 1, 1, 2, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buckets, err := GenerateBuckets(interval, tt.granularity, time.UTC)
			if err != nil {
				t.Fatalf("GenerateBuckets() error = %v", err)
			}

			if len(buckets) != tt.wantCount {
				t.Errorf("GenerateBuckets() bucket count = %v, want %v", len(buckets), tt.wantCount)
			}

			if len(buckets) > 0 {
				if !buckets[0].Start.Equal(tt.firstBucket.Start) || !buckets[0].End.Equal(tt.firstBucket.End) {
					t.Errorf("First bucket = {%v, %v}, want {%v, %v}",
						buckets[0].Start, buckets[0].End, tt.firstBucket.Start, tt.firstBucket.End)
				}
			}
		})
	}
}

func TestGenerateBuckets_DailyWithTimezone(t *testing.T) {
	stockholm, _ := time.LoadLocation("Europe/Stockholm")

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, stockholm)
	end := time.Date(2024, 1, 4, 0, 0, 0, 0, stockholm)
	interval := Interval{Start: start, End: end}

	buckets, err := GenerateBuckets(interval, "P1D", stockholm)
	if err != nil {
		t.Fatalf("GenerateBuckets() error = %v", err)
	}

	// Should have 3 full days
	if len(buckets) != 3 {
		t.Errorf("GenerateBuckets() bucket count = %v, want 3", len(buckets))
	}

	// First bucket should start at midnight Stockholm time on Jan 1
	wantStart := time.Date(2024, 1, 1, 0, 0, 0, 0, stockholm)
	if !buckets[0].Start.Equal(wantStart) {
		t.Errorf("First bucket start = %v, want %v", buckets[0].Start, wantStart)
	}

	// Each bucket should be exactly 24 hours
	for i, bucket := range buckets {
		duration := bucket.End.Sub(bucket.Start)
		if duration != 24*time.Hour {
			t.Errorf("Bucket %d duration = %v, want 24h", i, duration)
		}
	}
}

func TestGenerateBuckets_Weekly(t *testing.T) {
	// Start on a Wednesday
	start := time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC) // Wednesday, Jan 3
	end := time.Date(2024, 1, 24, 12, 0, 0, 0, time.UTC)  // Wednesday, Jan 24
	interval := Interval{Start: start, End: end}

	buckets, err := GenerateBuckets(interval, "P1W", time.UTC)
	if err != nil {
		t.Fatalf("GenerateBuckets() error = %v", err)
	}

	// Should have 3 weeks: Jan 8-15, Jan 15-22, Jan 22-29
	if len(buckets) != 3 {
		t.Errorf("GenerateBuckets() bucket count = %v, want 3", len(buckets))
	}

	// First bucket should start on Monday, Jan 8
	wantStart := time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)
	if !buckets[0].Start.Equal(wantStart) {
		t.Errorf("First bucket start = %v (weekday %v), want %v (Monday)",
			buckets[0].Start, buckets[0].Start.Weekday(), wantStart)
	}

	// First two buckets should be full weeks (7 days)
	// Last bucket might be truncated to the interval end
	for i := 0; i < 2; i++ {
		duration := buckets[i].End.Sub(buckets[i].Start)
		expectedDuration := 7 * 24 * time.Hour
		if duration != expectedDuration {
			t.Errorf("Bucket %d duration = %v, want %v", i, duration, expectedDuration)
		}
	}

	// Last bucket should end at the interval end (Jan 24 12:00)
	if !buckets[2].End.Equal(end) {
		t.Errorf("Last bucket end = %v, want %v", buckets[2].End, end)
	}
}

func TestGenerateBuckets_Monthly(t *testing.T) {
	start := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	end := time.Date(2024, 4, 15, 12, 0, 0, 0, time.UTC)
	interval := Interval{Start: start, End: end}

	buckets, err := GenerateBuckets(interval, "P1M", time.UTC)
	if err != nil {
		t.Fatalf("GenerateBuckets() error = %v", err)
	}

	// Should have 3 months: Feb, Mar, Apr
	if len(buckets) != 3 {
		t.Errorf("GenerateBuckets() bucket count = %v, want 3", len(buckets))
	}

	// First bucket should start on Feb 1
	wantStart := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	if !buckets[0].Start.Equal(wantStart) {
		t.Errorf("First bucket start = %v, want %v", buckets[0].Start, wantStart)
	}

	// Check that buckets start on the 1st of each month
	expectedStarts := []time.Time{
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
	}

	for i, bucket := range buckets {
		if !bucket.Start.Equal(expectedStarts[i]) {
			t.Errorf("Bucket %d start = %v, want %v", i, bucket.Start, expectedStarts[i])
		}
	}
}

func TestSplitDurationAcrossBuckets(t *testing.T) {
	// Create buckets: 12:00-13:00, 13:00-14:00, 14:00-15:00
	buckets := []Bucket{
		{
			Start: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
		},
		{
			Start: time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC),
		},
		{
			Start: time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
		},
	}

	tests := []struct {
		name          string
		spanStart     time.Time
		spanEnd       time.Time
		totalDuration time.Duration
		want          map[int]time.Duration
	}{
		{
			name:          "span within single bucket",
			spanStart:     time.Date(2024, 1, 1, 12, 15, 0, 0, time.UTC),
			spanEnd:       time.Date(2024, 1, 1, 12, 45, 0, 0, time.UTC),
			totalDuration: 30 * time.Minute,
			want: map[int]time.Duration{
				0: 30 * time.Minute, // all 30 minutes in bucket 0
			},
		},
		{
			name:          "span across two buckets evenly",
			spanStart:     time.Date(2024, 1, 1, 12, 30, 0, 0, time.UTC),
			spanEnd:       time.Date(2024, 1, 1, 13, 30, 0, 0, time.UTC),
			totalDuration: 60 * time.Minute,
			want: map[int]time.Duration{
				0: 30 * time.Minute, // 50% of duration
				1: 30 * time.Minute, // 50% of duration
			},
		},
		{
			name:          "span across all three buckets",
			spanStart:     time.Date(2024, 1, 1, 12, 20, 0, 0, time.UTC),
			spanEnd:       time.Date(2024, 1, 1, 14, 40, 0, 0, time.UTC),
			totalDuration: 140 * time.Minute,
			want: map[int]time.Duration{
				0: 40 * time.Minute,  // 40/140 of total
				1: 60 * time.Minute,  // 60/140 of total
				2: 40 * time.Minute,  // 40/140 of total
			},
		},
		{
			name:          "span with no overlap",
			spanStart:     time.Date(2024, 1, 1, 16, 0, 0, 0, time.UTC),
			spanEnd:       time.Date(2024, 1, 1, 17, 0, 0, 0, time.UTC),
			totalDuration: 60 * time.Minute,
			want:          map[int]time.Duration{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SplitDurationAcrossBuckets(buckets, tt.spanStart, tt.spanEnd, tt.totalDuration)

			if len(got) != len(tt.want) {
				t.Errorf("SplitDurationAcrossBuckets() result count = %v, want %v", len(got), len(tt.want))
			}

			for bucketIdx, wantDuration := range tt.want {
				gotDuration, ok := got[bucketIdx]
				if !ok {
					t.Errorf("Missing bucket %d in result", bucketIdx)
					continue
				}

				// Allow small rounding differences (within 1 second)
				diff := gotDuration - wantDuration
				if diff < 0 {
					diff = -diff
				}
				if diff > time.Second {
					t.Errorf("Bucket %d duration = %v, want %v (diff %v)", bucketIdx, gotDuration, wantDuration, diff)
				}
			}
		})
	}
}
