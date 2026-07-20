package handlers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/larssonoliver/inundated/internal/api"
	"github.com/larssonoliver/inundated/internal/api/handlers"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/service"
	"github.com/oapi-codegen/runtime/types"
)

func TestUserHandler_GetCurrentUser(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name          string
		setupMock     func(u *service.UserServiceMock)
		expectedResp  api.GetCurrentUserResponseObject
		expectedErr   string
	}{
		{
			name: "Success - returns 200 JSON payload",
			setupMock: func(u *service.UserServiceMock) {
				u.GetCurrentUserFn = func(ctx context.Context) (model.User, error) {
					return model.User{
						Id:    userID,
						Sub:   "auth0|123",
						Name:  "Jane Doe",
						Email: "jane@example.com",
					}, nil
				}
			},
			expectedResp: api.GetCurrentUser200JSONResponse{
				Id:    userID,
				Sub:   "auth0|123",
				Name:  func(s string) *string { return &s }("Jane Doe"),
				Email: types.Email("jane@example.com"),
			},
		},
		{
			name: "User not found - returns 401 response object",
			setupMock: func(u *service.UserServiceMock) {
				u.GetCurrentUserFn = func(ctx context.Context) (model.User, error) {
					return model.User{}, model.ErrNotFound
				}
			},
			expectedResp: api.GetCurrentUser401Response{},
		},
		{
			name: "Service layer crash - returns explicit internal error",
			setupMock: func(u *service.UserServiceMock) {
				u.GetCurrentUserFn = func(ctx context.Context) (model.User, error) {
					return model.User{}, errors.New("database disconnected")
				}
			},
			expectedErr: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &service.UserServiceMock{}
			tt.setupMock(mockSvc)
			handler := handlers.NewUserHandler(mockSvc)

			resp, err := handler.GetCurrentUser(context.Background(), api.GetCurrentUserRequestObject{})

			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr, err.Error())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}
		})
	}
}

func TestUserHandler_UpdateCurrentUser(t *testing.T) {
	userID := uuid.New()
	initialUser := model.User{
		Id:    userID,
		Sub:   "auth0|123",
		Name:  "Old Name",
		Email: "old@example.com",
	}

	newName := "New Name"
	newEmail := types.Email("new@example.com")

	tests := []struct {
		name          string
		request       api.UpdateCurrentUserRequestObject
		setupMock     func(u *service.UserServiceMock)
		expectedResp  api.UpdateCurrentUserResponseObject
		expectedErr   string
	}{
		{
			name: "Success - updates name and email fields specifically",
			request: api.UpdateCurrentUserRequestObject{
				Body: &api.UpdateCurrentUserJSONRequestBody{
					Name:  &newName,
					Email: &newEmail,
				},
			},
			setupMock: func(u *service.UserServiceMock) {
				u.GetCurrentUserFn = func(ctx context.Context) (model.User, error) {
					return initialUser, nil
				}
				u.UpdateCurrentUserFn = func(ctx context.Context, user model.User) (model.User, error) {
					// Verify changes were applied before updating
					assert.Equal(t, "New Name", user.Name)
					assert.Equal(t, "new@example.com", user.Email)
					return user, nil
				}
			},
			expectedResp: api.UpdateCurrentUser200JSONResponse{
				Id:    userID,
				Sub:   "auth0|123",
				Name:  &newName,
				Email: newEmail,
			},
		},
		{
			name: "User not found initially - returns 401 response object",
			request: api.UpdateCurrentUserRequestObject{
				Body: &api.UpdateCurrentUserJSONRequestBody{},
			},
			setupMock: func(u *service.UserServiceMock) {
				u.GetCurrentUserFn = func(ctx context.Context) (model.User, error) {
					return model.User{}, model.ErrNotFound
				}
			},
			expectedResp: api.UpdateCurrentUser401Response{},
		},
		{
			name: "Invalid argument validation failure - returns 400 response object",
			request: api.UpdateCurrentUserRequestObject{
				Body: &api.UpdateCurrentUserJSONRequestBody{Name: &newName},
			},
			setupMock: func(u *service.UserServiceMock) {
				u.GetCurrentUserFn = func(ctx context.Context) (model.User, error) {
					return initialUser, nil
				}
				u.UpdateCurrentUserFn = func(ctx context.Context, user model.User) (model.User, error) {
					return model.User{}, model.ErrInvalidArgument
				}
			},
			expectedResp: api.UpdateCurrentUser400Response{},
		},
		{
			name: "Update pipeline crash - returns explicit internal error",
			request: api.UpdateCurrentUserRequestObject{
				Body: &api.UpdateCurrentUserJSONRequestBody{Name: &newName},
			},
			setupMock: func(u *service.UserServiceMock) {
				u.GetCurrentUserFn = func(ctx context.Context) (model.User, error) {
					return initialUser, nil
				}
				u.UpdateCurrentUserFn = func(ctx context.Context, user model.User) (model.User, error) {
					return model.User{}, errors.New("db execution timed out")
				}
			},
			expectedErr: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &service.UserServiceMock{}
			tt.setupMock(mockSvc)
			handler := handlers.NewUserHandler(mockSvc)

			resp, err := handler.UpdateCurrentUser(context.Background(), tt.request)

			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr, err.Error())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}
		})
	}
}
