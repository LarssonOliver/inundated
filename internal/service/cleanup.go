package service

import (
	"context"
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
	_ = c.sessionRepository.DeleteAllExpiredSessions(ctx)
	_ = c.loginStateRepository.DeleteAllExpiredLoginStates(ctx)
}
