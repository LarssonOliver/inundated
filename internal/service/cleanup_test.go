package service_test

import (
	"bytes"
	"context"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCleanupService_LogsSweepErrorsAndKeepsGoing(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

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
}

func TestCleanupService_QuietWhenSweepsSucceed(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	sessions := &repository.SessionRepoMock{
		DeleteAllExpiredSessionsFn: func(context.Context) error { return nil },
	}
	loginStates := &repository.LoginStateRepoMock{
		DeleteAllExpiredLoginStatesFn: func(context.Context) error { return nil },
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	service.NewCleanupService(sessions, loginStates, time.Hour).Run(ctx)

	assert.Empty(t, buf.String(), "a clean sweep must not log anything")
}
