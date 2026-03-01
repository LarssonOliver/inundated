package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

// Config holds all runtime configuration for the application.
// Extend this struct to add new configuration values.
type Config struct {
	// Server
	Host string
	Port int

	// Database
	DatabaseURL string

	// Observability
	LogLevel string
}

// Option is a functional option for configuring the loader itself.
type Option func(*loader)

// WithArgs overrides the CLI arguments (defaults to os.Args[1:]).
// Useful for testing.
func WithArgs(args []string) Option {
	return func(l *loader) { l.args = args }
}

// WithEnvLookup overrides the environment variable lookup function.
// Useful for testing.
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

// envOr returns the env var value if set, otherwise the provided default.
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

// func (l *loader) envOrBool(key string, def bool) bool {
// 	if v, ok := l.envLookup(key); ok && v != "" {
// 		if b, err := strconv.ParseBool(v); err == nil {
// 			return b
// 		}
// 	}
// 	return def
// }

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

	// ------------------------------------------------------------------ //

	// Override the default Usage so -help / --help prints our custom page.
	fs.Usage = func() { printHelp(fs) }

	if err := fs.Parse(l.args); err != nil {
		// flag.ContinueOnError returns flag.ErrHelp when -help is passed;
		// the caller can check for this if needed.
		return nil, fmt.Errorf("config: parse error: %w", err)
	}

	cfg := &Config{
		Host:        *host,
		Port:        *port,
		DatabaseURL: *databaseURL,
		LogLevel:    *logLevel,
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate performs basic semantic validation after all values are resolved.
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
	return nil
}

// printHelp writes a nicely formatted help page to stderr.
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
