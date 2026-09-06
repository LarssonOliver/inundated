package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/service"
	"github.com/stretchr/testify/require"
)

func TestUserService_GetCurrentUser(t *testing.T) {
	testUser := model.User{
		Id:    uuid.New(),
		Sub:   "auth0|abc123",
		Email: "test@example.com",
		Name:  "Test User",
	}

	tests := []struct {
		name    string
		ctx     func() context.Context
		want    model.User
		wantErr bool
	}{
		{
			name: "user present in context",
			ctx: func() context.Context {
				return context.WithValue(context.Background(), model.UserContextKey, testUser)
			},
			want:    testUser,
			wantErr: false,
		},
		{
			name: "no user in context",
			ctx: func() context.Context {
				return context.Background()
			},
			want:    model.User{},
			wantErr: true,
		},
		{
			name: "wrong type in context",
			ctx: func() context.Context {
				return context.WithValue(context.Background(), model.UserContextKey, "not-a-user")
			},
			want:    model.User{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repository.RepoMock{}
			s := service.NewService(repo)

			got, gotErr := s.GetCurrentUser(tt.ctx())
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			require.Equal(t, tt.want.Id, got.Id)
			require.Equal(t, tt.want.Sub, got.Sub)
			require.Equal(t, tt.want.Email, got.Email)
			require.Equal(t, tt.want.Name, got.Name)
		})
	}
}

func TestUserService_GetOrCreateUserBySub(t *testing.T) {
	existingUser := model.User{
		Id:    uuid.New(),
		Sub:   "auth0|existing",
		Email: "existing@example.com",
		Name:  "Existing User",
	}

	tests := []struct {
		name        string
		sub         string
		getBySubFn  func(ctx context.Context, sub string) (model.User, error)
		createFn    func(ctx context.Context, user model.User) (model.User, model.OrphanAdoption, error)
		wantCreated bool
		want        model.User
		wantErr     bool
	}{
		{
			name: "user already exists",
			sub:  "auth0|existing",
			getBySubFn: func(ctx context.Context, sub string) (model.User, error) {
				return existingUser, nil
			},
			createFn: func(ctx context.Context, user model.User) (model.User, model.OrphanAdoption, error) {
				t.Fatal("CreateUserAdoptingOrphans should not be called when user already exists")
				return model.User{}, model.OrphanAdoption{}, nil
			},
			want:    existingUser,
			wantErr: false,
		},
		{
			name: "user does not exist - creates new user and adopts orphans",
			sub:  "auth0|new",
			getBySubFn: func(ctx context.Context, sub string) (model.User, error) {
				return model.User{}, model.ErrNotFound
			},
			createFn: func(ctx context.Context, user model.User) (model.User, model.OrphanAdoption, error) {
				require.Equal(t, "auth0|new", user.Sub)
				user.Id = uuid.New()
				return user, model.OrphanAdoption{Projects: 2, Tags: 1, Timespans: 3}, nil
			},
			want:    model.User{Sub: "auth0|new"},
			wantErr: false,
		},
		{
			name: "repository error on lookup",
			sub:  "auth0|broken",
			getBySubFn: func(ctx context.Context, sub string) (model.User, error) {
				return model.User{}, errors.New("database error")
			},
			createFn: func(ctx context.Context, user model.User) (model.User, model.OrphanAdoption, error) {
				t.Fatal("CreateUserAdoptingOrphans should not be called on a non-not-found lookup error")
				return model.User{}, model.OrphanAdoption{}, nil
			},
			wantErr: true,
		},
		{
			name: "repository error on create",
			sub:  "auth0|new",
			getBySubFn: func(ctx context.Context, sub string) (model.User, error) {
				return model.User{}, model.ErrNotFound
			},
			createFn: func(ctx context.Context, user model.User) (model.User, model.OrphanAdoption, error) {
				return model.User{}, model.OrphanAdoption{}, errors.New("database error")
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repository.RepoMock{
				GetUserBySubFn:              tt.getBySubFn,
				CreateUserAdoptingOrphansFn: tt.createFn,
			}
			s := service.NewService(repo)

			got, gotErr := s.GetOrCreateUserBySub(context.Background(), tt.sub)
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			require.NotEqual(t, uuid.Nil, got.Id)
			require.Equal(t, tt.want.Sub, got.Sub)
		})
	}
}

func TestUserService_UpdateCurrentUser(t *testing.T) {
	userId := uuid.New()

	tests := []struct {
		name      string
		user      model.User
		contextFn func() context.Context
		updateFn  func(ctx context.Context, user model.User) (model.User, error)
		want      model.User
		wantErr   bool
	}{
		{
			name: "successful update",
			user: model.User{Id: userId, Sub: "auth0|abc", Email: "new@example.com", Name: "New Name"},
			contextFn: func() context.Context {
				return context.WithValue(context.Background(), model.UserContextKey, model.User{Id: userId, Sub: "auth0|abc"})
			},
			updateFn: func(ctx context.Context, user model.User) (model.User, error) {
				return user, nil
			},
			want:    model.User{Id: userId, Sub: "auth0|abc", Email: "new@example.com", Name: "New Name"},
			wantErr: false,
		},
		{
			name: "repository error",
			user: model.User{Id: userId, Sub: "auth0|abc", Email: "new@example.com", Name: "New Name"},
			contextFn: func() context.Context {
				return context.WithValue(context.Background(), model.UserContextKey, model.User{Id: userId, Sub: "auth0|abc"})
			},
			updateFn: func(ctx context.Context, user model.User) (model.User, error) {
				return model.User{}, errors.New("database error")
			},
			want:    model.User{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repository.RepoMock{
				UpdateUserFn: tt.updateFn,
			}
			s := service.NewService(repo)

			got, gotErr := s.UpdateCurrentUser(tt.contextFn(), tt.user)
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			require.Equal(t, tt.want.Id, got.Id)
			require.Equal(t, tt.want.Email, got.Email)
			require.Equal(t, tt.want.Name, got.Name)
		})
	}
}
