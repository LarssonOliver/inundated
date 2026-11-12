package service

import (
	"context"
	"crypto/rand"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

const (
	appPasswordTouchInterval = 5 * time.Minute
	appPasswordMaxNameLength = 100
)

func (s *ServiceImpl) ListAppPasswords(ctx context.Context) ([]model.AppPassword, error) {
	scope, err := ownerScope(ctx)
	if err != nil {
		return nil, err
	}
	return s.repository.ListAppPasswords(ctx, scope)
}

func (s *ServiceImpl) CreateAppPassword(ctx context.Context, name string) (model.AppPassword, string, error) {
	scope, err := ownerScope(ctx)
	if err != nil {
		return model.AppPassword{}, "", err
	}
	name = strings.TrimSpace(name)
	if length := utf8.RuneCountInString(name); length == 0 || length > appPasswordMaxNameLength {
		return model.AppPassword{}, "", model.ErrInvalidArgument
	}

	token := rand.Text()
	password, err := s.repository.CreateAppPassword(ctx, scope, model.AppPassword{
		Id:        uuid.New(),
		Name:      name,
		CreatedAt: time.Now().UTC(),
	}, token)
	if err != nil {
		return model.AppPassword{}, "", err
	}
	return password, token, nil
}

func (s *ServiceImpl) DeleteAppPassword(ctx context.Context, id uuid.UUID) error {
	scope, err := ownerScope(ctx)
	if err != nil {
		return err
	}
	return s.repository.DeleteAppPassword(ctx, scope, id)
}

func (s *ServiceImpl) AuthenticateAppPassword(ctx context.Context, token string) (model.AppPassword, error) {
	password, err := s.repository.GetAppPasswordByToken(ctx, token)
	if err != nil {
		return model.AppPassword{}, err
	}

	now := time.Now().UTC()
	if password.LastUsedAt == nil || now.Sub(*password.LastUsedAt) >= appPasswordTouchInterval {
		if err := s.repository.TouchAppPassword(ctx, password.Id, now); err != nil {
			slog.WarnContext(ctx, "recording app password use failed", "error", err, "app_password_id", password.Id)
		}
	}
	return password, nil
}
