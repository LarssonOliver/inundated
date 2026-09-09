package service_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureInfoLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return &buf
}

func TestCleanupService_LogsSweepErrorsAndKeepsGoing(t *testing.T) {
	buf := captureInfoLogs(t)

	var sessionsCalled, loginStatesCalled bool
	sessions := &repository.SessionRepoMock{
		DeleteAllExpiredSessionsFn: func(context.Context) error {
			sessionsCalled = true
			return errors.New("sessions sweep failed")
		},
	}
	loginStates := &repository.LoginStateRepoMock{
		DeleteAllExpiredLoginStatesFn: func(context.Context) error {
			loginStatesCalled = true
			return errors.New("login-states sweep failed")
		},
	}

	// A cancelled context makes Run do exactly one sweep pass, then return.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	service.NewCleanupService(sessions, loginStates, time.Hour).Run(ctx)

	assert.True(t, sessionsCalled, "the session sweep must run")
	assert.True(t, loginStatesCalled, "a failed session sweep must not skip the login-state sweep")

	out := buf.String()
	require.Contains(t, out, "sessions sweep failed")
	require.Contains(t, out, "login-states sweep failed")
	require.Contains(t, out, `"level":"ERROR"`)
}

func TestCleanupService_QuietWhenSweepsSucceed(t *testing.T) {
	buf := captureInfoLogs(t)

	sessions := &repository.SessionRepoMock{
		DeleteAllExpiredSessionsFn: func(context.Context) error { return nil },
	}
	loginStates := &repository.LoginStateRepoMock{
		DeleteAllExpiredLoginStatesFn: func(context.Context) error { return nil },
	}

	// A cancelled context makes Run do exactly one sweep pass, then return.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	service.NewCleanupService(sessions, loginStates, time.Hour).Run(ctx)

	assert.Empty(t, buf.String(), "a clean sweep must not log at info level or above")
}

func TestCleanupService_EmitsDebugOnCompletedPass(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(prev) })

	sessions := &repository.SessionRepoMock{
		DeleteAllExpiredSessionsFn: func(context.Context) error { return nil },
	}
	loginStates := &repository.LoginStateRepoMock{
		DeleteAllExpiredLoginStatesFn: func(context.Context) error { return nil },
	}

	// A cancelled context makes Run do exactly one sweep pass, then return.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	service.NewCleanupService(sessions, loginStates, time.Hour).Run(ctx)

	var rec map[string]any
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &rec))
	assert.Equal(t, "cleanup run complete", rec["msg"])
	assert.Equal(t, "DEBUG", rec["level"])
}
