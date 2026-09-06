package contract_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// userScopedModels maps every model that carries a per-user owner to its table.
// A resource here must be adopted by the first user in CreateUserAdoptingOrphans
// in every repository implementation. TestUserScopedModelsAreCovered fails if a
// model gains or loses a UserId field without this map being updated.
var userScopedModels = map[string]string{
	"Project":  "projects",
	"Tag":      "tags",
	"Timespan": "timespans",
}

func TestUserScopedModelsAreCovered(t *testing.T) {
	want := make([]string, 0, len(userScopedModels))
	for name := range userScopedModels {
		want = append(want, name)
	}

	require.ElementsMatch(t, want, modelsWithUserIDField(t, "../../model"),
		"a model's `UserId *uuid.UUID` field changed; update userScopedModels and "+
			"CreateUserAdoptingOrphans in every repository implementation")
}

// modelsWithUserIDField returns the names of struct types in dir that declare a
// field `UserId *uuid.UUID`.
func modelsWithUserIDField(t *testing.T, dir string) []string {
	t.Helper()

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	fset := token.NewFileSet()
	var names []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}

		file, err := parser.ParseFile(fset, filepath.Join(dir, entry.Name()), nil, 0)
		require.NoError(t, err)

		ast.Inspect(file, func(n ast.Node) bool {
			ts, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				return true
			}
			for _, field := range st.Fields.List {
				if len(field.Names) == 1 && field.Names[0].Name == "UserId" && isPointerToUUID(field.Type) {
					names = append(names, ts.Name.Name)
				}
			}
			return true
		})
	}
	return names
}

func isPointerToUUID(expr ast.Expr) bool {
	star, ok := expr.(*ast.StarExpr)
	if !ok {
		return false
	}
	sel, ok := star.X.(*ast.SelectorExpr)
	return ok && sel.Sel.Name == "UUID"
}
