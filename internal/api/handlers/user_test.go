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
		name         string
		setupMock    func(u *service.UserServiceMock)
		expectedResp api.GetCurrentUserResponseObject
		expectedErr  string
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
