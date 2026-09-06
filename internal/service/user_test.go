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

func TestUserService_GetUserBySub(t *testing.T) {
	existingUser := model.User{Id: uuid.New(), Sub: "auth0|existing", Email: "e@example.com", Name: "E"}

	t.Run("found", func(t *testing.T) {
		repo := &repository.RepoMock{
			GetUserBySubFn: func(ctx context.Context, sub string) (model.User, error) {
				require.Equal(t, "auth0|existing", sub)
				return existingUser, nil
			},
		}
		got, err := service.NewService(repo).GetUserBySub(context.Background(), "auth0|existing")
		require.NoError(t, err)
		require.Equal(t, existingUser, got)
	})

	t.Run("not found is propagated", func(t *testing.T) {
		repo := &repository.RepoMock{
			GetUserBySubFn: func(ctx context.Context, sub string) (model.User, error) {
				return model.User{}, model.ErrNotFound
			},
		}
		_, err := service.NewService(repo).GetUserBySub(context.Background(), "auth0|ghost")
		require.ErrorIs(t, err, model.ErrNotFound)
	})
}

func TestUserService_GetOrCreateUserByIdentity(t *testing.T) {
	existingUser := model.User{
		Id:    uuid.New(),
		Sub:   "auth0|existing",
		Email: "existing@example.com",
		Name:  "Existing User",
	}

	newIdentity := model.UserIdentity{Sub: "auth0|new", Email: "new@example.com", Name: "New User"}

	tests := []struct {
		name       string
		identity   model.UserIdentity
		getBySubFn func(ctx context.Context, sub string) (model.User, error)
		createFn   func(ctx context.Context, user model.User) (model.User, model.OrphanAdoption, error)
		updateFn   func(ctx context.Context, user model.User) (model.User, error)
		want       model.User
		wantErr    bool
	}{
		{
			name:     "existing user, identity unchanged - returned as is, no write",
			identity: model.UserIdentity{Sub: existingUser.Sub, Email: existingUser.Email, Name: existingUser.Name},
			getBySubFn: func(ctx context.Context, sub string) (model.User, error) {
				return existingUser, nil
			},
			createFn: func(ctx context.Context, user model.User) (model.User, model.OrphanAdoption, error) {
				t.Fatal("CreateUserAdoptingOrphans should not be called when user already exists")
				return model.User{}, model.OrphanAdoption{}, nil
			},
			updateFn: func(ctx context.Context, user model.User) (model.User, error) {
				t.Fatal("UpdateUser should not be called when the identity has not drifted")
				return model.User{}, nil
			},
			want: existingUser,
		},
		{
			name:     "existing user, email drifted - updated to identity",
			identity: model.UserIdentity{Sub: existingUser.Sub, Email: "changed@example.com", Name: existingUser.Name},
			getBySubFn: func(ctx context.Context, sub string) (model.User, error) {
				return existingUser, nil
			},
			updateFn: func(ctx context.Context, user model.User) (model.User, error) {
				require.Equal(t, existingUser.Id, user.Id)
				require.Equal(t, "changed@example.com", user.Email)
				return user, nil
			},
			want: model.User{Id: existingUser.Id, Sub: existingUser.Sub, Email: "changed@example.com", Name: existingUser.Name},
		},
		{
			name:     "new subject - created from identity and adopts orphans",
			identity: newIdentity,
			getBySubFn: func(ctx context.Context, sub string) (model.User, error) {
				return model.User{}, model.ErrNotFound
			},
			createFn: func(ctx context.Context, user model.User) (model.User, model.OrphanAdoption, error) {
				require.Equal(t, newIdentity.Sub, user.Sub)
				require.Equal(t, newIdentity.Email, user.Email)
				require.Equal(t, newIdentity.Name, user.Name)
				user.Id = uuid.New()
				return user, model.OrphanAdoption{Projects: 2, Tags: 1, Timespans: 3}, nil
			},
			want: model.User{Sub: newIdentity.Sub, Email: newIdentity.Email, Name: newIdentity.Name},
		},
		{
			name:     "repository error on lookup",
			identity: model.UserIdentity{Sub: "auth0|broken", Email: "b@example.com"},
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
			name:     "repository error on create",
			identity: newIdentity,
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
				UpdateUserFn:                tt.updateFn,
			}
			s := service.NewService(repo)

			got, gotErr := s.GetOrCreateUserByIdentity(context.Background(), tt.identity)
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			require.Equal(t, tt.want.Sub, got.Sub)
			require.Equal(t, tt.want.Email, got.Email)
			require.Equal(t, tt.want.Name, got.Name)
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
