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

// Each user-scoped model must have a repository contract isolation subtest,
// named "<memory|postgres>ScopeIsolation", proving cross-owner access is denied.
// This fails if a user-scoped resource is added without scoping its queries.
func TestUserScopedModelsHaveIsolationCoverage(t *testing.T) {
	for modelName := range userScopedModels {
		t.Run(modelName, func(t *testing.T) {
			require.Truef(t, isolationSubtestExists(t, modelName),
				"Test%sRepositoryContract needs a t.Run(repoName+\"ScopeIsolation\", …) subtest "+
					"proving cross-owner access is denied", modelName)
		})
	}
}

// isolationSubtestExists reports whether Test<model>RepositoryContract in this
// package registers a subtest named "<repoName>ScopeIsolation".
func isolationSubtestExists(t *testing.T, modelName string) bool {
	t.Helper()

	entries, err := os.ReadDir(".")
	require.NoError(t, err)

	fnName := "Test" + modelName + "RepositoryContract"

	fset := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, entry.Name(), nil, 0)
		require.NoError(t, err)

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name.Name != fnName {
				continue
			}
			var found bool
			ast.Inspect(fn, func(n ast.Node) bool {
				bl, ok := n.(*ast.BasicLit)
				if ok && bl.Kind == token.STRING && strings.Contains(bl.Value, "ScopeIsolation") {
					found = true
				}
				return !found
			})
			if found {
				return true
			}
		}
	}
	return false
}

func isPointerToUUID(expr ast.Expr) bool {
	star, ok := expr.(*ast.StarExpr)
	if !ok {
		return false
	}
	sel, ok := star.X.(*ast.SelectorExpr)
	return ok && sel.Sel.Name == "UUID"
}
