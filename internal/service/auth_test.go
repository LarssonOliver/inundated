package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/auth"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAuthServiceImpl_BeginLogin(t *testing.T) {

	userService := &service.UserServiceMock{}
	sessionRepository := &repository.SessionRepoMock{}

	t.Run("CreatesLoginState", func(t *testing.T) {

		createLoginStateCalled := false
		redirectUri := "http://localhost/callback"
		authorizationUrl := "http://auth.example.com/authorize?state=some-state"
		stateId := uuid.New()

		loginStateRepository := &repository.LoginStateRepoMock{
			CreateLoginStateFn: func(ctx context.Context, state model.LoginState) (model.LoginState, error) {
				require.False(t, createLoginStateCalled, "CreateLoginState should only be called once")

				require.NotEmpty(t, state.Id, "LoginState State should not be empty")
				require.NotEqual(t, uuid.Nil, state.Id, "LoginState ID should not be nil")
				require.Equal(t, redirectUri, state.RedirectUri, "LoginState RedirectURI should match the provided redirectURI")
				require.WithinDuration(t, state.ExpiresAt, time.Now(), 30*time.Minute, "LoginState ExpiresAt should be within 5 minutes from now")

				state.Id = stateId // Set the ID to a known value for testing

				createLoginStateCalled = true
				return state, nil
			},
		}
		oidcClient := &auth.OIDCClientMock{
			BeginAuthorizationFn: func(state string) (auth.OIDCAuthorizationRequest, error) {
				return auth.OIDCAuthorizationRequest{
					Uri:          authorizationUrl,
					CodeVerifier: "some-code-verifier",
				}, nil
			},
		}

		authService := service.NewAuthService(userService, sessionRepository, loginStateRepository, oidcClient)

		authUrl, err := authService.BeginLogin(t.Context(), redirectUri)

		require.NoError(t, err)
		require.True(t, createLoginStateCalled, "CreateLoginState should have been called")
		require.Equal(t, authorizationUrl, authUrl, "Authorization URL should match the expected value")
	})

	t.Run("ReturnsErrorWhenOIDCClientFails", func(t *testing.T) {
		loginStateRepository := &repository.LoginStateRepoMock{}
		oidcClient := &auth.OIDCClientMock{
			BeginAuthorizationFn: func(state string) (auth.OIDCAuthorizationRequest, error) {
				return auth.OIDCAuthorizationRequest{}, errors.New("OIDC client error")
			},
		}
		authService := service.NewAuthService(userService, sessionRepository, loginStateRepository, oidcClient)
		authUrl, err := authService.BeginLogin(t.Context(), "http://localhost/callback")
		require.Error(t, err)
		require.Empty(t, authUrl, "Authorization URL should be empty when there is an error")
	})

	t.Run("ReturnsErrorWhenLoginStateCreationFails", func(t *testing.T) {
		loginStateRepository := &repository.LoginStateRepoMock{
			CreateLoginStateFn: func(ctx context.Context, state model.LoginState) (model.LoginState, error) {
				return model.LoginState{}, errors.New("failed to create login state")
			},
		}
		oidcClient := &auth.OIDCClientMock{
			BeginAuthorizationFn: func(state string) (auth.OIDCAuthorizationRequest, error) {
				return auth.OIDCAuthorizationRequest{
					Uri:          "http://auth.example.com/authorize?state=some-state",
					CodeVerifier: "some-code-verifier",
				}, nil
			},
		}
		authService := service.NewAuthService(userService, sessionRepository, loginStateRepository, oidcClient)
		authUrl, err := authService.BeginLogin(t.Context(), "http://localhost/callback")
		require.Error(t, err)
		require.Empty(t, authUrl, "Authorization URL should be empty when there is an error")
	})
}

func TestAuthServiceImpl_HandleCallback(t *testing.T) {

	t.Run("SuccessfullyHandlesCallback", func(t *testing.T) {
		codeVal := "some-code"
		verifier := "some-code-verifier"

		loginStateID := uuid.New()
		sessionID := uuid.New()
		userSub := "user-sub-123"
		userId := uuid.New()
		userName := "John Doe"
		userEmail := "john@example.com"
		redirectUri := "http://localhost/callback"

		loginStateRepository := &repository.LoginStateRepoMock{
			GetLoginStateFn: func(ctx context.Context, stateID uuid.UUID) (model.LoginState, error) {
				require.Equal(t, loginStateID, stateID, "LoginState ID should match the one provided in the callback")
				return model.LoginState{
					Id:           loginStateID,
					RedirectUri:  redirectUri,
					CodeVerifier: "some-code-verifier",
					ExpiresAt:    time.Now().Add(5 * time.Minute),
				}, nil
			},
			DeleteLoginStateFn: func(ctx context.Context, stateID uuid.UUID) error {
				require.Equal(t, loginStateID, stateID, "LoginState ID should match the one provided in the callback")
				return nil
			},
		}
		userService := &service.UserServiceMock{
			GetOrCreateUserByIdentityFn: func(ctx context.Context, identity model.UserIdentity) (model.User, error) {
				require.Equal(t, userSub, identity.Sub, "identity sub should match the one returned by the OIDC client")
				require.Equal(t, userEmail, identity.Email, "identity email should match the one returned by the OIDC client")
				require.Equal(t, userName, identity.Name, "identity name should match the one returned by the OIDC client")
				return model.User{
					Id:    userId,
					Sub:   identity.Sub,
					Email: identity.Email,
					Name:  identity.Name,
				}, nil
			},
		}
		sessionRepository := &repository.SessionRepoMock{
			CreateSessionFn: func(ctx context.Context, session model.Session) (model.Session, error) {
				session.Id = sessionID
				return session, nil
			},
		}
		oidcClient := &auth.OIDCClientMock{
			ExchangeCodeFn: func(ctx context.Context, code string, codeVerifier string) (auth.OIDCIdentity, error) {
				require.Equal(t, verifier, codeVerifier, "Code verifier should match the one stored in login state")
				require.Equal(t, codeVal, code, "Code should match the one provided in the callback")
				return auth.OIDCIdentity{
					Sub:   userSub,
					Name:  userName,
					Email: userEmail,
				}, nil
			},
		}

		authService := service.NewAuthService(userService, sessionRepository, loginStateRepository, oidcClient)
		session, redUri, err := authService.HandleCallback(t.Context(), loginStateID, codeVal)
		require.NoError(t, err)
		require.Equal(t, redirectUri, redUri, "Redirect URI should match the one stored in login state")
		require.Equal(t, sessionID, session.Id, "Session ID should match the one returned by the session repository")
		require.Equal(t, userSub, session.Sub, "Session user sub should match the one returned by the OIDC client")
		require.Equal(t, userId, session.UserId, "Session user ID should match the one returned by the user service")
		require.WithinDuration(t, time.Now(), session.ExpiresAt, 24*time.Hour, "Session expiration should be within 30 minutes from now")
	})

	t.Run("ReturnsErrorWhenLoginStateNotFound", func(t *testing.T) {
		loginStateRepository := &repository.LoginStateRepoMock{
			GetLoginStateFn: func(ctx context.Context, stateID uuid.UUID) (model.LoginState, error) {
				return model.LoginState{}, errors.New("login state not found")
			},
		}
		sessionRepository := &repository.SessionRepoMock{}
		userService := &service.UserServiceMock{}
		oidcClient := &auth.OIDCClientMock{}
		authService := service.NewAuthService(userService, sessionRepository, loginStateRepository, oidcClient)
		session, redirectUri, err := authService.HandleCallback(t.Context(), uuid.New(), "some-code")
		require.Error(t, err)
		require.Empty(t, session, "Session should be empty when there is an error")
		require.Empty(t, redirectUri, "Redirect URI should be empty when there is an error")
	})

	t.Run("ReturnsErrorWhenLoginStateExpired", func(t *testing.T) {
		deleteCalled := false
		loginStateRepository := &repository.LoginStateRepoMock{
			GetLoginStateFn: func(ctx context.Context, stateID uuid.UUID) (model.LoginState, error) {
				return model.LoginState{
					Id:           stateID,
					RedirectUri:  "http://localhost/callback",
					CodeVerifier: "some-code-verifier",
					ExpiresAt:    time.Now().Add(-5 * time.Minute), // Expired
				}, nil
			},
			DeleteLoginStateFn: func(ctx context.Context, stateID uuid.UUID) error {
				deleteCalled = true
				return nil
			},
		}
		sessionRepository := &repository.SessionRepoMock{}
		userService := &service.UserServiceMock{}
		oidcClient := &auth.OIDCClientMock{}
		authService := service.NewAuthService(userService, sessionRepository, loginStateRepository, oidcClient)
		session, redirectUri, err := authService.HandleCallback(t.Context(), uuid.New(), "some-code")
		require.Error(t, err)
		require.Equal(t, model.ErrLoginStateExpired, err, "Error should indicate that the login state has expired")
		require.Empty(t, session, "Session should be empty when there is an error")
		require.Empty(t, redirectUri, "Redirect URI should be empty when there is an error")
		require.True(t, deleteCalled, "DeleteLoginState should have been called even when the login state is expired")
	})

	t.Run("ReturnsErrorWhenOIDCClientFails", func(t *testing.T) {
		deleteCalled := false
		loginStateRepository := &repository.LoginStateRepoMock{
			GetLoginStateFn: func(ctx context.Context, stateID uuid.UUID) (model.LoginState, error) {
				return model.LoginState{
					Id:           stateID,
					RedirectUri:  "http://localhost/callback",
					CodeVerifier: "some-code-verifier",
					ExpiresAt:    time.Now().Add(15 * time.Minute), // Expired
				}, nil
			},
			DeleteLoginStateFn: func(ctx context.Context, stateID uuid.UUID) error {
				deleteCalled = true
				return nil
			},
		}
		sessionRepository := &repository.SessionRepoMock{}
		oidcClient := &auth.OIDCClientMock{
			ExchangeCodeFn: func(ctx context.Context, code string, codeVerifier string) (auth.OIDCIdentity, error) {
				return auth.OIDCIdentity{}, errors.New("OIDC client error")
			},
		}
		userService := &service.UserServiceMock{}
		authService := service.NewAuthService(userService, sessionRepository, loginStateRepository, oidcClient)
		session, redirectUri, err := authService.HandleCallback(t.Context(), uuid.New(), "some-code")
		require.Error(t, err)
		require.Empty(t, session, "Session should be empty when there is an error")
		require.Empty(t, redirectUri, "Redirect URI should be empty when there is an error")
		require.True(t, deleteCalled, "DeleteLoginState should have been called even when the login state is expired")
	})

	t.Run("ReturnsErrorWhenUserServiceFails", func(t *testing.T) {
		deleteCalled := false
		loginStateRepository := &repository.LoginStateRepoMock{
			GetLoginStateFn: func(ctx context.Context, stateID uuid.UUID) (model.LoginState, error) {
				return model.LoginState{
					Id:           stateID,
					RedirectUri:  "http://localhost/callback",
					CodeVerifier: "some-code-verifier",
					ExpiresAt:    time.Now().Add(15 * time.Minute), // Expired
				}, nil
			},
			DeleteLoginStateFn: func(ctx context.Context, stateID uuid.UUID) error {
				deleteCalled = true
				return nil
			},
		}
		oidcClient := &auth.OIDCClientMock{
			ExchangeCodeFn: func(ctx context.Context, code string, codeVerifier string) (auth.OIDCIdentity, error) {
				return auth.OIDCIdentity{
					Sub:   "user-sub-123",
					Name:  "user",
					Email: "test@example.com",
				}, nil
			},
		}
		userService := &service.UserServiceMock{
			GetOrCreateUserByIdentityFn: func(ctx context.Context, identity model.UserIdentity) (model.User, error) {
				return model.User{}, errors.New("user service error")
			},
		}
		sessionRepository := &repository.SessionRepoMock{}
		authService := service.NewAuthService(userService, sessionRepository, loginStateRepository, oidcClient)
		session, redirectUri, err := authService.HandleCallback(t.Context(), uuid.New(), "some-code")
		require.Error(t, err)
		require.Empty(t, session, "Session should be empty when there is an error")
		require.Empty(t, redirectUri, "Redirect URI should be empty when there is an error")
		require.True(t, deleteCalled, "DeleteLoginState should have been called even when the login state is expired")
	})

	t.Run("ReturnsErrorWhenSessionCreationFails", func(t *testing.T) {
		deleteCalled := false
		loginStateRepository := &repository.LoginStateRepoMock{
			GetLoginStateFn: func(ctx context.Context, stateID uuid.UUID) (model.LoginState, error) {
				return model.LoginState{
					Id:           stateID,
					RedirectUri:  "http://localhost/callback",
					CodeVerifier: "some-code-verifier",
					ExpiresAt:    time.Now().Add(15 * time.Minute), // Expired
				}, nil
			},
			DeleteLoginStateFn: func(ctx context.Context, stateID uuid.UUID) error {
				deleteCalled = true
				return nil
			},
		}
		oidcClient := &auth.OIDCClientMock{
			ExchangeCodeFn: func(ctx context.Context, code string, codeVerifier string) (auth.OIDCIdentity, error) {
				return auth.OIDCIdentity{
					Sub:   "user-sub-123",
					Name:  "user",
					Email: "test@example.com",
				}, nil
			},
		}
		userService := &service.UserServiceMock{
			GetOrCreateUserByIdentityFn: func(ctx context.Context, identity model.UserIdentity) (model.User, error) {
				return model.User{
					Id:    uuid.New(),
					Name:  identity.Name,
					Sub:   identity.Sub,
					Email: identity.Email,
				}, nil
			},
		}
		sessionRepository := &repository.SessionRepoMock{
			CreateSessionFn: func(ctx context.Context, session model.Session) (model.Session, error) {
				return model.Session{}, errors.New("session creation error")
			},
		}
		authService := service.NewAuthService(userService, sessionRepository, loginStateRepository, oidcClient)
		session, redirectUri, err := authService.HandleCallback(t.Context(), uuid.New(), "some-code")
		require.Error(t, err)
		require.Empty(t, session, "Session should be empty when there is an error")
		require.Empty(t, redirectUri, "Redirect URI should be empty when there is an error")
		require.True(t, deleteCalled, "DeleteLoginState should have been called even when the login state is expired")
	})

	t.Run("LogoutDeletesSession", func(t *testing.T) {
		id := uuid.New()
		deleteCalled := false

		loginStateRepository := &repository.LoginStateRepoMock{}
		userService := &service.UserServiceMock{}
		oidcClient := &auth.OIDCClientMock{}
		sessionRepository := &repository.SessionRepoMock{
			DeleteSessionFn: func(ctx context.Context, sessionID uuid.UUID) error {
				require.Equal(t, id, sessionID, "Session ID should match the one provided for deletion")
				deleteCalled = true
				return nil
			},
		}

		authService := service.NewAuthService(userService, sessionRepository, loginStateRepository, oidcClient)
		err := authService.LogoutSession(t.Context(), id)
		require.NoError(t, err)
		require.True(t, deleteCalled, "DeleteSession should have been called")
	})
}
