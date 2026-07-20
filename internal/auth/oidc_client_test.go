package auth_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	josejwt "github.com/go-jose/go-jose/v4/jwt"
	"github.com/larssonoliver/inundated/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBeginAuthorization_BuildsAuthURLWithPKCEAndState(t *testing.T) {
	fp := newFakeProvider(t)
	defer fp.Close()

	client := auth.NewOIDCClientWithConfig(fp.testConfig())

	req, err := client.BeginAuthorization("my-state-value")
	require.NoError(t, err)

	require.NotEmpty(t, req.CodeVerifier)
	assert.GreaterOrEqual(t, len(req.CodeVerifier), 43, "code_verifier should be at least 43 chars per RFC 7636")

	u, err := url.Parse(req.Uri)
	require.NoError(t, err, "authorization Uri should be a valid URL")

	q := u.Query()
	assert.Equal(t, "my-state-value", q.Get("state"))
	assert.Equal(t, "test-client-id", q.Get("client_id"))
	assert.Equal(t, "https://app.example.com/auth/callback", q.Get("redirect_uri"))
	assert.Equal(t, "code", q.Get("response_type"))
	assert.Equal(t, "S256", q.Get("code_challenge_method"))
	assert.NotEmpty(t, q.Get("code_challenge"))
	assert.Contains(t, q.Get("scope"), "openid")
	assert.True(t, strings.HasPrefix(req.Uri, fp.issuer),
		"authorization URL %q should be rooted at the provider's issuer %q", req.Uri, fp.issuer)
}

func TestBeginAuthorization_DifferentCallsGetDifferentVerifiers(t *testing.T) {
	fp := newFakeProvider(t)
	defer fp.Close()

	client := auth.NewOIDCClientWithConfig(fp.testConfig())

	req1, err := client.BeginAuthorization("state-1")
	require.NoError(t, err)
	req2, err := client.BeginAuthorization("state-2")
	require.NoError(t, err)

	assert.NotEqual(t, req1.CodeVerifier, req2.CodeVerifier,
		"expected distinct code verifiers across separate authorization attempts")
}

func TestBeginAuthorization_RejectsEmptyState(t *testing.T) {
	fp := newFakeProvider(t)
	defer fp.Close()

	client := auth.NewOIDCClientWithConfig(fp.testConfig())

	_, err := client.BeginAuthorization("")
	assert.Error(t, err)
}

func TestExchangeCode_Success(t *testing.T) {
	fp := newFakeProvider(t)
	defer fp.Close()

	fp.registerCode("valid-code", idTokenClaims{
		Sub:   "user-123",
		Email: "alice@example.com",
		Name:  "Alice Example",
	})

	client := auth.NewOIDCClientWithConfig(fp.testConfig())

	identity, err := client.ExchangeCode(context.Background(), "valid-code", "some-verifier")
	require.NoError(t, err)

	assert.Equal(t, auth.OIDCIdentity{
		Sub:   "user-123",
		Email: "alice@example.com",
		Name:  "Alice Example",
	}, identity)
}

func TestExchangeCode_UnknownCodeFails(t *testing.T) {
	fp := newFakeProvider(t)
	defer fp.Close()

	client := auth.NewOIDCClientWithConfig(fp.testConfig())

	_, err := client.ExchangeCode(context.Background(), "never-registered", "some-verifier")
	assert.Error(t, err)
}

func TestExchangeCode_RejectsEmptyArgs(t *testing.T) {
	fp := newFakeProvider(t)
	defer fp.Close()

	client := auth.NewOIDCClientWithConfig(fp.testConfig())

	_, err := client.ExchangeCode(context.Background(), "", "verifier")
	assert.Error(t, err, "expected error for empty code")

	_, err = client.ExchangeCode(context.Background(), "code", "")
	assert.Error(t, err, "expected error for empty codeVerifier")
}

func TestExchangeCode_RejectsExpiredIDToken(t *testing.T) {
	fp := newFakeProvider(t)
	defer fp.Close()
	fp.expireTokensImmediately = true

	fp.registerCode("expired-code", idTokenClaims{Sub: "user-1"})

	client := auth.NewOIDCClientWithConfig(fp.testConfig())

	_, err := client.ExchangeCode(context.Background(), "expired-code", "verifier")
	assert.Error(t, err)
}

func TestExchangeCode_RejectsBadSignature(t *testing.T) {
	fp := newFakeProvider(t)
	defer fp.Close()
	fp.badSignature = true

	fp.registerCode("tampered-code", idTokenClaims{Sub: "user-1"})

	client := auth.NewOIDCClientWithConfig(fp.testConfig())

	_, err := client.ExchangeCode(context.Background(), "tampered-code", "verifier")
	assert.Error(t, err)
}

func TestExchangeCode_RejectsWrongClientSecret(t *testing.T) {
	fp := newFakeProvider(t)
	defer fp.Close()

	fp.registerCode("some-code", idTokenClaims{Sub: "user-1"})

	cfg := fp.testConfig()
	cfg.ClientSecret = "wrong-secret"
	client := auth.NewOIDCClientWithConfig(cfg)

	_, err := client.ExchangeCode(context.Background(), "some-code", "verifier")
	assert.Error(t, err)
}

func TestNewOIDCClientWithConfig_DiscoveryFailureIsRetriedOnNextCall(t *testing.T) {
	// Point at an issuer with nothing listening; discovery should fail on
	// BeginAuthorization but a client pointed at a now-healthy provider
	// afterwards should still work (i.e. we don't wedge into a permanently
	// broken state after one failed discovery attempt).
	client := auth.NewOIDCClientWithConfig(auth.OIDCClientConfig{
		IssuerURL:   "http://127.0.0.1:1", // nothing listens here
		ClientID:    "test-client-id",
		RedirectURL: "https://app.example.com/auth/callback",
		Scopes:      []string{"openid"},
		HTTPTimeout: 500 * time.Millisecond,
	})

	_, err := client.BeginAuthorization("state")
	require.Error(t, err, "expected discovery against an unreachable issuer to fail")

	fp := newFakeProvider(t)
	defer fp.Close()

	// Re-point the same client instance at a live provider and confirm it
	// recovers rather than caching the earlier failure forever.
	client.Cfg = fp.testConfig()

	_, err = client.BeginAuthorization("state")
	require.NoError(t, err, "expected discovery to succeed after pointing at a live provider")
}

// fakeProvider is a minimal, in-process stand-in for a real OIDC provider.
// It serves the discovery document, a JWKS, and a token endpoint that hands
// back a signed ID token for whatever "code" it's given, letting tests drive
// OIDCClientImpl end-to-end without any network access.
type fakeProvider struct {
	t *testing.T

	server *httptest.Server
	key    *rsa.PrivateKey
	keyID  string

	issuer string

	// nextCodeIdentity maps an authorization code to the identity/claims
	// the token endpoint should mint an ID token for. Tests populate this
	// before calling ExchangeCode.
	nextCodeIdentity map[string]idTokenClaims

	// wantClientSecret, if set, causes the token endpoint to reject
	// requests presenting a different client_secret.
	wantClientSecret string

	// expireTokensImmediately mints already-expired ID tokens, for testing
	// expiry handling.
	expireTokensImmediately bool

	// badSignature corrupts the returned ID token's signature.
	badSignature bool
}

type idTokenClaims struct {
	Sub   string
	Email string
	Name  string
}

func newFakeProvider(t *testing.T) *fakeProvider {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "generating RSA key")

	fp := &fakeProvider{
		t:                t,
		key:              key,
		keyID:            "test-key-1",
		nextCodeIdentity: map[string]idTokenClaims{},
		wantClientSecret: "test-client-secret",
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", fp.handleDiscovery)
	mux.HandleFunc("/jwks", fp.handleJWKS)
	mux.HandleFunc("/token", fp.handleToken)
	mux.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not used in tests", http.StatusNotImplemented)
	})

	fp.server = httptest.NewServer(mux)
	fp.issuer = fp.server.URL

	return fp
}

func (fp *fakeProvider) Close() {
	fp.server.Close()
}

func (fp *fakeProvider) handleDiscovery(w http.ResponseWriter, r *http.Request) {
	doc := map[string]any{
		"issuer":                                fp.issuer,
		"authorization_endpoint":                fp.issuer + "/authorize",
		"token_endpoint":                        fp.issuer + "/token",
		"jwks_uri":                              fp.issuer + "/jwks",
		"id_token_signing_alg_values_supported": []string{"RS256"},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(doc)
}

func (fp *fakeProvider) handleJWKS(w http.ResponseWriter, r *http.Request) {
	jwk := jose.JSONWebKey{
		Key:       &fp.key.PublicKey,
		KeyID:     fp.keyID,
		Algorithm: string(jose.RS256),
		Use:       "sig",
	}
	set := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{jwk}}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(set)
}

// registerCode arranges for the given authorization code to yield an ID
// token asserting the given claims when exchanged.
func (fp *fakeProvider) registerCode(code string, claims idTokenClaims) {
	fp.nextCodeIdentity[code] = claims
}

func (fp *fakeProvider) handleToken(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	if fp.wantClientSecret != "" && r.Form.Get("client_secret") != fp.wantClientSecret {
		http.Error(w, "invalid client credentials", http.StatusUnauthorized)
		return
	}

	// PKCE check: verify code_verifier hashes to whatever challenge would
	// have been sent. We don't track the original challenge in this fake
	// (BeginAuthorization talks to the real /authorize semantics via the
	// oauth2 library, which we don't exercise here), so we just require a
	// non-empty verifier, matching how the token endpoint would 400 on a
	// missing one.
	if r.Form.Get("code_verifier") == "" {
		http.Error(w, "missing code_verifier", http.StatusBadRequest)
		return
	}

	code := r.Form.Get("code")
	claims, ok := fp.nextCodeIdentity[code]
	if !ok {
		http.Error(w, "unknown code", http.StatusBadRequest)
		return
	}

	rawIDToken := fp.mintIDToken(claims)

	resp := map[string]any{
		"access_token": "fake-access-token",
		"token_type":   "Bearer",
		"expires_in":   3600,
		"id_token":     rawIDToken,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (fp *fakeProvider) mintIDToken(claims idTokenClaims) string {
	fp.t.Helper()

	signer, err := jose.NewSigner(jose.SigningKey{
		Algorithm: jose.RS256,
		Key:       fp.key,
	}, (&jose.SignerOptions{}).WithHeader("kid", fp.keyID).WithType("JWT"))
	require.NoError(fp.t, err, "creating signer")

	now := time.Now()
	exp := now.Add(time.Hour)
	if fp.expireTokensImmediately {
		exp = now.Add(-time.Hour)
	}

	builder := josejwt.Signed(signer).Claims(map[string]any{
		"iss":   fp.issuer,
		"sub":   claims.Sub,
		"aud":   "test-client-id",
		"exp":   exp.Unix(),
		"iat":   now.Unix(),
		"email": claims.Email,
		"name":  claims.Name,
	})

	raw, err := builder.Serialize()
	require.NoError(fp.t, err, "serializing id_token")

	if fp.badSignature {
		raw = corruptSignature(raw)
	}

	return raw
}

// corruptSignature flips bytes in the signature segment of a compact JWS so
// signature verification fails, without otherwise touching header/payload.
func corruptSignature(compact string) string {
	// Compact JWS is header.payload.signature (base64url segments).
	dot := len(compact) - 1
	for dot >= 0 && compact[dot] != '.' {
		dot--
	}
	if dot < 0 {
		return compact
	}
	sig := compact[dot+1:]
	sigBytes, err := base64.RawURLEncoding.DecodeString(sig)
	if err != nil || len(sigBytes) == 0 {
		return compact
	}
	sigBytes[0] ^= 0xFF
	return compact[:dot+1] + base64.RawURLEncoding.EncodeToString(sigBytes)
}

// testConfig returns a Config pointed at this fake provider.
func (fp *fakeProvider) testConfig() auth.OIDCClientConfig {
	return auth.OIDCClientConfig{
		IssuerURL:    fp.issuer,
		ClientID:     "test-client-id",
		ClientSecret: fp.wantClientSecret,
		RedirectURL:  "https://app.example.com/auth/callback",
		Scopes:       []string{"openid", "profile", "email"},
		HTTPTimeout:  5 * time.Second,
	}
}
