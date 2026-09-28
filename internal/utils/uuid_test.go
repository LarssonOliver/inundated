package utils_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/utils"
	"github.com/stretchr/testify/require"
)

func TestDedupeUUIDs(t *testing.T) {
	a := uuid.New()
	b := uuid.New()

	require.Nil(t, utils.DedupeUUIDs(nil))
	require.Equal(t, []uuid.UUID{a}, utils.DedupeUUIDs([]uuid.UUID{a}))
	require.Equal(t, []uuid.UUID{a}, utils.DedupeUUIDs([]uuid.UUID{a, a}))
	require.Equal(t, []uuid.UUID{a, b}, utils.DedupeUUIDs([]uuid.UUID{a, b, a, b}))
}

func TestSameUUIDSet(t *testing.T) {
	a, b, c := uuid.New(), uuid.New(), uuid.New()
	tests := []struct {
		name string
		x, y []uuid.UUID
		want bool
	}{
		{"both empty", nil, []uuid.UUID{}, true},
		{"same order", []uuid.UUID{a, b}, []uuid.UUID{a, b}, true},
		{"other order", []uuid.UUID{a, b}, []uuid.UUID{b, a}, true},
		{"duplicates ignored", []uuid.UUID{a, a, b}, []uuid.UUID{b, a}, true},
		{"missing one", []uuid.UUID{a, b}, []uuid.UUID{a}, false},
		{"extra one", []uuid.UUID{a}, []uuid.UUID{a, b}, false},
		{"different", []uuid.UUID{a, b}, []uuid.UUID{a, c}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.SameUUIDSet(tt.x, tt.y); got != tt.want {
				t.Errorf("SameUUIDSet(%v, %v) = %v, want %v", tt.x, tt.y, got, tt.want)
			}
		})
	}
}
