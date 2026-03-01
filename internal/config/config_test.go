package config_test

import (
	"flag"
	"testing"

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
	assert.Equal(t, cfg.LogLevel, "info")
}

func TestEnvVarOverridesDefault(t *testing.T) {
	env := map[string]string{
		"HOST":      "1.1.1.1",
		"PORT":      "9090",
		"LOG_LEVEL": "warn",
	}
	cfg, err := config.Load(config.WithArgs(nil), config.WithEnvLookup(fakeEnv(env)))
	assert.NoError(t, err)
	assert.Equal(t, cfg.Host, "1.1.1.1")
	assert.Equal(t, cfg.Port, 9090)
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

func TestHelpFlag(t *testing.T) {
	_, err := config.Load(config.WithArgs([]string{"-help"}), config.WithEnvLookup(fakeEnv(nil)))
	assert.ErrorIs(t, err, flag.ErrHelp)
}
