package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

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
	}{
		{name: "success"},
		{name: "not found", repoErr: model.ErrNotFound},
		{name: "logged time", repoErr: model.ErrConflict},
		{name: "repository error", repoErr: errors.New("database error")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repository.RepoMock{
				DeleteTaskFn: func(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error {
					return tt.repoErr
				},
			}
			err := service.NewService(repo).DeleteTask(context.Background(), uuid.New())
			if tt.repoErr != nil {
				require.ErrorIs(t, err, tt.repoErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestTaskService_UpdateTaskRejectsBadPatches(t *testing.T) {
	bogus := model.CloseReason("bogus")
	tests := map[string]model.TaskPatch{
		"negative estimate":      {Estimate: new(-time.Hour)},
		"unknown close reason":   {CloseReason: &bogus},
		"reason while opening":   {CloseReason: new(model.CloseReasonDone), Closed: new(false)},
		"set and clear due date": {DueDate: new(time.Now()), ClearDueDate: true},
		"set and clear estimate": {Estimate: new(time.Hour), ClearEstimate: true},
	}
	for name, patch := range tests {
		t.Run(name, func(t *testing.T) {
			repo := &repository.RepoMock{
				UpdateTaskFn: func(context.Context, model.OwnerScope, uuid.UUID, model.TaskPatch) (model.Task, error) {
					t.Fatal("a bad patch must not reach the repository")
					return model.Task{}, nil
				},
			}
			_, err := service.NewService(repo).UpdateTask(context.Background(), uuid.New(), patch)
			require.ErrorIs(t, err, model.ErrInvalidArgument)
		})
	}
}
