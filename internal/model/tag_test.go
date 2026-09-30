package model

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestTaskTagColor(t *testing.T) {
	low := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	high := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	tests := []struct {
		name   string
		labels []Tag
		want   string
	}{
		{"no labels", nil, DefaultTaskTagColor},
		{"first by name ignoring case", []Tag{{Name: "banana", Color: "#222222"}, {Name: "Apple", Color: "#111111"}}, "#111111"},
		{"same name folded, then by bytes", []Tag{{Name: "apple", Color: "#222222"}, {Name: "Apple", Color: "#111111"}}, "#111111"},
		{"same name, then by id", []Tag{{Id: high, Name: "a", Color: "#222222"}, {Id: low, Name: "a", Color: "#111111"}}, "#111111"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, TaskTagColor(tt.labels))
		})
	}
}
