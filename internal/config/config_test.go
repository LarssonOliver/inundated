package config_test

import (
	"flag"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/larssonoliver/inundated/internal/config"
)

// fakeEnv returns a WithEnvLookup option backed by a plain map.
func fakeEnv(m map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		v, ok := m[key]
		return v, ok
	}
}

func TestDefaults(t *testing.T) {
	cfg, err := config.Load(config.WithArgs(nil), config.WithEnvLookup(fakeEnv(nil)))
	assert.NoError(t, err)
	assert.Equal(t, cfg.Host, "0.0.0.0")
	assert.Equal(t, cfg.Port, 8080)
	assert.Equal(t, cfg.DatabaseURL, "in-memory")
	assert.Equal(t, cfg.LogLevel, "info")
}

func TestEnvVarOverridesDefault(t *testing.T) {
	env := map[string]string{
		"HOST":         "1.1.1.1",
		"PORT":         "9090",
		"DATABASE_URL": "postgres://user:pass@localhost:5432/db",
		"LOG_LEVEL":    "warn",
	}
	cfg, err := config.Load(config.WithArgs(nil), config.WithEnvLookup(fakeEnv(env)))
	assert.NoError(t, err)
	assert.Equal(t, cfg.Host, "1.1.1.1")
	assert.Equal(t, cfg.Port, 9090)
	assert.Equal(t, cfg.DatabaseURL, "postgres://user:pass@localhost:5432/db")
	assert.Equal(t, cfg.LogLevel, "warn")
}

func TestCLIFlagOverridesEnv(t *testing.T) {
	env := map[string]string{"PORT": "9090"}
	args := []string{"-port=3000"}
	cfg, err := config.Load(config.WithArgs(args), config.WithEnvLookup(fakeEnv(env)))
	assert.NoError(t, err)
	assert.Equal(t, cfg.Port, 3000)
}

func TestValidationRejectsInvalidPort(t *testing.T) {
	_, err := config.Load(
		config.WithArgs([]string{"-port=99999"}),
		config.WithEnvLookup(fakeEnv(nil)),
	)
	assert.Error(t, err)
}

func TestValidationRejectsInvalidLogLevel(t *testing.T) {
	_, err := config.Load(
		config.WithArgs([]string{"-log-level=verbose"}),
		config.WithEnvLookup(fakeEnv(nil)),
	)
	assert.Error(t, err)
}

func TestValidationRejectsEmptyDatabaseUrl(t *testing.T) {
	_, err := config.Load(
		config.WithArgs([]string{"-database-url="}),
		config.WithEnvLookup(fakeEnv(nil)),
	)
	assert.Error(t, err)
}

func TestHelpFlag(t *testing.T) {
	_, err := config.Load(config.WithArgs([]string{"-help"}), config.WithEnvLookup(fakeEnv(nil)))
	assert.ErrorIs(t, err, flag.ErrHelp)
}

func TestOIDCDisabledByDefault(t *testing.T) {
	cfg, err := config.Load(config.WithArgs(nil), config.WithEnvLookup(fakeEnv(nil)))
	assert.NoError(t, err)
	assert.False(t, cfg.OIDC.Enabled())
	assert.Empty(t, cfg.OIDC.IssuerURL)
}

func TestOIDCConfigFromEnv(t *testing.T) {
	env := map[string]string{
		"OIDC_ISSUER_URL":    "https://issuer.example.com",
		"OIDC_CLIENT_ID":     "client-abc",
		"OIDC_CLIENT_SECRET": "secret-xyz",
		"OIDC_REDIRECT_URL":  "https://app.example.com/api/auth/callback",
		"OIDC_SCOPES":        "openid,profile,email,groups",
		"OIDC_HTTP_TIMEOUT":  "15s",
	}
	cfg, err := config.Load(config.WithArgs(nil), config.WithEnvLookup(fakeEnv(env)))
	assert.NoError(t, err)
	assert.True(t, cfg.OIDC.Enabled())
	assert.Equal(t, "https://issuer.example.com", cfg.OIDC.IssuerURL)
	assert.Equal(t, "client-abc", cfg.OIDC.ClientID)
	assert.Equal(t, "secret-xyz", cfg.OIDC.ClientSecret)
	assert.Equal(t, "https://app.example.com/api/auth/callback", cfg.OIDC.RedirectURL)
	assert.Equal(t, []string{"openid", "profile", "email", "groups"}, cfg.OIDC.Scopes)
	assert.Equal(t, 15*time.Second, cfg.OIDC.HTTPTimeout)
}

func TestOIDCDefaultsScopesAndTimeout(t *testing.T) {
	env := map[string]string{
		"OIDC_ISSUER_URL":    "https://issuer.example.com",
		"OIDC_CLIENT_ID":     "client-abc",
		"OIDC_CLIENT_SECRET": "secret-xyz",
		"OIDC_REDIRECT_URL":  "https://app.example.com/api/auth/callback",
	}
	cfg, err := config.Load(config.WithArgs(nil), config.WithEnvLookup(fakeEnv(env)))
	assert.NoError(t, err)
	assert.Equal(t, []string{"openid", "profile", "email"}, cfg.OIDC.Scopes)
	assert.Equal(t, 10*time.Second, cfg.OIDC.HTTPTimeout)
}

func TestOIDCPartialConfigRejected(t *testing.T) {
	base := map[string]string{
		"OIDC_ISSUER_URL":    "https://issuer.example.com",
		"OIDC_CLIENT_ID":     "client-abc",
		"OIDC_CLIENT_SECRET": "secret-xyz",
		"OIDC_REDIRECT_URL":  "https://app.example.com/api/auth/callback",
	}
	for _, missing := range []string{"OIDC_CLIENT_ID", "OIDC_CLIENT_SECRET", "OIDC_REDIRECT_URL"} {
		t.Run("missing_"+missing, func(t *testing.T) {
			env := map[string]string{}
			for k, v := range base {
				if k != missing {
					env[k] = v
				}
			}
			_, err := config.Load(config.WithArgs(nil), config.WithEnvLookup(fakeEnv(env)))
			assert.Error(t, err)
		})
	}
}

func TestCSRFAuthKeyEmptyByDefault(t *testing.T) {
	cfg, err := config.Load(config.WithArgs(nil), config.WithEnvLookup(fakeEnv(nil)))
	assert.NoError(t, err)
	assert.Empty(t, cfg.CSRFAuthKey)
}

func TestCSRFAuthKeyFromEnv(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef" // 32 bytes
	env := map[string]string{"CSRF_AUTH_KEY": key}
	cfg, err := config.Load(config.WithArgs(nil), config.WithEnvLookup(fakeEnv(env)))
	assert.NoError(t, err)
	assert.Equal(t, key, cfg.CSRFAuthKey)
}

func TestCSRFAuthKeyWrongLengthRejected(t *testing.T) {
	env := map[string]string{"CSRF_AUTH_KEY": "too-short"}
	_, err := config.Load(config.WithArgs(nil), config.WithEnvLookup(fakeEnv(env)))
	assert.Error(t, err)
}

func TestOIDCInvalidTimeoutFallsBackToDefault(t *testing.T) {
	env := map[string]string{
		"OIDC_ISSUER_URL":    "https://issuer.example.com",
		"OIDC_CLIENT_ID":     "client-abc",
		"OIDC_CLIENT_SECRET": "secret-xyz",
		"OIDC_REDIRECT_URL":  "https://app.example.com/api/auth/callback",
		"OIDC_HTTP_TIMEOUT":  "not-a-duration",
	}
	cfg, err := config.Load(config.WithArgs(nil), config.WithEnvLookup(fakeEnv(env)))
	assert.NoError(t, err)
	assert.Equal(t, 10*time.Second, cfg.OIDC.HTTPTimeout)
}
