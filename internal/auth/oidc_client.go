package auth

import "context"

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
}

func NewOIDCClient() *OIDCClientImpl {
	return &OIDCClientImpl{}
}

// BeginAuthorization implements [OIDCClient].
func (o *OIDCClientImpl) BeginAuthorization(state string) (OIDCAuthorizationRequest, error) {
	panic("unimplemented")
}

// ExchangeCode implements [OIDCClient].
func (o *OIDCClientImpl) ExchangeCode(ctx context.Context, code string, codeVerifier string) (OIDCIdentity, error) {
	panic("unimplemented")
}
