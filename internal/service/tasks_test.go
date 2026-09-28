package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/service"
	"github.com/stretchr/testify/require"
)

func TestTaskService_DeleteTask(t *testing.T) {
	tests := []struct {
		name    string
		repoErr error
		wantErr error
	}{
		{name: "success"},
		{name: "not found", repoErr: model.ErrNotFound, wantErr: model.ErrNotFound},
		{name: "logged time", repoErr: model.ErrConflict, wantErr: model.ErrConflict},
		{
			// The repository rejects the zero UUID as malformed; no task
			// can have it, so it's a miss rather than a server error.
			name:    "malformed id is not found",
			repoErr: fmt.Errorf("DeleteTask: id: %w", model.ErrInvalidArgument),
			wantErr: model.ErrNotFound,
		},
		{name: "repository error", repoErr: errors.New("database error")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repository.RepoMock{
				DeleteTaskFn: func(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error {
					return tt.repoErr
				},
			}
			err := service.NewService(repo).DeleteTask(context.Background(), uuid.Nil)
			switch {
			case tt.wantErr != nil:
				require.ErrorIs(t, err, tt.wantErr)
			case tt.repoErr != nil:
				require.ErrorIs(t, err, tt.repoErr)
				require.NotErrorIs(t, err, model.ErrNotFound)
			default:
				require.NoError(t, err)
			}
		})
	}
}
