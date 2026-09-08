package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/auth"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
)

type AuthService interface {
	BeginLogin(ctx context.Context, redirectURI string) (authorizationURL string, err error)
	HandleCallback(ctx context.Context, stateID uuid.UUID, code string) (session model.Session, redirectURI string, err error)
	LogoutSession(ctx context.Context, sessionId uuid.UUID) error
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

	nonce, err := newNonce()
	if err != nil {
		return "", err
	}

	authRequest, err := a.oidcClient.BeginAuthorization(stateId.String(), nonce)
	if err != nil {
		return "", err
	}

	loginState := model.LoginState{
		Id:           stateId,
		RedirectUri:  redirectURI,
		CodeVerifier: authRequest.CodeVerifier,
		Nonce:        nonce,
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

	identity, err := a.oidcClient.ExchangeCode(ctx, code, loginState.CodeVerifier, loginState.Nonce)
	if err != nil {
		return model.Session{}, "", err
	}

	user, err := a.userService.GetOrCreateUserByIdentity(ctx, model.UserIdentity{
		Sub:   identity.Sub,
		Email: identity.Email,
		Name:  identity.Name,
	})
	if err != nil {
		return model.Session{}, "", err
	}

	token, err := newSessionToken()
	if err != nil {
		return model.Session{}, "", err
	}

	now := time.Now()
	session = model.Session{
		Id:        uuid.New(),
		UserId:    user.Id,
		Sub:       identity.Sub,
		Token:     token,
		CreatedAt: now,
		ExpiresAt: now.Add(8 * time.Hour),
	}

	session, err = a.sessionRepository.CreateSession(ctx, session)
	if err != nil {
		return model.Session{}, "", err
	}

	return session, loginState.RedirectUri, nil
}

// newNonce returns a cryptographically random, URL-safe OIDC nonce.
func newNonce() (string, error) {
	return randomToken("nonce")
}

// newSessionToken returns the cryptographically random secret placed in the
// session cookie. Only its hash is stored (see model.HashSessionToken).
func newSessionToken() (string, error) {
	return randomToken("session token")
}

func randomToken(label string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating %s: %w", label, err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// LogoutSession implements [AuthService].
func (a *AuthServiceImpl) LogoutSession(ctx context.Context, sessionId uuid.UUID) error {
	err := a.sessionRepository.DeleteSession(ctx, sessionId)

	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}
