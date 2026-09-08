package auth

import "context"

var _ OIDCClient = (*OIDCClientMock)(nil)

type OIDCClientMock struct {
	BeginAuthorizationFn func(state string, nonce string) (OIDCAuthorizationRequest, error)
	ExchangeCodeFn       func(ctx context.Context, code string, codeVerifier string, expectedNonce string) (OIDCIdentity, error)
}

func NewOIDCClientMock() *OIDCClientMock {
	return &OIDCClientMock{}
}

// BeginAuthorization implements [OIDCClient].
func (o *OIDCClientMock) BeginAuthorization(state string, nonce string) (OIDCAuthorizationRequest, error) {
	return o.BeginAuthorizationFn(state, nonce)
}

// ExchangeCode implements [OIDCClient].
func (o *OIDCClientMock) ExchangeCode(ctx context.Context, code string, codeVerifier string, expectedNonce string) (OIDCIdentity, error) {
	return o.ExchangeCodeFn(ctx, code, codeVerifier, expectedNonce)
}
