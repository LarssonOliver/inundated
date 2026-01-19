package helpers_test

import (
	"github.com/larssonoliver/inundated/internal/helpers"
	"testing"
)

func TestIsValidColor(t *testing.T) {
	tests := []struct {
		name  string
		color string
		want  bool
	}{
		{
			name:  "valid 6-digit hex color",
			color: "#1A2B3C",
			want:  true,
		},
		{
			name:  "valid 3-digit hex color",
			color: "#ABC",
			want:  true,
		},
		{
			name:  "invalid color missing #",
			color: "123456",
			want:  false,
		},
		{
			name:  "invalid color too short",
			color: "#12",
			want:  false,
		},
		{
			name:  "invalid color too long",
			color: "#1234567",
			want:  false,
		},
		{
			name:  "invalid color with non-hex characters",
			color: "#12G45Z",
			want:  false,
		},
		{
			name:  "empty string",
			color: "",
			want:  false,
		},
		{
			name: "valid lowercase hex color",
			color: "#abcdef",
			want:  true,
		},
		{
			name: "valid uppercase hex color",
			color: "#ABCDEF",
			want:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := helpers.IsValidColor(tt.color)
			if got != tt.want {
				t.Errorf("IsValidColor() = %v, want %v", got, tt.want)
			}
		})
	}
}
