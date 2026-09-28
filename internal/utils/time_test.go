package utils_test

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/larssonoliver/inundated/internal/utils"
)

func ptrf(f float64) *float64 {
	return &f
}

func ptrd(d time.Duration) *time.Duration {
	return &d
}

func TestFloatHoursToDuration(t *testing.T) {
	tests := []struct {
		name  string
		hours *float64
		want  *time.Duration
	}{
		{
			name:  "Convert 1.5 hours to duration",
			hours: ptrf(1.5),
			want:  ptrd(time.Duration(1*time.Hour + 30*time.Minute)),
		},
		{
			name:  "Convert 0 hours to duration",
			hours: ptrf(0),
			want:  ptrd(time.Duration(0)),
		},
		{
			name:  "Convert 2.25 hours to duration",
			hours: ptrf(2.25),
			want:  ptrd(time.Duration(2*time.Hour + 15*time.Minute)),
		},
		{
			name:  "Convert 0.1 hours to duration",
			hours: ptrf(0.1),
			want:  ptrd(time.Duration(6 * time.Minute)),
		},
		{
			name:  "Negative hours",
			hours: ptrf(-1.5),
			want:  ptrd(time.Duration(-1*time.Hour - 30*time.Minute)),
		},
		{
			name:  "nil hours",
			hours: nil,
			want:  nil,
		},
		{
			name:  "Largest whole hours that fit",
			hours: ptrf(2562047),
			want:  ptrd(2562047 * time.Hour),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := utils.FloatHoursToDuration(tt.hours)
			if err != nil {
				t.Fatalf("FloatHoursToDuration() error = %v", err)
			}
			if got == nil && tt.want == nil {
				return
			}
			if *got != *tt.want {
				t.Errorf("FloatHoursToDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFloatHoursToDuration_OutOfRange(t *testing.T) {
	for _, hours := range []float64{1e7, -1e7, 1e300, math.Inf(1), math.NaN()} {
		if _, err := utils.FloatHoursToDuration(&hours); !errors.Is(err, utils.ErrHoursOutOfRange) {
			t.Errorf("FloatHoursToDuration(%v) error = %v, want ErrHoursOutOfRange", hours, err)
		}
	}
}

func TestDurationToFloatHours(t *testing.T) {
	tests := []struct {
		name string
		d    *time.Duration
		want *float64
	}{
		{
			name: "Convert 1 hour 30 minutes to float hours",
			d:    ptrd(time.Duration(1*time.Hour + 30*time.Minute)),
			want: ptrf(1.5),
		},
		{
			name: "Convert 0 duration to float hours",
			d:    ptrd(time.Duration(0)),
			want: ptrf(0),
		},
		{
			name: "Convert 2 hours 15 minutes to float hours",
			d:    ptrd(time.Duration(2*time.Hour + 15*time.Minute)),
			want: ptrf(2.25),
		},
		{
			name: "nil duration",
			d:    nil,
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.DurationToFloatHours(tt.d)
			if got == nil && tt.want == nil {
				return
			}
			if *got != *tt.want {
				t.Errorf("DurationToFloatHours() = %v, want %v", got, tt.want)
			}
		})
	}
}
