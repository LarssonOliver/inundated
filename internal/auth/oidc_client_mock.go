package auth

import "context"

var _ OIDCClient = (*OIDCClientMock)(nil)

type OIDCClientMock struct {
	BeginAuthorizationFn func(state string) (OIDCAuthorizationRequest, error)
	ExchangeCodeFn       func(ctx context.Context, code string, codeVerifier string) (OIDCIdentity, error)
}

func NewOIDCClientMock() *OIDCClientMock {
	return &OIDCClientMock{}
}

// BeginAuthorization implements [OIDCClient].
func (o *OIDCClientMock) BeginAuthorization(state string) (OIDCAuthorizationRequest, error) {
	return o.BeginAuthorizationFn(state)
}

// ExchangeCode implements [OIDCClient].
func (o *OIDCClientMock) ExchangeCode(ctx context.Context, code string, codeVerifier string) (OIDCIdentity, error) {
	return o.ExchangeCodeFn(ctx, code, codeVerifier)
}
