package auth

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"golang.org/x/sync/singleflight"
)

type OIDCClient interface {
	BeginAuthorization(state string, nonce string) (OIDCAuthorizationRequest, error)
	ExchangeCode(ctx context.Context, code string, codeVerifier string, expectedNonce string) (OIDCIdentity, error)
}

type OIDCIdentity struct {
	Sub   string
	Email string
	Name  string
}

type OIDCAuthorizationRequest struct {
	Uri          string
	CodeVerifier string
}

var _ OIDCClient = (*OIDCClientImpl)(nil)

type OIDCClientImpl struct {
	Cfg OIDCClientConfig

	discovery singleflight.Group
	mu        sync.RWMutex
	provider  *oidc.Provider
	oauthCfg  oauth2.Config
	verifier  *oidc.IDTokenVerifier
}

// NewOIDCClient returns a client with no provider configured. It is only useful
// where authentication is disabled (userless mode); any call that needs the
// provider will fail. Use [NewOIDCClientWithConfig] to configure OIDC.
func NewOIDCClient() *OIDCClientImpl {
	return &OIDCClientImpl{}
}

func NewOIDCClientWithConfig(cfg OIDCClientConfig) *OIDCClientImpl {
	return &OIDCClientImpl{Cfg: cfg}
}

type OIDCClientConfig struct {
	IssuerURL string

	ClientID     string
	ClientSecret string

	// RedirectURL must exactly match a redirect URI registered with the provider.
	RedirectURL string
	Scopes      []string
	HTTPTimeout time.Duration
}

// BeginAuthorization implements [OIDCClient].
func (o *OIDCClientImpl) BeginAuthorization(state string, nonce string) (OIDCAuthorizationRequest, error) {
	if state == "" {
		return OIDCAuthorizationRequest{}, errors.New("state must not be empty")
	}
	if nonce == "" {
		return OIDCAuthorizationRequest{}, errors.New("nonce must not be empty")
	}

	oauthCfg, _, err := o.ready(context.Background())
	if err != nil {
		return OIDCAuthorizationRequest{}, err
	}

	verifier := oauth2.GenerateVerifier()
	authURL := oauthCfg.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier), oidc.Nonce(nonce))

	return OIDCAuthorizationRequest{
		Uri:          authURL,
		CodeVerifier: verifier,
	}, nil
}

// ExchangeCode implements [OIDCClient].
func (o *OIDCClientImpl) ExchangeCode(ctx context.Context, code string, codeVerifier string, expectedNonce string) (OIDCIdentity, error) {
	if code == "" {
		return OIDCIdentity{}, errors.New("code must not be empty")
	}
	if codeVerifier == "" {
		return OIDCIdentity{}, errors.New("codeVerifier must not be empty")
	}
	if expectedNonce == "" {
		return OIDCIdentity{}, errors.New("expectedNonce must not be empty")
	}

	oauthCfg, verifier, err := o.ready(ctx)
	if err != nil {
		return OIDCIdentity{}, err
	}

	token, err := oauthCfg.Exchange(ctx, code, oauth2.VerifierOption(codeVerifier))
	if err != nil {
		return OIDCIdentity{}, fmt.Errorf("exchanging authorization code: %w", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return OIDCIdentity{}, errors.New("token response did not contain an id_token")
	}

	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return OIDCIdentity{}, fmt.Errorf("verifying id_token: %w", err)
	}

	if idToken.Nonce != expectedNonce {
		return OIDCIdentity{}, errors.New("id_token nonce does not match the authentication request")
	}

	var claims struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return OIDCIdentity{}, fmt.Errorf("parsing id_token claims: %w", err)
	}

	return OIDCIdentity{
		Sub:   idToken.Subject,
		Email: claims.Email,
		Name:  claims.Name,
	}, nil
}

func (o *OIDCClientImpl) ready(ctx context.Context) (oauth2.Config, *oidc.IDTokenVerifier, error) {
	o.mu.RLock()
	cfg, verifier, ready := o.oauthCfg, o.verifier, o.provider != nil
	o.mu.RUnlock()
	if ready {
		return cfg, verifier, nil
	}

	if _, err, _ := o.discovery.Do("discover", func() (any, error) {
		return nil, o.discover(ctx)
	}); err != nil {
		return oauth2.Config{}, nil, err
	}

	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.oauthCfg, o.verifier, nil
}

func (o *OIDCClientImpl) discover(ctx context.Context) error {
	o.mu.RLock()
	already := o.provider != nil
	o.mu.RUnlock()
	if already {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), o.Cfg.HTTPTimeout)
	defer cancel()

	provider, err := oidc.NewProvider(ctx, o.Cfg.IssuerURL)
	if err != nil {
		return fmt.Errorf("discovering OIDC provider %q: %w", o.Cfg.IssuerURL, err)
	}

	oauthCfg := oauth2.Config{
		ClientID:     o.Cfg.ClientID,
		ClientSecret: o.Cfg.ClientSecret,
		RedirectURL:  o.Cfg.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       o.Cfg.Scopes,
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: o.Cfg.ClientID})

	o.mu.Lock()
	o.provider = provider
	o.oauthCfg = oauthCfg
	o.verifier = verifier
	o.mu.Unlock()

	return nil
}
