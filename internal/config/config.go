package config

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Host          string
	Port          int
	DatabaseURL   string
	LogLevel      string
	PublicBaseURL string
	OIDC          OIDCConfig
	CSRFAuthKey   string

	// DisableUserRegistration, when true, rejects logins from OIDC identities
	// that don't already have an account. Existing users are unaffected. Flip
	// this on once every expected user has been enrolled.
	DisableUserRegistration bool
}

const csrfAuthKeyLen = 32

type OIDCConfig struct {
	IssuerURL    string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	HTTPTimeout  time.Duration
}

const oidcCallbackPath = "/api/auth/callback"

func (c OIDCConfig) Enabled() bool {
	return c.IssuerURL != ""
}

var defaultOIDCScopes = []string{"openid", "profile", "email"}

const defaultOIDCHTTPTimeout = 10 * time.Second

type Option func(*loader)

func WithArgs(args []string) Option {
	return func(l *loader) { l.args = args }
}

func WithEnvLookup(fn func(string) (string, bool)) Option {
	return func(l *loader) { l.envLookup = fn }
}

type loader struct {
	args      []string
	envLookup func(string) (string, bool)
}

// Load resolves configuration in priority order (highest → lowest):
//  1. CLI flags
//  2. Environment variables
//  3. Default values
func Load(opts ...Option) (*Config, error) {
	l := &loader{
		args:      os.Args[1:],
		envLookup: os.LookupEnv,
	}
	for _, o := range opts {
		o(l)
	}
	return l.load()
}

func (l *loader) envOr(key, def string) string {
	if v, ok := l.envLookup(key); ok && v != "" {
		return v
	}
	return def
}

func (l *loader) envOrInt(key string, def int) int {
	if v, ok := l.envLookup(key); ok && v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func (l *loader) envOrBool(key string, def bool) bool {
	if v, ok := l.envLookup(key); ok && v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func (l *loader) envOrDuration(key string, def time.Duration) time.Duration {
	if v, ok := l.envLookup(key); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func (l *loader) load() (*Config, error) {
	fs := flag.NewFlagSet("inundated", flag.ContinueOnError)

	// ------------------------------------------------------------------ //
	// Fields: register each flag with its env-var-aware default.         //
	// Convention: flag name = lowercase-hyphen, env = UPPER_SNAKE_CASE   //
	// ------------------------------------------------------------------ //

	host := fs.String("host", l.envOr("HOST", "0.0.0.0"),
		"Server listen host (env: HOST)")

	port := fs.Int("port", l.envOrInt("PORT", 8080),
		"Server listen port (env: PORT)")

	databaseURL := fs.String("database-url", l.envOr("DATABASE_URL", "in-memory"),
		"PostgreSQL connection string (env: DATABASE_URL)")

	logLevel := fs.String("log-level", l.envOr("LOG_LEVEL", "info"),
		"Log level: debug|info|warn|error (env: LOG_LEVEL)")

	oidcIssuerURL := fs.String("oidc-issuer-url", l.envOr("OIDC_ISSUER_URL", ""),
		"OIDC provider issuer URL; enables authentication when set (env: OIDC_ISSUER_URL)")

	oidcClientID := fs.String("oidc-client-id", l.envOr("OIDC_CLIENT_ID", ""),
		"OIDC client ID (env: OIDC_CLIENT_ID)")

	oidcClientSecret := fs.String("oidc-client-secret", l.envOr("OIDC_CLIENT_SECRET", ""),
		"OIDC client secret (env: OIDC_CLIENT_SECRET)")

	publicBaseURL := fs.String("public-base-url", l.envOr("PUBLIC_BASE_URL", ""),
		"Public origin of this app (scheme + host), e.g. https://inundated.example.com; required when OIDC is enabled (env: PUBLIC_BASE_URL)")

	oidcScopes := fs.String("oidc-scopes", l.envOr("OIDC_SCOPES", strings.Join(defaultOIDCScopes, ",")),
		"Comma-separated OIDC scopes to request (env: OIDC_SCOPES)")

	oidcHTTPTimeout := fs.Duration("oidc-http-timeout", l.envOrDuration("OIDC_HTTP_TIMEOUT", defaultOIDCHTTPTimeout),
		"Timeout for OIDC discovery, JWKS, and token requests (env: OIDC_HTTP_TIMEOUT)")

	csrfAuthKey := fs.String("csrf-auth-key", l.envOr("CSRF_AUTH_KEY", ""),
		fmt.Sprintf("%d-byte key for signing CSRF tokens; generated ephemerally if unset (env: CSRF_AUTH_KEY)", csrfAuthKeyLen))

	disableUserRegistration := fs.Bool("disable-user-registration", l.envOrBool("DISABLE_USER_REGISTRATION", false),
		"Reject logins from OIDC identities without an existing account (env: DISABLE_USER_REGISTRATION)")

	// ------------------------------------------------------------------ //

	fs.Usage = func() { printHelp(fs) }

	if err := fs.Parse(l.args); err != nil {
		// flag.ContinueOnError returns flag.ErrHelp when -help is passed;
		// the caller can check for this if needed.
		return nil, fmt.Errorf("config: parse error: %w", err)
	}

	baseURL := strings.TrimRight(*publicBaseURL, "/")

	cfg := &Config{
		Host:          *host,
		Port:          *port,
		DatabaseURL:   *databaseURL,
		LogLevel:      *logLevel,
		PublicBaseURL: baseURL,
		OIDC: OIDCConfig{
			IssuerURL:    *oidcIssuerURL,
			ClientID:     *oidcClientID,
			ClientSecret: *oidcClientSecret,
			RedirectURL:  redirectURLFor(baseURL),
			Scopes:       splitAndTrim(*oidcScopes),
			HTTPTimeout:  *oidcHTTPTimeout,
		},
		CSRFAuthKey:             *csrfAuthKey,
		DisableUserRegistration: *disableUserRegistration,
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("config: port %d is out of range [1, 65535]", c.Port)
	}
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[c.LogLevel] {
		return fmt.Errorf("config: log-level %q is not one of debug|info|warn|error", c.LogLevel)
	}
	if c.DatabaseURL == "" {
		return fmt.Errorf("config: database-url must not be empty")
	}
	if c.PublicBaseURL != "" {
		if err := validatePublicBaseURL(c.PublicBaseURL); err != nil {
			return err
		}
	}
	if err := c.OIDC.validate(); err != nil {
		return err
	}
	if c.OIDC.Enabled() && c.PublicBaseURL == "" {
		return fmt.Errorf("config: public-base-url is required when oidc-issuer-url is set")
	}
	if c.CSRFAuthKey != "" && len(c.CSRFAuthKey) != csrfAuthKeyLen {
		return fmt.Errorf("config: csrf-auth-key must be exactly %d bytes, got %d", csrfAuthKeyLen, len(c.CSRFAuthKey))
	}
	return nil
}

func (c *OIDCConfig) validate() error {
	if !c.Enabled() {
		return nil
	}
	var missing []string
	if c.ClientID == "" {
		missing = append(missing, "oidc-client-id")
	}
	if c.ClientSecret == "" {
		missing = append(missing, "oidc-client-secret")
	}
	if len(missing) > 0 {
		return fmt.Errorf("config: oidc-issuer-url is set but these are missing: %s", strings.Join(missing, ", "))
	}
	for _, required := range []string{"openid", "email"} {
		if !slices.Contains(c.Scopes, required) {
			return fmt.Errorf("config: oidc-scopes must include %q (got %q)", required, strings.Join(c.Scopes, ","))
		}
	}
	if c.HTTPTimeout <= 0 {
		return fmt.Errorf("config: oidc-http-timeout must be positive, got %s", c.HTTPTimeout)
	}
	return nil
}

func redirectURLFor(baseURL string) string {
	if baseURL == "" {
		return ""
	}
	return baseURL + oidcCallbackPath
}

func validatePublicBaseURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("config: public-base-url is not a valid URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("config: public-base-url must start with http:// or https://")
	}
	if u.Host == "" {
		return fmt.Errorf("config: public-base-url must include a host")
	}
	if strings.Trim(u.Path, "/") != "" || u.RawQuery != "" || u.Fragment != "" {
		return fmt.Errorf("config: public-base-url must be a bare origin (scheme + host), e.g. https://inundated.example.com")
	}
	return nil
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func printHelp(fs *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, `
Usage: %s [flags]

Flags can also be set via environment variables (shown in parentheses).
CLI flags take precedence over environment variables.

Flags:
`, fs.Name())
	fs.PrintDefaults()
	fmt.Fprintln(os.Stderr)
}
