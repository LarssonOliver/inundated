package model_test

import (
	"testing"

	"github.com/larssonoliver/inundated/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashToken(t *testing.T) {
	t.Run("is deterministic", func(t *testing.T) {
		assert.Equal(t, model.HashToken("abc"), model.HashToken("abc"))
	})

	t.Run("is a 32-byte digest that hides the input", func(t *testing.T) {
		h := model.HashToken("a-secret-session-token")
		require.Len(t, h, 32)
		assert.NotContains(t, string(h), "a-secret-session-token")
	})

	t.Run("differs per input", func(t *testing.T) {
		assert.NotEqual(t, model.HashToken("token-a"), model.HashToken("token-b"))
	})
}
