package contract_test

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type parameterFile map[string]struct {
	Schema struct {
		Pattern string `yaml:"pattern"`
	} `yaml:"schema"`
}

func TestIntervalParameterPattern(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)

	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
	raw, err := os.ReadFile(filepath.Join(repoRoot, "openapi", "parameters", "shared.yaml"))
	require.NoError(t, err)

	var params parameterFile
	require.NoError(t, yaml.Unmarshal(raw, &params))

	pattern := params["interval"].Schema.Pattern
	require.NotEmpty(t, pattern)

	re, err := regexp.Compile(pattern)
	require.NoError(t, err)

	valid := []string{
		"2024-01-01T00:00:00Z/2024-01-31T23:59:59Z",
		"2024-01-01T00:00:00Z/P30D",
		"P30D/2024-01-31T23:59:59Z",
	}

	invalid := []string{
		"2024-01-01/2024-01-31",
		"P1D/P2D",
		"PT/2024-01-31T23:59:59Z",
		"foo/bar",
	}

	for _, v := range valid {
		require.Truef(t, re.MatchString(v), "expected interval pattern to match %q", v)
	}

	for _, v := range invalid {
		require.Falsef(t, re.MatchString(v), "expected interval pattern to reject %q", v)
	}
}
