package logging_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/larssonoliver/inundated/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFiltersBelowConfiguredLevel(t *testing.T) {
	var buf bytes.Buffer
	l, err := logging.New(logging.Options{Level: "warn", Format: "json", Writer: &buf})
	require.NoError(t, err)

	l.Info("should be hidden")
	l.Warn("should be shown")

	out := buf.String()
	assert.NotContains(t, out, "should be hidden")
	assert.Contains(t, out, "should be shown")
}

func TestNewDefaultsToTextFormat(t *testing.T) {
	var buf bytes.Buffer
	l, err := logging.New(logging.Options{Writer: &buf})
	require.NoError(t, err)

	l.Info("hello", "key", "value")

	assert.Contains(t, buf.String(), "key=value")
}

func TestNewJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	l, err := logging.New(logging.Options{Format: "json", Writer: &buf})
	require.NoError(t, err)

	l.Info("hello", "key", "value")

	var rec map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &rec))
	assert.Equal(t, "hello", rec["msg"])
	assert.Equal(t, "value", rec["key"])
}

func TestNewRejectsUnknownLevel(t *testing.T) {
	_, err := logging.New(logging.Options{Level: "verbose"})
	assert.Error(t, err)
}

func TestNewRejectsUnknownFormat(t *testing.T) {
	_, err := logging.New(logging.Options{Format: "yaml"})
	assert.Error(t, err)
}

func TestContextAttrsAppearInEveryRecord(t *testing.T) {
	var buf bytes.Buffer
	l, err := logging.New(logging.Options{Format: "json", Writer: &buf})
	require.NoError(t, err)

	ctx := logging.ContextWith(context.Background(), slog.String("request_id", "abc123"))
	l.InfoContext(ctx, "hello")

	var rec map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &rec))
	assert.Equal(t, "abc123", rec["request_id"])
}

func TestContextWithAccumulates(t *testing.T) {
	var buf bytes.Buffer
	l, err := logging.New(logging.Options{Format: "json", Writer: &buf})
	require.NoError(t, err)

	ctx := logging.ContextWith(context.Background(), slog.String("request_id", "abc"))
	ctx = logging.ContextWith(ctx, slog.String("user_id", "u-1"))
	l.InfoContext(ctx, "hello")

	var rec map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &rec))
	assert.Equal(t, "abc", rec["request_id"])
	assert.Equal(t, "u-1", rec["user_id"])
}

func TestContextAttrsSurviveLoggerWith(t *testing.T) {
	var buf bytes.Buffer
	l, err := logging.New(logging.Options{Format: "json", Writer: &buf})
	require.NoError(t, err)

	ctx := logging.ContextWith(context.Background(), slog.String("request_id", "abc"))
	l.With("component", "test").InfoContext(ctx, "hello")

	var rec map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &rec))
	assert.Equal(t, "abc", rec["request_id"])
	assert.Equal(t, "test", rec["component"])
}
