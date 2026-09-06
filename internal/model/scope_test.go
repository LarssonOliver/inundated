package model_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/stretchr/testify/require"
)

func TestUserScope_CarriesID(t *testing.T) {
	id := uuid.New()

	got := model.UserScope(id).UserID()

	require.NotNil(t, got)
	require.Equal(t, id, *got)
}

func TestUnownedScope_HasNoID(t *testing.T) {
	require.Nil(t, model.UnownedScope().UserID())
}

func TestOwnerScope_ZeroValueIsUnowned(t *testing.T) {
	require.Nil(t, model.OwnerScope{}.UserID())
}

func TestUserScope_DoesNotAliasCallerVariable(t *testing.T) {
	id := uuid.New()
	scope := model.UserScope(id)

	// mutating the original variable must not change the scope
	id = uuid.New()

	require.NotEqual(t, id, *scope.UserID())
}
