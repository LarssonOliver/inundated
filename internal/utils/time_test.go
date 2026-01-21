package utils_test

import (
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.FloatHoursToDuration(tt.hours)
			if got == nil && tt.want == nil {
				return
			}
			if *got != *tt.want {
				t.Errorf("FloatHoursToDuration() = %v, want %v", got, tt.want)
			}
		})
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
