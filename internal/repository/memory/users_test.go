package memory_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository/memory"
	"github.com/stretchr/testify/require"
)

func TestUserStore_Create(t *testing.T) {
	tests := []struct {
		name    string
		user    model.User
		wantErr bool
		errType error
	}{
		{
			name: "Test CreateUser with valid input",
			user: model.User{
				Id:    uuid.New(),
				Sub:   "auth0|user123",
				Email: "user@example.com",
				Name:  "Test User",
			},
			wantErr: false,
		},
		{
			name: "Test CreateUser with empty name",
			user: model.User{
				Id:    uuid.New(),
				Sub:   "auth0|noname",
				Email: "noname@example.com",
				Name:  "",
			},
			wantErr: false,
		},
		{
			name: "Test CreateUser with empty sub",
			user: model.User{
				Id:    uuid.New(),
				Sub:   "",
				Email: "user@example.com",
				Name:  "User",
			},
			wantErr: true,
			errType: model.ErrInvalidArgument,
		},
		{
			name: "Test CreateUser with empty email",
			user: model.User{
				Id:    uuid.New(),
				Sub:   "auth0|user",
				Email: "",
				Name:  "User",
			},
			wantErr: true,
			errType: model.ErrInvalidArgument,
		},
		{
			name: "Test CreateUser with duplicate sub",
			user: model.User{
				Id:    uuid.New(),
				Sub:   "auth0|duplicate",
				Email: "first@example.com",
				Name:  "First",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := memory.NewMemoryStore()

			// For duplicate sub test, create the first user first
			if tt.name == "Test CreateUser with duplicate sub" {
				firstUser := model.User{
					Id:    uuid.New(),
					Sub:   "auth0|duplicate",
					Email: "first@example.com",
					Name:  "First",
				}
				_, err := store.CreateUser(context.Background(), firstUser)
				require.NoError(t, err)

				// Now try to create the duplicate
				_, gotErr := store.CreateUser(context.Background(), tt.user)
				require.Error(t, gotErr)
				require.ErrorIs(t, gotErr, model.ErrAlreadyExists)
				return
			}

			_, gotErr := store.CreateUser(context.Background(), tt.user)
			if tt.wantErr {
				require.Error(t, gotErr)
				if tt.errType != nil {
					require.ErrorIs(t, gotErr, tt.errType)
				}
				return
			}

			require.NoError(t, gotErr)
		})
	}
}

func TestUserStore_GetByID(t *testing.T) {
	ctx := context.Background()
	store := memory.NewMemoryStore()

	user := model.User{
		Id:    uuid.New(),
		Sub:   "auth0|getbyid",
		Email: "getbyid@example.com",
		Name:  "Get By ID",
	}

	_, err := store.CreateUser(ctx, user)
	require.NoError(t, err)

	got, err := store.GetUser(ctx, user.Id)
	require.NoError(t, err)
	require.Equal(t, user.Id, got.Id)
	require.Equal(t, user.Sub, got.Sub)
	require.Equal(t, user.Email, got.Email)
	require.Equal(t, user.Name, got.Name)
}

func TestUserStore_GetByIDNotFound(t *testing.T) {
	ctx := context.Background()
	store := memory.NewMemoryStore()

	_, err := store.GetUser(ctx, uuid.New())
	require.ErrorIs(t, err, model.ErrNotFound)
}

func TestUserStore_GetBySub(t *testing.T) {
	ctx := context.Background()
	store := memory.NewMemoryStore()

	user := model.User{
		Id:    uuid.New(),
		Sub:   "google|getbysub123",
		Email: "getbysub@example.com",
		Name:  "Get By Sub",
	}

	_, err := store.CreateUser(ctx, user)
	require.NoError(t, err)

	got, err := store.GetUserBySub(ctx, user.Sub)
	require.NoError(t, err)
	require.Equal(t, user.Id, got.Id)
	require.Equal(t, user.Sub, got.Sub)
	require.Equal(t, user.Email, got.Email)
	require.Equal(t, user.Name, got.Name)
}

func TestUserStore_GetBySubNotFound(t *testing.T) {
	ctx := context.Background()
	store := memory.NewMemoryStore()

	_, err := store.GetUserBySub(ctx, "nonexistent|sub")
	require.ErrorIs(t, err, model.ErrNotFound)
}

func TestUserStore_Update(t *testing.T) {
	ctx := context.Background()
	store := memory.NewMemoryStore()

	user := model.User{
		Id:    uuid.New(),
		Sub:   "auth0|update",
		Email: "old@example.com",
		Name:  "Old Name",
	}

	_, err := store.CreateUser(ctx, user)
	require.NoError(t, err)

	newEmail := "new@example.com"
	newName := "New Name"
	updated, err := store.UpdateUser(ctx, model.User{
		Id:    user.Id,
		Sub:   user.Sub,
		Email: newEmail,
		Name:  newName,
	})
	require.NoError(t, err)
	require.Equal(t, newEmail, updated.Email)
	require.Equal(t, newName, updated.Name)

	// Verify persistence
	got, _ := store.GetUser(ctx, user.Id)
	require.Equal(t, newName, got.Name)
	require.Equal(t, newEmail, got.Email)
}

func TestUserStore_UpdateMissing(t *testing.T) {
	ctx := context.Background()
	store := memory.NewMemoryStore()

	email := "ghost@example.com"
	_, err := store.UpdateUser(ctx, model.User{
		Id:    uuid.New(),
		Email: email,
	})
	require.ErrorIs(t, err, model.ErrNotFound)
}

func TestUserStore_CreateMultiple(t *testing.T) {
	ctx := context.Background()
	store := memory.NewMemoryStore()

	user1 := model.User{
		Id:    uuid.New(),
		Sub:   "auth0|user1",
		Email: "user1@example.com",
		Name:  "User 1",
	}

	user2 := model.User{
		Id:    uuid.New(),
		Sub:   "auth0|user2",
		Email: "user2@example.com",
		Name:  "User 2",
	}

	_, err := store.CreateUser(ctx, user1)
	require.NoError(t, err)

	_, err = store.CreateUser(ctx, user2)
	require.NoError(t, err)

	got1, _ := store.GetUser(ctx, user1.Id)
	require.Equal(t, "User 1", got1.Name)

	got2, _ := store.GetUser(ctx, user2.Id)
	require.Equal(t, "User 2", got2.Name)
}
