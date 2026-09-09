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
		"PUBLIC_BASE_URL":    "https://app.example.com",
		"OIDC_SCOPES":        "openid,profile,email,groups",
		"OIDC_HTTP_TIMEOUT":  "15s",
	}
	cfg, err := config.Load(config.WithArgs(nil), config.WithEnvLookup(fakeEnv(env)))
	assert.NoError(t, err)
	assert.True(t, cfg.OIDC.Enabled())
	assert.Equal(t, "https://issuer.example.com", cfg.OIDC.IssuerURL)
	assert.Equal(t, "client-abc", cfg.OIDC.ClientID)
	assert.Equal(t, "secret-xyz", cfg.OIDC.ClientSecret)
	assert.Equal(t, "https://app.example.com", cfg.PublicBaseURL)
	assert.Equal(t, "https://app.example.com/api/auth/callback", cfg.OIDC.RedirectURL)
	assert.Equal(t, []string{"openid", "profile", "email", "groups"}, cfg.OIDC.Scopes)
	assert.Equal(t, 15*time.Second, cfg.OIDC.HTTPTimeout)
}

func TestPublicBaseURLTrailingSlashStripped(t *testing.T) {
	env := map[string]string{
		"OIDC_ISSUER_URL":    "https://issuer.example.com",
		"OIDC_CLIENT_ID":     "client-abc",
		"OIDC_CLIENT_SECRET": "secret-xyz",
		"PUBLIC_BASE_URL":    "https://app.example.com/",
	}
	cfg, err := config.Load(config.WithArgs(nil), config.WithEnvLookup(fakeEnv(env)))
	assert.NoError(t, err)
	assert.Equal(t, "https://app.example.com", cfg.PublicBaseURL)
	assert.Equal(t, "https://app.example.com/api/auth/callback", cfg.OIDC.RedirectURL)
}

func TestPublicBaseURLEmptyByDefault(t *testing.T) {
	cfg, err := config.Load(config.WithArgs(nil), config.WithEnvLookup(fakeEnv(nil)))
	assert.NoError(t, err)
	assert.Empty(t, cfg.PublicBaseURL)
	assert.Empty(t, cfg.OIDC.RedirectURL)
}

func TestPublicBaseURLInvalidRejected(t *testing.T) {
	for _, v := range []string{"not a url", "ftp://app.example.com", "https://app.example.com/sub", "app.example.com"} {
		t.Run(v, func(t *testing.T) {
			env := map[string]string{"PUBLIC_BASE_URL": v}
			_, err := config.Load(config.WithArgs(nil), config.WithEnvLookup(fakeEnv(env)))
			assert.Error(t, err)
		})
	}
}

func TestOIDCDefaultsScopesAndTimeout(t *testing.T) {
	env := map[string]string{
		"OIDC_ISSUER_URL":    "https://issuer.example.com",
		"OIDC_CLIENT_ID":     "client-abc",
		"OIDC_CLIENT_SECRET": "secret-xyz",
		"PUBLIC_BASE_URL":    "https://app.example.com",
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
		"PUBLIC_BASE_URL":    "https://app.example.com",
	}
	for _, missing := range []string{"OIDC_CLIENT_ID", "OIDC_CLIENT_SECRET", "PUBLIC_BASE_URL"} {
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

func TestOIDCRequiresOpenIDAndEmailScopes(t *testing.T) {
	base := map[string]string{
		"OIDC_ISSUER_URL":    "https://issuer.example.com",
		"OIDC_CLIENT_ID":     "client-abc",
		"OIDC_CLIENT_SECRET": "secret-xyz",
		"PUBLIC_BASE_URL":    "https://app.example.com",
	}

	for _, scopes := range []string{"openid,profile", "profile,email", "openid"} {
		t.Run(scopes, func(t *testing.T) {
			env := map[string]string{"OIDC_SCOPES": scopes}
			for k, v := range base {
				env[k] = v
			}
			_, err := config.Load(config.WithArgs(nil), config.WithEnvLookup(fakeEnv(env)))
			assert.Error(t, err, "OIDC needs both openid and email scopes to identify and create users")
		})
	}

	t.Run("openid,email accepted", func(t *testing.T) {
		env := map[string]string{"OIDC_SCOPES": "openid,email"}
		for k, v := range base {
			env[k] = v
		}
		_, err := config.Load(config.WithArgs(nil), config.WithEnvLookup(fakeEnv(env)))
		assert.NoError(t, err)
	})
}

func TestDisableUserRegistrationDefaultsFalse(t *testing.T) {
	cfg, err := config.Load(config.WithArgs(nil), config.WithEnvLookup(fakeEnv(nil)))
	assert.NoError(t, err)
	assert.False(t, cfg.DisableUserRegistration)
}

func TestDisableUserRegistrationFromEnv(t *testing.T) {
	env := map[string]string{"DISABLE_USER_REGISTRATION": "true"}
	cfg, err := config.Load(config.WithArgs(nil), config.WithEnvLookup(fakeEnv(env)))
	assert.NoError(t, err)
	assert.True(t, cfg.DisableUserRegistration)
}

func TestDisableUserRegistrationFromFlag(t *testing.T) {
	cfg, err := config.Load(
		config.WithArgs([]string{"-disable-user-registration"}),
		config.WithEnvLookup(fakeEnv(nil)),
	)
	assert.NoError(t, err)
	assert.True(t, cfg.DisableUserRegistration)
}

func TestDisableUserRegistrationInvalidEnvFallsBackToFalse(t *testing.T) {
	env := map[string]string{"DISABLE_USER_REGISTRATION": "not-a-bool"}
	cfg, err := config.Load(config.WithArgs(nil), config.WithEnvLookup(fakeEnv(env)))
	assert.NoError(t, err)
	assert.False(t, cfg.DisableUserRegistration)
}

func TestOIDCInvalidTimeoutFallsBackToDefault(t *testing.T) {
	env := map[string]string{
		"OIDC_ISSUER_URL":    "https://issuer.example.com",
		"OIDC_CLIENT_ID":     "client-abc",
		"OIDC_CLIENT_SECRET": "secret-xyz",
		"PUBLIC_BASE_URL":    "https://app.example.com",
		"OIDC_HTTP_TIMEOUT":  "not-a-duration",
	}
	cfg, err := config.Load(config.WithArgs(nil), config.WithEnvLookup(fakeEnv(env)))
	assert.NoError(t, err)
	assert.Equal(t, 10*time.Second, cfg.OIDC.HTTPTimeout)
}

func TestOIDCRejectsNonPositiveTimeout(t *testing.T) {
	base := map[string]string{
		"OIDC_ISSUER_URL":    "https://issuer.example.com",
		"OIDC_CLIENT_ID":     "client-abc",
		"OIDC_CLIENT_SECRET": "secret-xyz",
		"PUBLIC_BASE_URL":    "https://app.example.com",
	}
	// These parse fine as durations, so they slip past the fall-back-to-default
	// path, but context.WithTimeout(ctx, <=0) fires immediately and wedges every
	// discovery/token call.
	for _, timeout := range []string{"0", "0s", "-1s"} {
		t.Run(timeout, func(t *testing.T) {
			env := map[string]string{"OIDC_HTTP_TIMEOUT": timeout}
			for k, v := range base {
				env[k] = v
			}
			_, err := config.Load(config.WithArgs(nil), config.WithEnvLookup(fakeEnv(env)))
			assert.Error(t, err)
		})
	}
}
