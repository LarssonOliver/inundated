package auth

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type OIDCClient interface {
	BeginAuthorization(state string) (OIDCAuthorizationRequest, error)
	ExchangeCode(ctx context.Context, code string, codeVerifier string) (OIDCIdentity, error)
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

	mu       sync.Mutex
	provider *oidc.Provider
	oauthCfg oauth2.Config
	verifier *oidc.IDTokenVerifier
}

func NewOIDCClient() *OIDCClientImpl {
	// TODO: Parse config
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

	Scopes []string

	// HTTPTimeout bounds calls to the discovery, JWKS, and token endpoints.
	HTTPTimeout time.Duration
}

// BeginAuthorization implements [OIDCClient]. It generates a fresh PKCE
// code_verifier and returns the provider authorization URL (bound to the
// caller-supplied state and the S256 code_challenge derived from the
// verifier). The caller is responsible for persisting the returned
// CodeVerifier (e.g. server-side, keyed by state) and supplying it back to
// ExchangeCode.
func (o *OIDCClientImpl) BeginAuthorization(state string) (OIDCAuthorizationRequest, error) {
	if state == "" {
		return OIDCAuthorizationRequest{}, errors.New("state must not be empty")
	}

	oauthCfg, _, err := o.ready(context.Background())
	if err != nil {
		return OIDCAuthorizationRequest{}, err
	}

	verifier := oauth2.GenerateVerifier()
	authURL := oauthCfg.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))

	return OIDCAuthorizationRequest{
		Uri:          authURL,
		CodeVerifier: verifier,
	}, nil
}

// ExchangeCode implements [OIDCClient]. It exchanges the authorization code
// for tokens (presenting codeVerifier to satisfy PKCE), then verifies the
// returned ID token's signature, issuer, audience and expiry before
// extracting identity claims from it.
func (o *OIDCClientImpl) ExchangeCode(ctx context.Context, code string, codeVerifier string) (OIDCIdentity, error) {
	if code == "" {
		return OIDCIdentity{}, errors.New("code must not be empty")
	}
	if codeVerifier == "" {
		return OIDCIdentity{}, errors.New("codeVerifier must not be empty")
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

// ready performs (and caches) OIDC discovery against the issuer. Safe for
// concurrent use; discovery is retried on subsequent calls if it previously
// failed (e.g. the provider was briefly unreachable at startup).
func (o *OIDCClientImpl) ready(ctx context.Context) (oauth2.Config, *oidc.IDTokenVerifier, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.provider != nil {
		return o.oauthCfg, o.verifier, nil
	}

	ctx, cancel := context.WithTimeout(ctx, o.Cfg.HTTPTimeout)
	defer cancel()

	provider, err := oidc.NewProvider(ctx, o.Cfg.IssuerURL)
	if err != nil {
		return oauth2.Config{}, nil, fmt.Errorf("discovering OIDC provider %q: %w", o.Cfg.IssuerURL, err)
	}

	oauthCfg := oauth2.Config{
		ClientID:     o.Cfg.ClientID,
		ClientSecret: o.Cfg.ClientSecret,
		RedirectURL:  o.Cfg.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       o.Cfg.Scopes,
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: o.Cfg.ClientID})

	o.provider = provider
	o.oauthCfg = oauthCfg
	o.verifier = verifier

	return oauthCfg, verifier, nil
}
