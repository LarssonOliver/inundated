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
