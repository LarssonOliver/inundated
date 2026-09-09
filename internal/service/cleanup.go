package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/larssonoliver/inundated/internal/repository"
)

type CleanupService interface {
	Run(ctx context.Context)
}

type CleanupServiceImpl struct {
	sessionRepository    repository.SessionRepository
	loginStateRepository repository.LoginStateRepository
	interval             time.Duration
}

var _ CleanupService = (*CleanupServiceImpl)(nil)

func NewCleanupService(sessionRepository repository.SessionRepository, loginStateRepository repository.LoginStateRepository, interval time.Duration) *CleanupServiceImpl {
	return &CleanupServiceImpl{
		sessionRepository:    sessionRepository,
		loginStateRepository: loginStateRepository,
		interval:             interval,
	}
}

// Run implements [CleanupService].
func (c *CleanupServiceImpl) Run(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	c.cleanup(ctx)

	for {
		select {
		case <-ticker.C:
			c.cleanup(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (c *CleanupServiceImpl) cleanup(ctx context.Context) {
	if err := c.sessionRepository.DeleteAllExpiredSessions(ctx); err != nil {
		slog.ErrorContext(ctx, "cleanup: deleting expired sessions failed", "error", err)
	}
	if err := c.loginStateRepository.DeleteAllExpiredLoginStates(ctx); err != nil {
		slog.ErrorContext(ctx, "cleanup: deleting expired login states failed", "error", err)
	}
	slog.DebugContext(ctx, "cleanup run complete")
}
