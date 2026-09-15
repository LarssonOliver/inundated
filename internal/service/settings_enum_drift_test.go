package service_test

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/service"
	"github.com/stretchr/testify/require"
)

// generatedEnumValues reads the literal values of every const declared with
// the given type name (e.g. "DateFormat") in internal/api/model.gen.go - the
// openapi-codegen output for openapi/schemas/settings.yaml's enum lists.
//
// This lets the test below prove validateSettings' hand-rolled allow-lists
// stay in lockstep with the generated code (and therefore the OpenAPI spec)
// without internal/service importing internal/api: only this test file does,
// which doesn't create the production dependency service -> api that would
// invert the layering the rest of the codebase deliberately keeps (api/
// handlers depend on service + convert to/from its own wire types, not the
// other way around).
func generatedEnumValues(t *testing.T, typeName string) []string {
	t.Helper()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filepath.Join("..", "api", "model.gen.go"), nil, 0)
	require.NoError(t, err)

	var values []string
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok || vs.Type == nil {
				continue
			}
			ident, ok := vs.Type.(*ast.Ident)
			if !ok || ident.Name != typeName {
				continue
			}
			for _, v := range vs.Values {
				lit, ok := v.(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					continue
				}
				unquoted, err := strconv.Unquote(lit.Value)
				require.NoError(t, err)
				values = append(values, unquoted)
			}
		}
	}
	require.NotEmptyf(t, values, "no %q const values found in internal/api/model.gen.go - has the type been renamed or regenerated differently?", typeName)
	return values
}

// TestSettingsValidation_AcceptsEveryGeneratedEnumValue proves that every
// enum value openapi-codegen generated from openapi/schemas/settings.yaml is
// accepted by the service-layer validation. If a new value is added to the
// YAML (regenerating internal/api/model.gen.go) but the service's map isn't
// updated to match, this test fails - catching the exact drift risk that
// hand-rolled, separately-maintained allow-lists create.
func TestSettingsValidation_AcceptsEveryGeneratedEnumValue(t *testing.T) {
	validBaseline := func() model.Settings {
		return model.Settings{
			WeekStartDay:   model.WeekStartMonday,
			Timezone:       "UTC",
			DurationFormat: model.DurationFormatLong,
			TimeFormat:     model.TimeFormat24h,
			DateFormat:     model.DateFormatISO,
		}
	}

	tests := []struct {
		enumTypeName string
		withValue    func(s model.Settings, value string) model.Settings
	}{
		{"WeekStartDay", func(s model.Settings, v string) model.Settings { s.WeekStartDay = v; return s }},
		{"DurationFormat", func(s model.Settings, v string) model.Settings { s.DurationFormat = v; return s }},
		{"TimeFormat", func(s model.Settings, v string) model.Settings { s.TimeFormat = v; return s }},
		{"DateFormat", func(s model.Settings, v string) model.Settings { s.DateFormat = v; return s }},
	}

	for _, tt := range tests {
		t.Run(tt.enumTypeName, func(t *testing.T) {
			for _, value := range generatedEnumValues(t, tt.enumTypeName) {
				repo := &repository.RepoMock{
					UpdateSettingsFn: func(ctx context.Context, scope model.OwnerScope, settings model.Settings) (model.Settings, error) {
						return settings, nil
					},
				}

				_, err := service.NewService(repo).UpdateSettings(context.Background(), tt.withValue(validBaseline(), value))
				require.NoErrorf(t, err, "generated %s value %q was rejected by validateSettings - the service-layer allow-list is missing it", tt.enumTypeName, value)
			}
		})
	}
}
