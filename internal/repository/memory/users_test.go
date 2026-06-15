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
		user    *model.User
		wantErr bool
		errType error
	}{
		{
			name: "Test CreateUser with valid input",
			user: &model.User{
				ID:    uuid.New(),
				Sub:   "auth0|user123",
				Email: "user@example.com",
				Name:  "Test User",
			},
			wantErr: false,
		},
		{
			name: "Test CreateUser with empty name",
			user: &model.User{
				ID:    uuid.New(),
				Sub:   "auth0|noname",
				Email: "noname@example.com",
				Name:  "",
			},
			wantErr: false,
		},
		{
			name: "Test CreateUser with empty sub",
			user: &model.User{
				ID:    uuid.New(),
				Sub:   "",
				Email: "user@example.com",
				Name:  "User",
			},
			wantErr: true,
			errType: model.ErrInvalidArgument,
		},
		{
			name: "Test CreateUser with empty email",
			user: &model.User{
				ID:    uuid.New(),
				Sub:   "auth0|user",
				Email: "",
				Name:  "User",
			},
			wantErr: true,
			errType: model.ErrInvalidArgument,
		},
		{
			name: "Test CreateUser with duplicate sub",
			user: &model.User{
				ID:    uuid.New(),
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
				firstUser := &model.User{
					ID:    uuid.New(),
					Sub:   "auth0|duplicate",
					Email: "first@example.com",
					Name:  "First",
				}
				err := store.Create(context.Background(), firstUser)
				require.NoError(t, err)

				// Now try to create the duplicate
				gotErr := store.Create(context.Background(), tt.user)
				require.Error(t, gotErr)
				require.ErrorIs(t, gotErr, model.ErrAlreadyExists)
				return
			}

			gotErr := store.Create(context.Background(), tt.user)
			if tt.wantErr {
				require.Error(t, gotErr)
				if tt.errType != nil {
					require.ErrorIs(t, gotErr, tt.errType)
				}
				return
			}

			require.NoError(t, gotErr)
			require.NotZero(t, tt.user.CreatedAt)
			require.NotZero(t, tt.user.UpdatedAt)
		})
	}
}

func TestUserStore_GetByID(t *testing.T) {
	ctx := context.Background()
	store := memory.NewMemoryStore()

	user := &model.User{
		ID:    uuid.New(),
		Sub:   "auth0|getbyid",
		Email: "getbyid@example.com",
		Name:  "Get By ID",
	}

	err := store.Create(ctx, user)
	require.NoError(t, err)

	got, err := store.GetByID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, user.ID, got.ID)
	require.Equal(t, user.Sub, got.Sub)
	require.Equal(t, user.Email, got.Email)
	require.Equal(t, user.Name, got.Name)
}

func TestUserStore_GetByIDNotFound(t *testing.T) {
	ctx := context.Background()
	store := memory.NewMemoryStore()

	_, err := store.GetByID(ctx, uuid.New())
	require.ErrorIs(t, err, model.ErrNotFound)
}

func TestUserStore_GetBySub(t *testing.T) {
	ctx := context.Background()
	store := memory.NewMemoryStore()

	user := &model.User{
		ID:    uuid.New(),
		Sub:   "google|getbysub123",
		Email: "getbysub@example.com",
		Name:  "Get By Sub",
	}

	err := store.Create(ctx, user)
	require.NoError(t, err)

	got, err := store.GetBySub(ctx, user.Sub)
	require.NoError(t, err)
	require.Equal(t, user.ID, got.ID)
	require.Equal(t, user.Sub, got.Sub)
	require.Equal(t, user.Email, got.Email)
	require.Equal(t, user.Name, got.Name)
}

func TestUserStore_GetBySubNotFound(t *testing.T) {
	ctx := context.Background()
	store := memory.NewMemoryStore()

	_, err := store.GetBySub(ctx, "nonexistent|sub")
	require.ErrorIs(t, err, model.ErrNotFound)
}

func TestUserStore_Update(t *testing.T) {
	ctx := context.Background()
	store := memory.NewMemoryStore()

	user := &model.User{
		ID:    uuid.New(),
		Sub:   "auth0|update",
		Email: "old@example.com",
		Name:  "Old Name",
	}

	err := store.Create(ctx, user)
	require.NoError(t, err)

	originalCreatedAt := user.CreatedAt

	newEmail := "new@example.com"
	newName := "New Name"
	updated, err := store.Update(ctx, user.ID, &model.UpdateUser{
		Email: &newEmail,
		Name:  &newName,
	})
	require.NoError(t, err)
	require.Equal(t, newEmail, updated.Email)
	require.Equal(t, newName, updated.Name)
	require.Equal(t, originalCreatedAt, updated.CreatedAt)
	require.NotEqual(t, originalCreatedAt, updated.UpdatedAt)

	// Verify persistence
	got, _ := store.GetByID(ctx, user.ID)
	require.Equal(t, newEmail, got.Email)
	require.Equal(t, newName, got.Name)
}

func TestUserStore_UpdatePartial(t *testing.T) {
	ctx := context.Background()
	store := memory.NewMemoryStore()

	user := &model.User{
		ID:    uuid.New(),
		Sub:   "auth0|partial",
		Email: "original@example.com",
		Name:  "Original Name",
	}

	err := store.Create(ctx, user)
	require.NoError(t, err)

	newEmail := "changed@example.com"
	updated, err := store.Update(ctx, user.ID, &model.UpdateUser{
		Email: &newEmail,
		Name:  nil,
	})
	require.NoError(t, err)
	require.Equal(t, newEmail, updated.Email)
	require.Equal(t, "Original Name", updated.Name)

	got, _ := store.GetByID(ctx, user.ID)
	require.Equal(t, newEmail, got.Email)
	require.Equal(t, "Original Name", got.Name)
}

func TestUserStore_UpdateMissing(t *testing.T) {
	ctx := context.Background()
	store := memory.NewMemoryStore()

	email := "ghost@example.com"
	_, err := store.Update(ctx, uuid.New(), &model.UpdateUser{
		Email: &email,
	})
	require.ErrorIs(t, err, model.ErrNotFound)
}

func TestUserStore_CreateMultiple(t *testing.T) {
	ctx := context.Background()
	store := memory.NewMemoryStore()

	user1 := &model.User{
		ID:    uuid.New(),
		Sub:   "auth0|user1",
		Email: "user1@example.com",
		Name:  "User 1",
	}

	user2 := &model.User{
		ID:    uuid.New(),
		Sub:   "auth0|user2",
		Email: "user2@example.com",
		Name:  "User 2",
	}

	err := store.Create(ctx, user1)
	require.NoError(t, err)

	err = store.Create(ctx, user2)
	require.NoError(t, err)

	got1, _ := store.GetByID(ctx, user1.ID)
	require.Equal(t, "User 1", got1.Name)

	got2, _ := store.GetByID(ctx, user2.ID)
	require.Equal(t, "User 2", got2.Name)
}
