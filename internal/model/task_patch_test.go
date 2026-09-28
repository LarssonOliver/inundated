package model_test

import (
	"testing"
	"time"

	"github.com/larssonoliver/inundated/internal/model"
	"github.com/stretchr/testify/require"
)

func TestTaskPatch_Apply(t *testing.T) {
	ignored := model.CloseReasonIgnored
	due := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	estimate := time.Hour
	open := model.Task{Name: "t", DueDate: &due, Estimate: &estimate}
	closed := open
	closed.CloseReason = &ignored

	t.Run("empty patch keeps everything", func(t *testing.T) {
		require.Equal(t, closed, model.TaskPatch{}.Apply(closed))
	})
	t.Run("clears fields", func(t *testing.T) {
		got := model.TaskPatch{ClearDueDate: true, ClearEstimate: true}.Apply(open)
		require.Nil(t, got.DueDate)
		require.Nil(t, got.Estimate)
		require.Equal(t, "t", got.Name)
	})
	t.Run("closing an open task defaults to done", func(t *testing.T) {
		got := model.TaskPatch{Closed: new(true)}.Apply(open)
		require.Equal(t, model.CloseReasonDone, *got.CloseReason)
	})
	t.Run("closing a closed task keeps its reason", func(t *testing.T) {
		got := model.TaskPatch{Closed: new(true)}.Apply(closed)
		require.Equal(t, model.CloseReasonIgnored, *got.CloseReason)
	})
	t.Run("a reason alone closes or changes the reason", func(t *testing.T) {
		got := model.TaskPatch{CloseReason: new(model.CloseReasonDone)}.Apply(closed)
		require.Equal(t, model.CloseReasonDone, *got.CloseReason)
		require.Equal(t, model.CloseReasonIgnored, *closed.CloseReason, "Apply must not alias the input")
	})
	t.Run("reopening", func(t *testing.T) {
		require.False(t, model.TaskPatch{Closed: new(false)}.Apply(closed).Closed())
	})
}
