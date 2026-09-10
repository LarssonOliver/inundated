package logging_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/larssonoliver/inundated/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDefaultWriterDoesNotBlockWhenStdoutStopsDraining(t *testing.T) {
	// Production logs to os.Stdout. If that is a pipe to a log collector that
	// stalls, logging must not stall with it -- otherwise every goroutine that
	// logs (every request handler) blocks behind slog's single handler mutex.
	pr, pw, err := os.Pipe()
	require.NoError(t, err)
	t.Cleanup(func() { _ = pr.Close(); _ = pw.Close() })

	orig := os.Stdout
	os.Stdout = pw
	t.Cleanup(func() { os.Stdout = orig })

	logger, cleanup, err := logging.New(logging.Options{Format: "json"})
	require.NoError(t, err)
	t.Cleanup(cleanup)

	done := make(chan struct{})
	go func() {
		// Far more than the OS pipe buffer holds, so a synchronous writer would
		// wedge partway through.
		for i := 0; i < 20_000; i++ {
			logger.Info("filling an unread pipe", "i", i)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("logging.New's default logger blocked when stdout stopped draining")
	}
}

func TestNewFiltersBelowConfiguredLevel(t *testing.T) {
	var buf bytes.Buffer
	l, cleanup, err := logging.New(logging.Options{Level: "warn", Format: "json", Writer: &buf})
	require.NoError(t, err)
	defer cleanup()

	l.Info("should be hidden")
	l.Warn("should be shown")

	out := buf.String()
	assert.NotContains(t, out, "should be hidden")
	assert.Contains(t, out, "should be shown")
}

func TestNewDefaultsToTextFormat(t *testing.T) {
	var buf bytes.Buffer
	l, cleanup, err := logging.New(logging.Options{Writer: &buf})
	require.NoError(t, err)
	defer cleanup()

	l.Info("hello", "key", "value")

	assert.Contains(t, buf.String(), "key=value")
}

func TestNewJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	l, cleanup, err := logging.New(logging.Options{Format: "json", Writer: &buf})
	require.NoError(t, err)
	defer cleanup()

	l.Info("hello", "key", "value")

	var rec map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &rec))
	assert.Equal(t, "hello", rec["msg"])
	assert.Equal(t, "value", rec["key"])
}

func TestNewRejectsUnknownLevel(t *testing.T) {
	_, _, err := logging.New(logging.Options{Level: "verbose"})
	assert.Error(t, err)
}

func TestNewRejectsUnknownFormat(t *testing.T) {
	_, _, err := logging.New(logging.Options{Format: "yaml"})
	assert.Error(t, err)
}

func TestContextAttrsAppearInEveryRecord(t *testing.T) {
	var buf bytes.Buffer
	l, cleanup, err := logging.New(logging.Options{Format: "json", Writer: &buf})
	require.NoError(t, err)
	defer cleanup()

	ctx := logging.ContextWith(context.Background(), slog.String("request_id", "abc123"))
	l.InfoContext(ctx, "hello")

	var rec map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &rec))
	assert.Equal(t, "abc123", rec["request_id"])
}

func TestContextWithAccumulates(t *testing.T) {
	var buf bytes.Buffer
	l, cleanup, err := logging.New(logging.Options{Format: "json", Writer: &buf})
	require.NoError(t, err)
	defer cleanup()

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
	l, cleanup, err := logging.New(logging.Options{Format: "json", Writer: &buf})
	require.NoError(t, err)
	defer cleanup()

	ctx := logging.ContextWith(context.Background(), slog.String("request_id", "abc"))
	l.With("component", "test").InfoContext(ctx, "hello")

	var rec map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &rec))
	assert.Equal(t, "abc", rec["request_id"])
	assert.Equal(t, "test", rec["component"])
}
