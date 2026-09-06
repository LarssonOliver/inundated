# Data Model User-Scoping Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Confine every project / tag / timespan repository operation to the current user, or to the unowned (`user_id IS NULL`) pool when the server runs without authentication.

**Architecture:** A `model.OwnerScope` value (a `UserScope(id)` or `UnownedScope()`) is resolved once per request from the context by a service helper and passed as an explicit parameter to every scoped repository method. Postgres enforces it with a single `user_id IS NOT DISTINCT FROM $N` predicate (NULL-safe, one code path for both modes); the memory store uses an equivalent `matchesScope` helper. No schema change, no API change.

**Tech Stack:** Go, `github.com/google/uuid`, `github.com/jackc/pgx/v5` (+ `pgxpool`), `github.com/pashagolub/pgxmock/v4`, `github.com/stretchr/testify`, testcontainers-go (postgres), golang-migrate.

**Spec:** `docs/superpowers/specs/2026-09-06-data-model-user-scoping-design.md`

## Global Constraints

- **Userless mode must keep working.** No current user ⇒ operate on `user_id IS NULL` rows. This is the zero value of `OwnerScope`.
- **No schema change.** `user_id` stays nullable. Do not add migrations. Note: the `idx_{projects,tags,timespans}_user_id` btree indexes from migration `0006` do **not** accelerate the scoped `ListX`/aggregate queries — `user_id IS NOT DISTINCT FROM $N` has no btree strategy, so those queries seq-scan. Acceptable now (not a regression — list queries seq-scanned before this branch; tables are small); a follow-up index is needed before the data grows. See the design doc's "Known limitation / follow-up".
- **No API surface change.** Do not touch `openapi/`, `internal/api/`, or `internal/api/handlers/`. Scope is derived server-side only.
- **Cross-user access returns 404** — falls out of the existing `ErrNotFound` path. Do not add 403 responses or an owner field.
- **`scope model.OwnerScope` is the parameter immediately after `ctx`** on every scoped repository method.
- **Not scoped:** `UserRepository` (incl. `CreateUserAdoptingOrphans`), `SessionRepository`, `LoginStateRepository`.
- Follow existing repository patterns: soft-delete `deleted_at IS NULL` predicate, `model.ErrNotFound` / `model.ErrInvalidReference` / `model.ErrInvalidArgument`, pgxmock unit tests, memory+postgres contract tests.
- Every commit message ends with the two trailer lines used on this branch:
  ```
  Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
  Claude-Session: https://claude.ai/code/session_01UzTPnp1FQiGpFJAG1JdbXW
  ```
- Verification commands: `go build ./...`, `go vet ./...`, `go test ./...`, `golangci-lint run ./internal/...`. Container tests need Docker running.

---

## File Structure

**New files:**
- `internal/model/scope.go` — the `OwnerScope` type.
- `internal/model/scope_test.go` — `OwnerScope` unit tests.
- `internal/service/scope.go` — the `ownerScope(ctx)` resolver.
- `internal/service/scope_test.go` — black-box test that services thread the ctx-derived scope to the repository.
- `internal/repository/memory/scope.go` — the `matchesScope` helper.

**Modified — production:**
- `internal/repository/repository.go` — add `scope` to `TagRepository`, `ProjectRepository`, `TimespanRepository`, `ProjectStatsRepository`.
- `internal/repository/repository_mock.go` — `RepoMock` field + method signatures.
- `internal/repository/postgres/mock.go` — `MockRepository` method signatures.
- `internal/repository/postgres/{tag,project,timespan}.go` — predicates, `tagsInScope` guard.
- `internal/repository/postgres/helpers.go` — (optional) nothing required; predicate text is inlined.
- `internal/repository/memory/{tags,projects,timespans}.go` — `matchesScope` filtering, `tagsExist` gains scope.
- `internal/service/{tags,projects,timespans,project_stats}.go` — resolve scope, pass it down.

**Modified — tests:**
- `internal/repository/contract_test/*.go` — thread scope; new isolation subtests.
- `internal/repository/postgres/{tag,project,timespan}_test.go` — query/arg expectations; cross-scope cases.
- `internal/repository/memory/{tags,projects,timespans}_test.go` — thread scope; cross-scope cases.
- `internal/service/{tags,projects,timespans,project_stats}_test.go` — `RepoMock` `...Fn` signatures.

---

## Task 1: `model.OwnerScope`

**Files:**
- Create: `internal/model/scope.go`
- Test: `internal/model/scope_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces:
  - `type OwnerScope struct { … }` (opaque; only the constructors and `UserID()` are exported)
  - `func UserScope(id uuid.UUID) OwnerScope`
  - `func UnownedScope() OwnerScope`
  - `func (s OwnerScope) UserID() *uuid.UUID` — returns `nil` for the unowned scope, a non-nil pointer otherwise. The zero value `OwnerScope{}` equals `UnownedScope()`.

- [ ] **Step 1: Write the failing test**

Create `internal/model/scope_test.go`:

```go
package model_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/stretchr/testify/require"
)

func TestUserScope_CarriesID(t *testing.T) {
	id := uuid.New()

	got := model.UserScope(id).UserID()

	require.NotNil(t, got)
	require.Equal(t, id, *got)
}

func TestUnownedScope_HasNoID(t *testing.T) {
	require.Nil(t, model.UnownedScope().UserID())
}

func TestOwnerScope_ZeroValueIsUnowned(t *testing.T) {
	require.Nil(t, model.OwnerScope{}.UserID())
}

func TestUserScope_DoesNotAliasCallerVariable(t *testing.T) {
	id := uuid.New()
	scope := model.UserScope(id)

	// mutating the original variable must not change the scope
	id = uuid.New()

	require.NotEqual(t, id, *scope.UserID())
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/model/ -run 'Scope' -v`
Expected: FAIL — `undefined: model.UserScope` / `model.UnownedScope` / `model.OwnerScope`.

- [ ] **Step 3: Write the implementation**

Create `internal/model/scope.go`:

```go
package model

import "github.com/google/uuid"

// OwnerScope identifies whose resources an operation may see or modify. The zero
// value is the unowned scope, used when the server runs without authentication
// (rows with user_id IS NULL).
type OwnerScope struct {
	userID *uuid.UUID
}

// UserScope scopes operations to a single authenticated user.
func UserScope(id uuid.UUID) OwnerScope {
	return OwnerScope{userID: &id}
}

// UnownedScope scopes operations to rows with no owner.
func UnownedScope() OwnerScope {
	return OwnerScope{}
}

// UserID returns the scoped user id, or nil for the unowned scope.
func (s OwnerScope) UserID() *uuid.UUID {
	return s.userID
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/model/ -run 'Scope' -v`
Expected: PASS (4 tests).

- [ ] **Step 5: Vet and commit**

```bash
go vet ./internal/model/
git add internal/model/scope.go internal/model/scope_test.go
git commit -m "feat: add model.OwnerScope

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
Claude-Session: https://claude.ai/code/session_01UzTPnp1FQiGpFJAG1JdbXW"
```

---

## Task 2: Thread `OwnerScope` through the repository layer (no enforcement)

This is a wide, purely mechanical change: add `scope model.OwnerScope` to four repository interfaces, both implementations (which **ignore** it for now), both mock types, wire the service layer through a new `ownerScope(ctx)` resolver, and update every test call site to compile. **No behavior changes** — every existing assertion stays exactly as-is and the full suite stays green.

**Files:**
- Create: `internal/service/scope.go`, `internal/service/scope_test.go`
- Modify:
  - `internal/repository/repository.go`
  - `internal/repository/repository_mock.go`
  - `internal/repository/postgres/mock.go`
  - `internal/repository/postgres/tag.go`, `project.go`, `timespan.go`
  - `internal/repository/memory/tags.go`, `projects.go`, `timespans.go`
  - `internal/service/tags.go`, `projects.go`, `timespans.go`, `project_stats.go`
  - `internal/repository/contract_test/repository_test.go`, `repository_tag_test.go`, `repository_project_test.go`, `repository_timespan_test.go`, `repository_project_stats_test.go`, `repository_user_adoption_pg_test.go`
  - `internal/repository/postgres/tag_test.go`, `project_test.go`, `timespan_test.go`
  - `internal/repository/memory/tags_test.go`, `projects_test.go`, `timespans_test.go`
  - `internal/service/tags_test.go`, `projects_test.go`, `timespans_test.go`, `project_stats_test.go`

**Interfaces:**
- Consumes: `model.OwnerScope`, `model.UserScope`, `model.UnownedScope` (Task 1).
- Produces:
  - Repository methods, new signatures (`scope` inserted after `ctx`):
    ```go
    // TagRepository
    GetTag(ctx context.Context, scope model.OwnerScope, id uuid.UUID) (model.Tag, error)
    ListTags(ctx context.Context, scope model.OwnerScope, params model.PaginationParams) (model.Page[model.Tag], error)
    CreateTag(ctx context.Context, scope model.OwnerScope, tag model.Tag) (model.Tag, error)
    UpdateTag(ctx context.Context, scope model.OwnerScope, tag model.Tag) (model.Tag, error)
    DeleteTag(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error
    // ProjectRepository — same shape
    GetProject(ctx context.Context, scope model.OwnerScope, id uuid.UUID) (model.Project, error)
    ListProjects(ctx context.Context, scope model.OwnerScope, params model.PaginationParams) (model.Page[model.Project], error)
    CreateProject(ctx context.Context, scope model.OwnerScope, project model.Project) (model.Project, error)
    UpdateProject(ctx context.Context, scope model.OwnerScope, project model.Project) (model.Project, error)
    DeleteProject(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error
    // TimespanRepository
    GetTimespan(ctx context.Context, scope model.OwnerScope, id uuid.UUID) (model.Timespan, error)
    ListTimespans(ctx context.Context, scope model.OwnerScope, params model.PaginationParams) (model.Page[model.Timespan], error)
    CreateTimespan(ctx context.Context, scope model.OwnerScope, timespan model.Timespan) (model.Timespan, error)
    UpdateTimespan(ctx context.Context, scope model.OwnerScope, timespan model.Timespan) (model.Timespan, error)
    DeleteTimespan(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error
    GetTotalDurationByTags(ctx context.Context, scope model.OwnerScope, tagIds []uuid.UUID) (time.Duration, error)
    // ProjectStatsRepository
    AggregateTimeSpentByTagsAndBuckets(ctx context.Context, scope model.OwnerScope, tagIds []uuid.UUID, buckets []model.BucketRange) ([]model.BucketValue, error)
    ```
  - `func ownerScope(ctx context.Context) model.OwnerScope` in package `service` (unexported).
  - Contract tests gain package-level `var testScope = model.UserScope(uuid.MustParse("11111111-1111-1111-1111-111111111111"))` in `repository_test.go`.
  - `seedTags` gains a `scope model.OwnerScope` parameter: `seedTags(t, ctx, repo, scope, n)`.

- [ ] **Step 1: Update the repository interfaces**

In `internal/repository/repository.go`, insert `scope model.OwnerScope` after `ctx context.Context` in every method of `TagRepository`, `ProjectRepository`, `TimespanRepository`, and `ProjectStatsRepository`. Use the exact signatures from the **Produces** block above. Leave `UserRepository`, `SessionRepository`, `LoginStateRepository` untouched.

- [ ] **Step 2: Update `RepoMock`**

In `internal/repository/repository_mock.go`:
- Add `scope model.OwnerScope` after `ctx` in each of these function-field types: `CreateProjectFn`, `GetProjectFn`, `ListProjectFn`, `UpdateProjectFn`, (`DeleteProjectFn`), `CreateTagFn`, `GetTagFn`, `ListTagFn`, `UpdateTagFn`, (`DeleteTagFn`), `CreateTimespanFn`, `GetTimespanFn`, `ListTimespanFn`, `UpdateTimespanFn`, `GetTotalDurationByTagsFn`, `AggregateTimeSpentByTagsAndBucketsFn`, and the `Delete*Fn` fields.
- Update the corresponding method wrappers to accept `scope` and forward it, e.g.:
  ```go
  func (t *RepoMock) GetTag(ctx context.Context, scope model.OwnerScope, id uuid.UUID) (model.Tag, error) {
  	return t.GetTagFn(ctx, scope, id)
  }
  ```
- Do **not** touch the `*User*Fn` fields/methods.

- [ ] **Step 3: Update `postgres.MockRepository`**

In `internal/repository/postgres/mock.go`, add `scope model.OwnerScope` after `ctx` in the signatures of `GetTag`, `ListTags`, `CreateTag`, `UpdateTag`, `DeleteTag`, `GetProject`, `ListProjects`, `CreateProject`, `UpdateProject`, `DeleteProject`, `GetTimespan`, `ListTimespans`, `CreateTimespan`, `UpdateTimespan`, `DeleteTimespan`, `GetTotalDurationByTags`, `AggregateTimeSpentByTagsAndBuckets`. Bodies keep calling `m.Called(...)` — add `scope` to the `m.Called` argument list too. Leave user/session/login-state methods untouched.

- [ ] **Step 4: Update the postgres implementation signatures (ignore scope)**

In `internal/repository/postgres/tag.go`, `project.go`, `timespan.go`: add `scope model.OwnerScope` after `ctx` to every exported method listed in the **Produces** block. **Do not change any SQL.** Add `_ = scope` at the top of each method body to keep the linter quiet, OR (preferred) leave it — an unused function parameter is not flagged by `go vet` or the configured linters. Keep the private helpers (`setProjectTags`, `projectTagIds`, `setTimespanTags`, `timespanTagIds`) unchanged for now.

- [ ] **Step 5: Update the memory implementation signatures (ignore scope)**

In `internal/repository/memory/tags.go`, `projects.go`, `timespans.go`: same — add `scope model.OwnerScope` after `ctx` to every exported method. Do not change logic. The private `tagsExist(ctx, tagIds)` helper: leave it for now (Task 3 threads scope into it); its callers in `CreateProject`/`UpdateProject`/`CreateTimespan`/`UpdateTimespan` still compile.

- [ ] **Step 6: Verify production build**

Run: `go build ./...`
Expected: PASS. (Test packages will not compile yet — that is expected and fixed in the following steps.)

- [ ] **Step 7: Create the `ownerScope` resolver**

Create `internal/service/scope.go`:

```go
package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

// ownerScope derives the resource ownership scope for the current request:
// the authenticated user, or the unowned pool when the server runs without auth.
func ownerScope(ctx context.Context) model.OwnerScope {
	if u, ok := model.GetCurrentUserFromContext(ctx); ok && u.Id != uuid.Nil {
		return model.UserScope(u.Id)
	}
	return model.UnownedScope()
}
```

- [ ] **Step 8: Wire the service layer**

In each of `internal/service/tags.go`, `projects.go`, `timespans.go`, `project_stats.go`: at the top of every method that calls the repository, add `scope := ownerScope(ctx)`, and pass `scope` as the second argument to every `s.repository.*` call for tags/projects/timespans/aggregates. Worked example — `internal/service/tags.go` `GetTag`:

```go
func (s *ServiceImpl) GetTag(ctx context.Context, id uuid.UUID, includes *TagServiceGetIncludes) (model.Tag, error) {
	scope := ownerScope(ctx)

	tag, err := s.repository.GetTag(ctx, scope, id)
	if err != nil {
		return model.Tag{}, model.ErrNotFound
	}

	if includes != nil && includes.TotalTime {
		totalTime, err := s.repository.GetTotalDurationByTags(ctx, scope, []uuid.UUID{tag.Id})
		if err != nil {
			return model.Tag{}, model.ErrNotFound
		}
		tag.TotalTime = &totalTime
	}

	return tag, nil
}
```

Apply the same transformation to: `ListTags`, `CreateTag`, `UpdateTag`, `DeleteTag`; `GetProject` (incl. its `GetTotalDurationByTags` call), `ListProjects`, `CreateProject`, `UpdateProject`, `DeleteProject`; `GetTimespan`, `ListTimespans`, `CreateTimespan`, `UpdateTimespan`, `DeleteTimespan`; and in `project_stats.go` `GetProjectStats` pass `scope` to both `s.repository.GetProject(...)` and `s.repository.AggregateTimeSpentByTagsAndBuckets(...)`.

- [ ] **Step 9: Fix `internal/service/*_test.go` to compile**

In `tags_test.go`, `projects_test.go`, `timespans_test.go`, `project_stats_test.go`: every `RepoMock` literal sets `...Fn` closures. Add `scope model.OwnerScope` after `ctx` in each closure's parameter list. Do not change closure bodies or assertions. Worked example:

```go
// before
GetTagFn: func(ctx context.Context, id uuid.UUID) (model.Tag, error) { ... }
// after
GetTagFn: func(ctx context.Context, scope model.OwnerScope, id uuid.UUID) (model.Tag, error) { ... }
```

- [ ] **Step 10: Add the service scope-threading test**

Create `internal/service/scope_test.go`:

```go
package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/service"
	"github.com/stretchr/testify/require"
)

func TestService_ThreadsUserScopeFromContext(t *testing.T) {
	var got model.OwnerScope
	repo := &repository.RepoMock{
		ListTagFn: func(ctx context.Context, scope model.OwnerScope, params model.PaginationParams) (model.Page[model.Tag], error) {
			got = scope
			return model.Page[model.Tag]{}, nil
		},
	}
	s := service.NewService(repo)

	user := model.User{Id: uuid.New()}
	_, err := s.ListTags(model.SetUserInContext(context.Background(), user), model.DefaultPaginationParams())
	require.NoError(t, err)
	require.NotNil(t, got.UserID())
	require.Equal(t, user.Id, *got.UserID())
}

func TestService_UsesUnownedScopeWithoutUser(t *testing.T) {
	var got model.OwnerScope
	repo := &repository.RepoMock{
		ListTagFn: func(ctx context.Context, scope model.OwnerScope, params model.PaginationParams) (model.Page[model.Tag], error) {
			got = scope
			return model.Page[model.Tag]{}, nil
		},
	}
	s := service.NewService(repo)

	_, err := s.ListTags(context.Background(), model.DefaultPaginationParams())
	require.NoError(t, err)
	require.Nil(t, got.UserID())
}
```

- [ ] **Step 11: Fix contract tests to compile**

In `internal/repository/contract_test/`:
- `repository_test.go`: add `import "github.com/google/uuid"` if not present, add
  ```go
  var testScope = model.UserScope(uuid.MustParse("11111111-1111-1111-1111-111111111111"))
  ```
  and change `seedTags` to take `scope model.OwnerScope` after `ctx`:
  ```go
  func seedTags(t *testing.T, ctx context.Context, repo repository.TagRepository, scope model.OwnerScope, n int) []uuid.UUID {
  	...
  	created, err := repo.CreateTag(ctx, scope, tag)
  	...
  }
  ```
  In `seedOrphanResources`, call `seedTags(t, ctx, repo, model.UnownedScope(), nTags)` and pass `model.UnownedScope()` to the `repo.CreateProject` / `repo.CreateTimespan` calls (orphans must stay unowned so `CreateUserAdoptingOrphans` still adopts them).
- `repository_tag_test.go`, `repository_project_test.go`, `repository_timespan_test.go`, `repository_project_stats_test.go`: pass `testScope` as the second arg to every `repo.GetTag/ListTags/CreateTag/UpdateTag/DeleteTag`, `repo.*Project*`, `repo.*Timespan*`, `repo.GetTotalDurationByTags`, `repo.AggregateTimeSpentByTagsAndBuckets` call, and `testScope` to every `seedTags(...)` call. Do not change any assertion.
- `repository_user_adoption_pg_test.go`: `seedOrphanResources` is already updated above; the `countOwnedBy` helper and adoption assertions are unchanged. If it calls `repo.CreateProject` etc. directly, pass `model.UnownedScope()`.

- [ ] **Step 12: Fix postgres unit tests to compile**

In `internal/repository/postgres/tag_test.go`, `project_test.go`, `timespan_test.go`: add `testScope` (define it once, e.g. in `helpers_test.go`: `var testScope = model.UserScope(uuid.MustParse("11111111-1111-1111-1111-111111111111"))`) as the second arg to every `repo.<Method>` call. Do **not** change any `mock.ExpectQuery` / `WithArgs` — the SQL is unchanged in this task. Run the package to confirm nothing broke behaviorally.

- [ ] **Step 13: Fix memory unit tests to compile**

In `internal/repository/memory/tags_test.go`, `projects_test.go`, `timespans_test.go`: add `testScope` (define once in an existing `_test.go` file in that package) as the second arg to every store method call. Do not change assertions.

- [ ] **Step 14: Run the full suite**

Run: `go build ./... && go vet ./... && go test ./...`
Expected: PASS — everything green, identical behavior to before this task (repositories ignore `scope`; the only new tests are `scope_test.go` in `service`).

Run: `golangci-lint run ./internal/...`
Expected: 0 issues.

- [ ] **Step 15: Commit**

```bash
git add -A
git commit -m "refactor: thread model.OwnerScope through the repository layer

Adds a scope parameter to every tag/project/timespan/stats repository
method and a service-side ownerScope(ctx) resolver. No enforcement yet:
implementations accept and ignore the scope. Behavior unchanged.

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
Claude-Session: https://claude.ai/code/session_01UzTPnp1FQiGpFJAG1JdbXW"
```

---

## Task 3: Tag scope enforcement

**Files:**
- Modify: `internal/repository/postgres/tag.go`
- Create: `internal/repository/memory/scope.go`
- Modify: `internal/repository/memory/tags.go`
- Test: `internal/repository/contract_test/repository_tag_test.go` (new isolation subtests)
- Test: `internal/repository/postgres/tag_test.go` (cross-scope cases + updated query expectations)
- Test: `internal/repository/memory/tags_test.go` (cross-scope cases)

**Interfaces:**
- Consumes: repository signatures from Task 2; `model.OwnerScope`, `model.OwnerScope.UserID()`.
- Produces:
  - `func matchesScope(userID *uuid.UUID, scope model.OwnerScope) bool` in package `memory` (`internal/repository/memory/scope.go`).
  - Tag repository methods now filter by `user_id IS NOT DISTINCT FROM <scope>` (postgres) / `matchesScope` (memory). `CreateTag` persists `user_id = scope.UserID()`. `GetTag`/`UpdateTag`/`DeleteTag` on an out-of-scope row return `model.ErrNotFound`. `ListTags` excludes out-of-scope rows from both `Data` and `TotalCount`.

- [ ] **Step 1: Write failing contract isolation subtests for tags**

In `internal/repository/contract_test/repository_tag_test.go`, inside the `run` closure (alongside the existing subtests), add:

```go
t.Run(repoName+"ScopeIsolation", func(t *testing.T) {
	repo := newRepo(t)
	scopeA := model.UserScope(uuid.New())
	scopeB := model.UserScope(uuid.New())

	tagA, err := repo.CreateTag(ctx, scopeA, model.Tag{Name: "a", Color: "#111111"})
	require.NoError(t, err)
	_, err = repo.CreateTag(ctx, scopeB, model.Tag{Name: "b", Color: "#222222"})
	require.NoError(t, err)

	// List is scoped
	pageA, err := repo.ListTags(ctx, scopeA, model.DefaultPaginationParams())
	require.NoError(t, err)
	require.Len(t, pageA.Data, 1)
	require.Equal(t, 1, pageA.TotalCount)
	require.Equal(t, tagA.Id, pageA.Data[0].Id)

	// Get across scope is a miss
	_, err = repo.GetTag(ctx, scopeB, tagA.Id)
	require.ErrorIs(t, err, model.ErrNotFound)

	// Update across scope is a miss
	tagA.Name = "hijack"
	_, err = repo.UpdateTag(ctx, scopeB, tagA)
	require.ErrorIs(t, err, model.ErrNotFound)

	// Delete across scope is a miss
	err = repo.DeleteTag(ctx, scopeB, tagA.Id)
	require.ErrorIs(t, err, model.ErrNotFound)

	// The owner still sees an untouched tag
	got, err := repo.GetTag(ctx, scopeA, tagA.Id)
	require.NoError(t, err)
	require.Equal(t, "a", got.Name)
})

t.Run(repoName+"UnownedScopeIsolation", func(t *testing.T) {
	repo := newRepo(t)
	user := model.UserScope(uuid.New())

	owned, err := repo.CreateTag(ctx, user, model.Tag{Name: "owned", Color: "#111111"})
	require.NoError(t, err)
	unowned, err := repo.CreateTag(ctx, model.UnownedScope(), model.Tag{Name: "unowned", Color: "#222222"})
	require.NoError(t, err)

	unownedPage, err := repo.ListTags(ctx, model.UnownedScope(), model.DefaultPaginationParams())
	require.NoError(t, err)
	require.Len(t, unownedPage.Data, 1)
	require.Equal(t, unowned.Id, unownedPage.Data[0].Id)

	_, err = repo.GetTag(ctx, model.UnownedScope(), owned.Id)
	require.ErrorIs(t, err, model.ErrNotFound)
	_, err = repo.GetTag(ctx, user, unowned.Id)
	require.ErrorIs(t, err, model.ErrNotFound)
})
```

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./internal/repository/contract_test/ -run 'TestTagRepositoryContract/(memory|postgres)ScopeIsolation|TestTagRepositoryContract/(memory|postgres)UnownedScopeIsolation' -v`
Expected: FAIL — `memory` fails first (List returns both tags; cross-scope Get succeeds). Postgres subtests also fail.

- [ ] **Step 3: Add the memory `matchesScope` helper**

Create `internal/repository/memory/scope.go`:

```go
package memory

import (
	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

// matchesScope reports whether a row with the given owner belongs to scope.
func matchesScope(userID *uuid.UUID, scope model.OwnerScope) bool {
	want := scope.UserID()
	if want == nil {
		return userID == nil
	}
	return userID != nil && *userID == *want
}
```

- [ ] **Step 4: Enforce scope in the memory tag store**

In `internal/repository/memory/tags.go`:
- `CreateTag`: set the owner on the stored struct:
  ```go
  newTag := model.Tag{
  	Id:     tag.Id,
  	Name:   tag.Name,
  	Color:  tag.Color,
  	UserId: scope.UserID(),
  }
  ```
- `GetTag`: after finding by id, return `model.ErrNotFound` unless `matchesScope(t.tags[idx].UserId, scope)`.
- `ListTags`: build `all` from only the tags where `matchesScope(tag.UserId, scope)`, then paginate/count as before.
- `UpdateTag`: locate with both predicates — `t.Id == tag.Id && matchesScope(t.UserId, scope)`; `ErrNotFound` if none. Preserve the owner on write: `tag.UserId = t.tags[idx].UserId` before `t.tags[idx] = tag`.
- `DeleteTag`: locate with both predicates; `ErrNotFound` if none.

- [ ] **Step 5: Run memory tag contract subtests**

Run: `go test ./internal/repository/contract_test/ -run 'TestTagRepositoryContract/memoryScopeIsolation|TestTagRepositoryContract/memoryUnownedScopeIsolation' -v`
Expected: PASS.

- [ ] **Step 6: Enforce scope in the postgres tag store**

In `internal/repository/postgres/tag.go`, add `AND user_id IS NOT DISTINCT FROM $N` to every query and bind `scope.UserID()`:
- `GetTag`:
  ```sql
  SELECT id, name, color FROM tags
  WHERE id = $1 AND deleted_at IS NULL AND user_id IS NOT DISTINCT FROM $2
  ```
  `r.db.QueryRow(ctx, q, id, scope.UserID())`.
- `ListTags` count query:
  ```sql
  SELECT COUNT(*) FROM tags WHERE deleted_at IS NULL AND user_id IS NOT DISTINCT FROM $1
  ```
  `r.db.QueryRow(ctx, countQ, scope.UserID())`.
  Data query:
  ```sql
  SELECT id, name, color FROM tags
  WHERE deleted_at IS NULL AND user_id IS NOT DISTINCT FROM $1
  ORDER BY name
  LIMIT $2 OFFSET $3
  ```
  `r.db.Query(ctx, dataQ, scope.UserID(), params.Limit, params.Offset)`.
- `CreateTag`:
  ```sql
  INSERT INTO tags (id, name, color, user_id)
  VALUES ($1, $2, $3, $4)
  RETURNING id, name, color
  ```
  `r.db.QueryRow(ctx, q, tag.Id, tag.Name, tag.Color, scope.UserID())`.
- `UpdateTag`:
  ```sql
  UPDATE tags SET name = $2, color = $3
  WHERE id = $1 AND deleted_at IS NULL AND user_id IS NOT DISTINCT FROM $4
  RETURNING id, name, color
  ```
  `r.db.QueryRow(ctx, q, tag.Id, tag.Name, tag.Color, scope.UserID())`.
- `DeleteTag`:
  ```sql
  UPDATE tags SET deleted_at = now()
  WHERE id = $1 AND deleted_at IS NULL AND user_id IS NOT DISTINCT FROM $2
  ```
  `r.db.Exec(ctx, q, id, scope.UserID())`; unchanged `RowsAffected() == 0 → ErrNotFound`.

- [ ] **Step 7: Run postgres tag contract subtests**

Run: `go test ./internal/repository/contract_test/ -run 'TestTagRepositoryContract/postgres' -v`
Expected: PASS (existing subtests still green with `testScope`; new isolation subtests green).

- [ ] **Step 8: Update postgres tag unit-test expectations**

In `internal/repository/postgres/tag_test.go`, update each `mock.ExpectQuery(...)` regex to include the new `user_id IS NOT DISTINCT FROM` clause and each `WithArgs(...)` to include `testScope.UserID()` in the right position. Add one cross-scope case per read/write:

```go
func TestGetTag_OutOfScopeIsNotFound(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMock(t)
	id := uuid.New()

	mock.ExpectQuery(`SELECT id, name, color FROM tags WHERE id = \$1 AND deleted_at IS NULL AND user_id IS NOT DISTINCT FROM \$2`).
		WithArgs(id, testScope.UserID()).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "color"}))

	_, err := repo.GetTag(ctx, testScope, id)
	require.ErrorIs(t, err, model.ErrNotFound)
}
```

- [ ] **Step 9: Add memory tag cross-scope unit cases**

In `internal/repository/memory/tags_test.go`, add a direct test:

```go
func TestMemoryStore_Tag_ScopeIsolation(t *testing.T) {
	ctx := context.Background()
	store := memory.NewMemoryStore()
	a := model.UserScope(uuid.New())
	b := model.UserScope(uuid.New())

	tag, err := store.CreateTag(ctx, a, model.Tag{Name: "x", Color: "#123456"})
	require.NoError(t, err)

	_, err = store.GetTag(ctx, b, tag.Id)
	require.ErrorIs(t, err, model.ErrNotFound)

	page, err := store.ListTags(ctx, b, model.DefaultPaginationParams())
	require.NoError(t, err)
	require.Empty(t, page.Data)
}
```

- [ ] **Step 10: Full verification for the task**

Run: `go test ./internal/repository/... ./internal/service/...`
Expected: PASS.
Run: `golangci-lint run ./internal/repository/...`
Expected: 0 issues.

- [ ] **Step 11: Commit**

```bash
git add -A
git commit -m "feat: scope tag repository operations to the owner

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
Claude-Session: https://claude.ai/code/session_01UzTPnp1FQiGpFJAG1JdbXW"
```

---

## Task 4: Project scope enforcement

Mirrors Task 3, plus a tag-ownership guard on attachment.

**Files:**
- Modify: `internal/repository/postgres/project.go`
- Modify: `internal/repository/memory/projects.go`, `internal/repository/memory/tags.go` (thread scope into `tagsExist`)
- Test: `internal/repository/contract_test/repository_project_test.go`
- Test: `internal/repository/postgres/project_test.go`
- Test: `internal/repository/memory/projects_test.go`

**Interfaces:**
- Consumes: `matchesScope` (Task 3, package `memory`), tag scoping (Task 3).
- Produces:
  - postgres private helper `func (r *PostgresStore) tagsInScope(ctx context.Context, scope model.OwnerScope, tagIds []uuid.UUID) (bool, error)` in `project.go` (also used by `timespan.go` in Task 5 — define it in `project.go`, it lives on `*PostgresStore`).
  - `setProjectTags` gains `scope model.OwnerScope` and rejects out-of-scope tag ids with `model.ErrInvalidReference`.
  - memory `tagsExist` gains `scope model.OwnerScope`: `func (t *MemoryStore) tagsExist(ctx context.Context, scope model.OwnerScope, tagIds []uuid.UUID) bool`.
  - Project repo methods scoped exactly like tags (`CreateProject` persists `user_id`, others filter, cross-scope ⇒ `ErrNotFound`). Attaching a tag owned by a different scope ⇒ `model.ErrInvalidReference` from `CreateProject`/`UpdateProject`.

- [ ] **Step 1: Write failing contract isolation subtests for projects**

In `repository_project_test.go`, inside `run`, add (analogous to Task 3 Step 1):

```go
t.Run(repoName+"ScopeIsolation", func(t *testing.T) {
	repo := newRepo(t)
	scopeA := model.UserScope(uuid.New())
	scopeB := model.UserScope(uuid.New())

	projA, err := repo.CreateProject(ctx, scopeA, model.Project{Name: "a", Color: "#111111"})
	require.NoError(t, err)
	_, err = repo.CreateProject(ctx, scopeB, model.Project{Name: "b", Color: "#222222"})
	require.NoError(t, err)

	pageA, err := repo.ListProjects(ctx, scopeA, model.DefaultPaginationParams())
	require.NoError(t, err)
	require.Len(t, pageA.Data, 1)
	require.Equal(t, 1, pageA.TotalCount)

	_, err = repo.GetProject(ctx, scopeB, projA.Id)
	require.ErrorIs(t, err, model.ErrNotFound)

	projA.Name = "hijack"
	_, err = repo.UpdateProject(ctx, scopeB, projA)
	require.ErrorIs(t, err, model.ErrNotFound)

	err = repo.DeleteProject(ctx, scopeB, projA.Id)
	require.ErrorIs(t, err, model.ErrNotFound)
})

t.Run(repoName+"CannotAttachAnotherUsersTag", func(t *testing.T) {
	repo := newRepo(t)
	scopeA := model.UserScope(uuid.New())
	scopeB := model.UserScope(uuid.New())

	tagB := seedTags(t, ctx, repo, scopeB, 1)[0]

	_, err := repo.CreateProject(ctx, scopeA, model.Project{
		Name: "a", Color: "#111111", TagIds: []uuid.UUID{tagB},
	})
	require.ErrorIs(t, err, model.ErrInvalidReference)
})

t.Run(repoName+"UnownedScopeIsolation", func(t *testing.T) {
	repo := newRepo(t)
	user := model.UserScope(uuid.New())

	owned, err := repo.CreateProject(ctx, user, model.Project{Name: "owned", Color: "#111111"})
	require.NoError(t, err)
	_, err = repo.CreateProject(ctx, model.UnownedScope(), model.Project{Name: "unowned", Color: "#222222"})
	require.NoError(t, err)

	page, err := repo.ListProjects(ctx, model.UnownedScope(), model.DefaultPaginationParams())
	require.NoError(t, err)
	require.Len(t, page.Data, 1)

	_, err = repo.GetProject(ctx, model.UnownedScope(), owned.Id)
	require.ErrorIs(t, err, model.ErrNotFound)
})
```

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./internal/repository/contract_test/ -run 'TestProjectRepositoryContract/(memory|postgres)(ScopeIsolation|CannotAttachAnotherUsersTag|UnownedScopeIsolation)' -v`
Expected: FAIL.

- [ ] **Step 3: Thread scope into memory `tagsExist`**

In `internal/repository/memory/tags.go`, change `tagsExist` to:

```go
func (t *MemoryStore) tagsExist(ctx context.Context, scope model.OwnerScope, tagIds []uuid.UUID) bool {
	for _, tagId := range tagIds {
		if _, err := t.GetTag(ctx, scope, tagId); err != nil {
			return false
		}
	}
	return true
}
```

Since `GetTag` is now scope-enforcing (Task 3), this rejects out-of-scope tags for free.

- [ ] **Step 4: Enforce scope in the memory project store**

In `internal/repository/memory/projects.go`:
- `CreateProject`: pass `scope` to `t.tagsExist(ctx, scope, project.TagIds)`; set `UserId: scope.UserID()` on `newProject`.
- `GetProject`: `ErrNotFound` unless `matchesScope(t.projects[idx].UserId, scope)`.
- `ListProjects`: filter by `matchesScope` before pagination/count.
- `UpdateProject`: pass `scope` to `t.tagsExist`; locate by `p.Id == project.Id && matchesScope(p.UserId, scope)`; preserve `project.UserId = t.projects[idx].UserId` before writing.
- `DeleteProject`: locate with both predicates.

- [ ] **Step 5: Run memory project contract subtests**

Run: `go test ./internal/repository/contract_test/ -run 'TestProjectRepositoryContract/memory' -v`
Expected: PASS.

- [ ] **Step 6: Add the postgres `tagsInScope` helper**

In `internal/repository/postgres/project.go`:

```go
// tagsInScope reports whether every id refers to a live tag owned by scope.
func (r *PostgresStore) tagsInScope(ctx context.Context, scope model.OwnerScope, tagIds []uuid.UUID) (bool, error) {
	if len(tagIds) == 0 {
		return true, nil
	}
	const q = `
		SELECT count(*) FROM tags
		WHERE id = ANY($1) AND deleted_at IS NULL AND user_id IS NOT DISTINCT FROM $2`
	var n int
	if err := r.db.QueryRow(ctx, q, tagIds, scope.UserID()).Scan(&n); err != nil {
		return false, fmt.Errorf("tagsInScope: %w", err)
	}
	return n == len(tagIds), nil
}
```

- [ ] **Step 7: Enforce scope in the postgres project store**

In `internal/repository/postgres/project.go`, apply the same predicate pattern as Task 3 Step 6 to `GetProject`, `ListProjects` (count + data), `UpdateProject`, `DeleteProject`. For `CreateProject`, add `user_id` to the INSERT column list bound to `scope.UserID()`. Change `setProjectTags` to:

```go
func (r *PostgresStore) setProjectTags(ctx context.Context, scope model.OwnerScope, projectId uuid.UUID, tagIds []uuid.UUID) error {
	ok, err := r.tagsInScope(ctx, scope, tagIds)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("setProjectTags: %w", model.ErrInvalidReference)
	}
	// ... existing DELETE + INSERT loop unchanged
}
```

Update the two `setProjectTags(...)` call sites in `CreateProject`/`UpdateProject` to pass `scope`.

- [ ] **Step 8: Run postgres project contract subtests**

Run: `go test ./internal/repository/contract_test/ -run 'TestProjectRepositoryContract/postgres' -v`
Expected: PASS.

- [ ] **Step 9: Update postgres project unit tests**

In `internal/repository/postgres/project_test.go`: update `ExpectQuery` regexes + `WithArgs` for the new predicate/arg; where `setProjectTags` runs, add a `mock.ExpectQuery(\`SELECT count\(\*\) FROM tags\`).WithArgs(...).WillReturnRows(...)` expectation returning the matching count. Add a `TestCreateProject_ForeignTagRejected` case where the count query returns fewer than `len(tagIds)` and assert `model.ErrInvalidReference`.

- [ ] **Step 10: Add memory project cross-scope unit case**

In `internal/repository/memory/projects_test.go`, add a `TestMemoryStore_Project_ScopeIsolation` analogous to Task 3 Step 9.

- [ ] **Step 11: Full verification + commit**

Run: `go test ./internal/repository/... ./internal/service/... && golangci-lint run ./internal/repository/...`
Expected: PASS, 0 issues.

```bash
git add -A
git commit -m "feat: scope project repository operations to the owner

CreateProject/UpdateProject reject attaching tags owned by another scope.

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
Claude-Session: https://claude.ai/code/session_01UzTPnp1FQiGpFJAG1JdbXW"
```

---

## Task 5: Timespan scope enforcement

Mirrors Task 4 (timespans also attach tags). `GetTotalDurationByTags` is covered in Task 6.

**Files:**
- Modify: `internal/repository/postgres/timespan.go`
- Modify: `internal/repository/memory/timespans.go`
- Test: `internal/repository/contract_test/repository_timespan_test.go`
- Test: `internal/repository/postgres/timespan_test.go`
- Test: `internal/repository/memory/timespans_test.go`

**Interfaces:**
- Consumes: `matchesScope` (memory), `tagsExist` w/ scope (memory), `(*PostgresStore).tagsInScope` (Task 4).
- Produces:
  - `setTimespanTags` gains `scope model.OwnerScope`, rejects out-of-scope tag ids with `model.ErrInvalidReference`.
  - Timespan `Get/List/Create/Update/Delete` scoped exactly like projects.

- [ ] **Step 1: Write failing contract isolation subtests for timespans**

In `repository_timespan_test.go`, inside `run`, add subtests analogous to Task 4 Step 1: `ScopeIsolation` (create A + B; List/Get/Update/Delete scoped), `CannotAttachAnotherUsersTag` (seed a tag under `scopeB`, `CreateTimespan` under `scopeA` with that tag → `model.ErrInvalidReference`), `UnownedScopeIsolation`. Use `model.Timespan{Name: "x", StartTime: <now>, EndTime: <now+1h>}` for valid rows (see existing tests in this file for the time-field pattern).

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./internal/repository/contract_test/ -run 'TestTimespanRepositoryContract/(memory|postgres)(ScopeIsolation|CannotAttachAnotherUsersTag|UnownedScopeIsolation)' -v`
Expected: FAIL.

- [ ] **Step 3: Enforce scope in the memory timespan store**

In `internal/repository/memory/timespans.go`:
- `CreateTimespan`: `t.tagsExist(ctx, scope, timespan.TagIds)`; set `UserId: scope.UserID()` on `newTimespan`.
- `GetTimespan`: `ErrNotFound` unless `matchesScope(t.timespans[idx].UserId, scope)`.
- `ListTimespans`: filter by `matchesScope` before the sort/pagination/count.
- `UpdateTimespan`: `t.tagsExist(ctx, scope, ...)`; locate by id + `matchesScope`; preserve `timespan.UserId = t.timespans[idx].UserId`.
- `DeleteTimespan`: locate by id + `matchesScope`.

- [ ] **Step 4: Run memory timespan contract subtests**

Run: `go test ./internal/repository/contract_test/ -run 'TestTimespanRepositoryContract/memory' -v`
Expected: PASS.

- [ ] **Step 5: Enforce scope in the postgres timespan store**

In `internal/repository/postgres/timespan.go`, apply the predicate pattern to `GetTimespan`, `ListTimespans` (count + data), `UpdateTimespan`, `DeleteTimespan`; add `user_id` to the `CreateTimespan` INSERT bound to `scope.UserID()`. Change `setTimespanTags` to take `scope` and guard with `r.tagsInScope(ctx, scope, tagIds)` → `model.ErrInvalidReference`, mirroring Task 4 Step 7. Update its two call sites.

- [ ] **Step 6: Run postgres timespan contract subtests**

Run: `go test ./internal/repository/contract_test/ -run 'TestTimespanRepositoryContract/postgres' -v`
Expected: PASS.

- [ ] **Step 7: Update postgres + memory timespan unit tests**

`internal/repository/postgres/timespan_test.go`: predicate/arg updates; `SELECT count(*) FROM tags` expectation where `setTimespanTags` runs; a foreign-tag-rejected case.
`internal/repository/memory/timespans_test.go`: add `TestMemoryStore_Timespan_ScopeIsolation`.

- [ ] **Step 8: Full verification + commit**

Run: `go test ./internal/repository/... ./internal/service/... && golangci-lint run ./internal/repository/...`
Expected: PASS, 0 issues.

```bash
git add -A
git commit -m "feat: scope timespan repository operations to the owner

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
Claude-Session: https://claude.ai/code/session_01UzTPnp1FQiGpFJAG1JdbXW"
```

---

## Task 6: Aggregate query scope enforcement

`GetTotalDurationByTags` (used by `GetProject`/`GetTag` "total time" includes) and `AggregateTimeSpentByTagsAndBuckets` (used by `GetProjectStats`).

**Files:**
- Modify: `internal/repository/postgres/timespan.go` (`GetTotalDurationByTags`, `AggregateTimeSpentByTagsAndBuckets`)
- Modify: `internal/repository/memory/timespans.go` (same two methods)
- Test: `internal/repository/contract_test/repository_timespan_test.go` and `repository_project_stats_test.go`
- Test: `internal/repository/postgres/timespan_test.go`
- Test: `internal/repository/memory/timespans_test.go`

**Interfaces:**
- Consumes: `matchesScope`; scoped timespan/tag stores.
- Produces: both aggregate methods restrict the `timespans` they sum to `scope` (in addition to the tag filter). An aggregate run under scope A never counts scope B's timespans.

- [ ] **Step 1: Write failing contract subtests**

In `repository_timespan_test.go` `run`:

```go
t.Run(repoName+"GetTotalDurationByTags_IsScoped", func(t *testing.T) {
	repo := newRepo(t)
	scopeA := model.UserScope(uuid.New())
	scopeB := model.UserScope(uuid.New())

	tagA := seedTags(t, ctx, repo, scopeA, 1)[0]
	tagB := seedTags(t, ctx, repo, scopeB, 1)[0]

	start := time.Now().UTC().Truncate(time.Second)
	_, err := repo.CreateTimespan(ctx, scopeA, model.Timespan{
		Name: "a", StartTime: start, EndTime: start.Add(time.Hour), TagIds: []uuid.UUID{tagA},
	})
	require.NoError(t, err)
	_, err = repo.CreateTimespan(ctx, scopeB, model.Timespan{
		Name: "b", StartTime: start, EndTime: start.Add(2 * time.Hour), TagIds: []uuid.UUID{tagB},
	})
	require.NoError(t, err)

	durA, err := repo.GetTotalDurationByTags(ctx, scopeA, []uuid.UUID{tagA})
	require.NoError(t, err)
	require.Equal(t, time.Hour, durA)

	// even asking for B's tag under scope A yields nothing
	durCross, err := repo.GetTotalDurationByTags(ctx, scopeA, []uuid.UUID{tagB})
	require.NoError(t, err)
	require.Equal(t, time.Duration(0), durCross)
})
```

In `repository_project_stats_test.go` `run`, add an analogous `AggregateTimeSpentByTagsAndBuckets_IsScoped` subtest: two owners each with a tagged timespan overlapping one bucket; asserting the scope-A aggregate reports only A's seconds.

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./internal/repository/contract_test/ -run '(GetTotalDurationByTags_IsScoped|AggregateTimeSpentByTagsAndBuckets_IsScoped)' -v`
Expected: FAIL (cross-scope timespans are counted).

- [ ] **Step 3: Enforce scope in the memory aggregates**

In `internal/repository/memory/timespans.go`:
- `GetTotalDurationByTags`: in the loop that collects `timespanIds`, skip any `timespan` where `!matchesScope(timespan.UserId, scope)`.
- `AggregateTimeSpentByTagsAndBuckets`: in the `for _, timespan := range t.timespans` loop, `continue` when `!matchesScope(timespan.UserId, scope)`.

- [ ] **Step 4: Enforce scope in the postgres aggregates**

In `internal/repository/postgres/timespan.go`:
- `GetTotalDurationByTags`: add `AND t.user_id IS NOT DISTINCT FROM $2` to the `WHERE`, bind `scope.UserID()` as `$2` (shift the tag-ids param to `$1` stays; add `$2`).
  ```sql
  SELECT SUM(t.end_time - t.start_time) AS total_time
  FROM timespans t
  WHERE t.deleted_at IS NULL
    AND t.user_id IS NOT DISTINCT FROM $2
    AND EXISTS (SELECT 1 FROM timespan_tags tt WHERE tt.timespan_id = t.id AND tt.tag_id = ANY($1))
  ```
  `r.db.QueryRow(ctx, q, tagIds, scope.UserID())`.
- `AggregateTimeSpentByTagsAndBuckets`: in the `matching_timespans` CTE add `AND t.user_id IS NOT DISTINCT FROM $4`, and pass `scope.UserID()` as the 4th query arg (`r.db.Query(ctx, q, tagIds, bucketStarts, bucketEnds, scope.UserID())`).

- [ ] **Step 5: Run the contract subtests**

Run: `go test ./internal/repository/contract_test/ -run '(GetTotalDurationByTags_IsScoped|AggregateTimeSpentByTagsAndBuckets_IsScoped)' -v`
Expected: PASS (memory + postgres).

- [ ] **Step 6: Update unit tests**

`internal/repository/postgres/timespan_test.go`: update the `GetTotalDurationByTags` and `AggregateTimeSpentByTagsAndBuckets` `ExpectQuery` regexes + `WithArgs` (extra `testScope.UserID()` arg).
`internal/repository/memory/timespans_test.go`: add a scoped-aggregate assertion if not already covered by contract tests.

- [ ] **Step 7: Full verification + commit**

Run: `go test ./internal/... && golangci-lint run ./internal/...`
Expected: PASS, 0 issues.

```bash
git add -A
git commit -m "feat: scope timespan aggregate queries to the owner

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
Claude-Session: https://claude.ai/code/session_01UzTPnp1FQiGpFJAG1JdbXW"
```

---

## Task 7: Coverage guard + final verification

**Files:**
- Modify: `internal/repository/contract_test/user_scoped_models_test.go`
- Test: full suite + e2e + lint.

**Interfaces:**
- Consumes: `userScopedModels` map (already exists), the `*ScopeIsolation` subtests added in Tasks 3–5.
- Produces: `TestUserScopedModelsHaveIsolationCoverage` — for every entry in `userScopedModels`, asserts a `<Type>RepositoryContract/*ScopeIsolation` subtest exists, so adding a user-scoped resource without scoping its queries fails the suite.

- [ ] **Step 1: Write the failing guard test**

In `internal/repository/contract_test/user_scoped_models_test.go` add:

```go
// Each user-scoped model must have a repository contract isolation subtest,
// named "<memory|postgres>ScopeIsolation", proving cross-owner access is denied.
func TestUserScopedModelsHaveIsolationCoverage(t *testing.T) {
	// map model type -> the Test function that must contain the subtest
	contractTests := map[string]func(*testing.T){
		"Project":  TestProjectRepositoryContract,
		"Tag":      TestTagRepositoryContract,
		"Timespan": TestTimespanRepositoryContract,
	}

	for modelName := range userScopedModels {
		fn, ok := contractTests[modelName]
		require.Truef(t, ok, "no contract test registered for user-scoped model %q", modelName)

		found := false
		captureT := &testing.T{}
		_ = captureT
		// Run the contract test and inspect subtest names via a marker test.
		t.Run(modelName, func(t *testing.T) {
			t.Run("hasIsolationSubtest", func(t *testing.T) {
				// See Step 3 for the actual mechanism.
				found = isolationSubtestExists(modelName)
				require.Truef(t, found, "%sRepositoryContract is missing a ScopeIsolation subtest", modelName)
			})
		})
		_ = fn
	}
}
```

> Note: Go gives no clean runtime API to enumerate a sibling test's subtests. Use the same AST approach as `modelsWithUserIDField`: scan this package's `*_test.go` files for `t.Run(repoName+"ScopeIsolation"` occurrences and map them to the `Test<Type>RepositoryContract` function they appear in.

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./internal/repository/contract_test/ -run 'TestUserScopedModelsHaveIsolationCoverage' -v`
Expected: FAIL — `isolationSubtestExists` undefined.

- [ ] **Step 3: Implement the AST-based check**

Replace the placeholder with a real implementation next to `modelsWithUserIDField`:

```go
// isolationSubtestExists reports whether Test<model>RepositoryContract in this
// package registers a subtest named "<repoName>ScopeIsolation".
func isolationSubtestExists(t *testing.T, modelName string) bool {
	t.Helper()

	entries, err := os.ReadDir(".")
	require.NoError(t, err)

	fnName := "Test" + modelName + "RepositoryContract"
	needle := `"ScopeIsolation"` // matched loosely: t.Run(repoName+"ScopeIsolation", ...)

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
```

Simplify `TestUserScopedModelsHaveIsolationCoverage` to:

```go
func TestUserScopedModelsHaveIsolationCoverage(t *testing.T) {
	for modelName := range userScopedModels {
		t.Run(modelName, func(t *testing.T) {
			require.Truef(t, isolationSubtestExists(t, modelName),
				"Test%sRepositoryContract needs a t.Run(repoName+\"ScopeIsolation\", …) subtest "+
					"proving cross-owner access is denied", modelName)
		})
	}
}
```

Add imports `go/ast`, `go/parser`, `go/token`, `os`, `strings` if not already present.

- [ ] **Step 4: Run to verify it passes**

Run: `go test ./internal/repository/contract_test/ -run 'TestUserScopedModelsHaveIsolationCoverage' -v`
Expected: PASS (Tag/Project/Timespan all have `ScopeIsolation` subtests from Tasks 3–5).

- [ ] **Step 5: Full project verification**

Run:
```bash
gofmt -l internal/
go build ./...
go vet ./...
go test ./...
golangci-lint run ./internal/...
```
Expected: `gofmt -l` prints nothing new (pre-existing entries `internal/api/handlers/tags_test.go`, `internal/api/handlers/user_test.go`, `internal/model/context_test.go`, `internal/repository/postgres/mock.go`, `internal/service/service_mock.go` are acceptable — do not fix them here); all builds/tests/lint pass.

- [ ] **Step 6: Manual review of the e2e path**

Confirm `test/e2e/main_test.go` still mounts no auth middleware, so e2e runs in the unowned scope and its (unchanged) tests pass — this is the userless-mode end-to-end check. No code change; just verify `go test ./test/e2e/...` is green.

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "test: require a scope-isolation contract subtest per user-scoped model

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
Claude-Session: https://claude.ai/code/session_01UzTPnp1FQiGpFJAG1JdbXW"
```

- [ ] **Step 8: Update the branch memory note**

Append to `/Users/olars/.claude/projects/-Users-olars-dev-inundated/memory/pr-85-user-support.md`: the "Data model migration" checklist item is done on branch `data-model-user-scoping` — repository layer is user-scoped via `model.OwnerScope`, resolved by `service.ownerScope(ctx)`; userless mode maps to `UnownedScope()` (`user_id IS NULL`).

---

## Self-Review

**1. Spec coverage**

| Spec section | Task |
|---|---|
| `model.OwnerScope` (§1) | Task 1 |
| `ownerScope(ctx)` resolver (§2) | Task 2 Step 7 |
| Repository interface `scope` param (§3) | Task 2 Steps 1–3 |
| Mocks updated by hand (§3) | Task 2 Steps 2–3 |
| Postgres `IS NOT DISTINCT FROM` predicate, reads/writes (§4) | Tasks 3, 4, 5 |
| Postgres tag-ownership guard on attach (§4) | Task 4 Steps 6–7, Task 5 Step 5 |
| Postgres aggregates scoped (§4) | Task 6 Step 4 |
| Join-table read helpers unchanged (§4) | Respected — Tasks 3–5 do not touch `projectTagIds`/`timespanTagIds` |
| Memory `matchesScope` + per-method filtering (§5) | Task 3 Step 3, Tasks 3–6 |
| Memory `tagsExistInScope` (§5) | Task 4 Step 3 (implemented by threading scope into `tagsExist`) |
| First-user adoption unchanged (§6) | No task touches `CreateUserAdoptingOrphans`; Task 2 Step 11 keeps `seedOrphanResources` unowned |
| Contract isolation subtests (§Testing) | Tasks 3–6 |
| `user_scoped_models_test.go` guard (§Testing) | Task 7 |
| Postgres/memory unit test updates (§Testing) | Tasks 2–6 |
| Service tests + ctx→scope assertions (§Testing) | Task 2 Steps 9–10 |
| e2e unchanged, userless path (§Testing) | Task 7 Step 6 |
| No schema change / no API change / 404 (§Constraints) | No migration task; no `openapi`/`api` files in any task; cross-scope ⇒ `ErrNotFound` throughout |

No gaps.

**2. Placeholder scan** — Task 7 Step 1 deliberately shows a placeholder (`isolationSubtestExists` undefined) as the *failing test*, resolved in Step 3 with full code. All other steps contain complete code or exact transformation rules with worked examples. No "TBD"/"add error handling"/"similar to Task N".

**3. Type consistency** — `model.OwnerScope`, `UserScope`, `UnownedScope`, `OwnerScope.UserID() *uuid.UUID` used identically across Tasks 1–7. `ownerScope(ctx)` (service, unexported) consistent. `matchesScope(userID *uuid.UUID, scope model.OwnerScope) bool` (memory) consistent Tasks 3–6. `(*PostgresStore).tagsInScope(ctx, scope, tagIds) (bool, error)` defined Task 4, reused Task 5. `seedTags(t, ctx, repo, scope, n)` signature consistent Tasks 2–6. `testScope` package var consistent. Predicate text `user_id IS NOT DISTINCT FROM $N` identical everywhere.
