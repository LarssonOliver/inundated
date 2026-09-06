package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/service"
	"github.com/stretchr/testify/require"
)

func TestService_ThreadsUserScopeFromContext(t *testing.T) {
	var got model.OwnerScope
	repo := &repository.RepoMock{
		ListTagFn: func(ctx context.Context, scope model.OwnerScope, params model.PaginationParams) (model.Page[model.Tag], error) {
			got = scope
			return model.Page[model.Tag]{}, nil
		},
	}
	s := service.NewService(repo)

	user := model.User{Id: uuid.New()}
	_, err := s.ListTags(model.SetUserInContext(context.Background(), user), model.DefaultPaginationParams())
	require.NoError(t, err)
	require.NotNil(t, got.UserID())
	require.Equal(t, user.Id, *got.UserID())
}

func TestService_UsesUnownedScopeWithoutUser(t *testing.T) {
	var got model.OwnerScope
	repo := &repository.RepoMock{
		ListTagFn: func(ctx context.Context, scope model.OwnerScope, params model.PaginationParams) (model.Page[model.Tag], error) {
			got = scope
			return model.Page[model.Tag]{}, nil
		},
	}
	s := service.NewService(repo)

	_, err := s.ListTags(context.Background(), model.DefaultPaginationParams())
	require.NoError(t, err)
	require.Nil(t, got.UserID())
}
