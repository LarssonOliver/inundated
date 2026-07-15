package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/auth"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
)

type AuthService interface {
	BeginLogin(ctx context.Context, redirectURI string) (authorizationURL string, err error)
	HandleCallback(ctx context.Context, stateID uuid.UUID, code string) (session model.Session, redirectURI string, err error)
}

type AuthServiceImpl struct {
	userService          UserService
	sessionRepository    repository.SessionRepository
	loginStateRepository repository.LoginStateRepository
	oidcClient           auth.OIDCClient
}

var _ AuthService = (*AuthServiceImpl)(nil)

func NewAuthService(userService UserService, sessionRepository repository.SessionRepository, loginStateRepository repository.LoginStateRepository, oidcClient auth.OIDCClient) *AuthServiceImpl {
	return &AuthServiceImpl{
		userService:          userService,
		sessionRepository:    sessionRepository,
		loginStateRepository: loginStateRepository,
		oidcClient:           oidcClient,
	}
}

// BeginLogin implements [AuthService].
func (a *AuthServiceImpl) BeginLogin(ctx context.Context, redirectURI string) (authorizationURL string, err error) {

	stateId := uuid.New()

	authRequest, err := a.oidcClient.BeginAuthorization(stateId.String())
	if err != nil {
		return "", err
	}

	loginState := model.LoginState{
		Id:           stateId,
		RedirectUri:  redirectURI,
		CodeVerifier: authRequest.CodeVerifier,
		ExpiresAt:    time.Now().Add(5 * time.Minute),
	}

	_, err = a.loginStateRepository.CreateLoginState(ctx, loginState)
	if err != nil {
		return "", err
	}

	return authRequest.Uri, nil
}

// HandleCallback implements [AuthService].
func (a *AuthServiceImpl) HandleCallback(ctx context.Context, stateId uuid.UUID, code string) (session model.Session, redirectURI string, err error) {
	loginState, err := a.loginStateRepository.GetLoginState(ctx, stateId)
	if err != nil {
		return model.Session{}, "", err
	}

	defer func() {
		_ = a.loginStateRepository.DeleteLoginState(ctx, stateId)
	}()

	if time.Now().After(loginState.ExpiresAt) {
		return model.Session{}, "", model.ErrLoginStateExpired
	}

	identity, err := a.oidcClient.ExchangeCode(ctx, code, loginState.CodeVerifier)
	if err != nil {
		return model.Session{}, "", err
	}

	user, err := a.userService.GetOrCreateUserBySub(ctx, identity.Sub)
	if err != nil {
		return model.Session{}, "", err
	}

	if user.Name != identity.Name || user.Email != identity.Email {
		user, err = a.userService.UpdateCurrentUser(ctx, model.User{
			Id:    user.Id,
			Sub:   user.Sub,
			Email: identity.Email,
			Name:  identity.Name,
		})
		if err != nil {
			return model.Session{}, "", err
		}
	}

	session = model.Session{
		Id:        uuid.New(),
		UserId:    user.Id,
		Sub:       identity.Sub,
		ExpiresAt: time.Now().Add(8 * time.Hour),
	}

	session, err = a.sessionRepository.CreateSession(ctx, session)
	if err != nil {
		return model.Session{}, "", err
	}

	return session, loginState.RedirectUri, nil
}
