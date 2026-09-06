# User-Scoping Follow-ups Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix the three documented limitations of the user-scoping work — index defeat on scoped list/aggregate queries, the Postgres store not returning `user_id`, and non-transactional project/timespan creates.

**Architecture:** All three are Postgres-store-only. #2 adds `user_id` to read/return column lists. #1 adds an `ownerPredicate` helper that emits `= $n` / `IS NULL` (index-usable) instead of `IS NOT DISTINCT FROM $n` at the 5 multi-row query sites. #3 gives `Querier` a `Begin` method and a `withTx` helper, and wraps the 4 project/timespan create/update methods so tag validation runs first inside a transaction that rolls back on any failure. No schema change, no API change; the memory store is untouched except for one new assertion it already satisfies.

**Tech Stack:** Go, `github.com/google/uuid`, `github.com/jackc/pgx/v5` (+ `pgxpool`, `pgx.Tx`), `github.com/pashagolub/pgxmock/v4` (`ExpectBegin`/`ExpectCommit`/`ExpectRollback`), `github.com/stretchr/testify`, testcontainers-go (postgres).

**Spec:** `docs/superpowers/specs/2026-09-06-user-scoping-followups-design.md`

## Global Constraints

- **No new migration.** The existing `0006` `idx_{tags,projects,timespans}_user_id` btree indexes serve both `= $n` and `IS NULL`.
- **No API / OpenAPI / handler change.** Cross-user access still returns `model.ErrNotFound`; no owner field is added to any response.
- **Memory store: do not change enforcement logic.** #2 adds a contract assertion the memory store already passes; #1 and #3 are Postgres-only (memory list-filtering is already correct, memory create/update is already atomic under its mutex).
- **`GetX` / `UpdateX` / `DeleteX` keep `user_id IS NOT DISTINCT FROM $N`** — they are driven by `WHERE id = $1` (primary key); the owner clause is a one-row filter and no index is involved. Only `ListX` (count + data) and the two aggregates change to `ownerPredicate`.
- Existing repository patterns: `model.ErrNotFound` / `ErrInvalidReference` / `ErrInvalidArgument`, pgxmock unit tests, memory+postgres contract tests. The postgres contract-test `newRepo` factories already call `seedScopeUser(t, ctx, repo, testScope)`; ad-hoc `model.UserScope(uuid.New())` scopes in a subtest need their own `seedScopeUser` call.
- `testScope` (postgres unit + contract tests) = `model.UserScope(uuid.MustParse("11111111-1111-1111-1111-111111111111"))`, defined in `internal/repository/postgres/helpers_test.go` and `internal/repository/contract_test/repository_test.go`.
- Every commit message ends with:
  ```
  Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
  Claude-Session: https://claude.ai/code/session_01UzTPnp1FQiGpFJAG1JdbXW
  ```
- Verification: `go build ./...`, `go vet ./...`, `go test ./...`, `golangci-lint run ./internal/...`. `gofmt -l internal/` must list only the known pre-existing entries (`internal/api/handlers/tags_test.go`, `internal/api/handlers/user_test.go`, `internal/model/context_test.go`, `internal/repository/postgres/mock.go`, `internal/service/service_mock.go`). Container tests need Docker; a testcontainers "context deadline / wait for ready" flake → re-run once.

---

## File Structure

**Modified — production:**
- `internal/repository/postgres/helpers.go` — new `ownerPredicate` helper (Task 2).
- `internal/repository/postgres/postgres.go` — `Querier` gains `Begin`; new `withTx` helper (Task 3).
- `internal/repository/postgres/tag.go` — `user_id` in `GetTag`/`ListTags` SELECT + `CreateTag`/`UpdateTag` RETURNING (Task 1); `ownerPredicate` in `ListTags` (Task 2).
- `internal/repository/postgres/project.go` — same for projects (Tasks 1, 2); `tagsInScope`/`setProjectTags` signature change + `CreateProject`/`UpdateProject` wrapped in `withTx` (Task 3).
- `internal/repository/postgres/timespan.go` — same for timespans (Tasks 1, 2); `ownerPredicate` in the two aggregates (Task 2); `setTimespanTags` signature change + `CreateTimespan`/`UpdateTimespan` wrapped in `withTx` (Task 3).

**Modified — tests:**
- `internal/repository/postgres/{tag,project,timespan}_test.go` — column lists / row builders gain `user_id`; List/aggregate query regexes + `WithArgs` updated for `ownerPredicate`; `expectSetProjectTags`/`expectSetTimespanTags` and the create/update tests gain `ExpectBegin`/`ExpectCommit`; new rollback tests.
- `internal/repository/postgres/helpers_test.go` — (Task 1) `aTag()`/`aProject()`/`aTimespan()` fixtures may gain a `UserId`; introduce `tagCols` if helpful.
- `internal/repository/contract_test/repository_{tag,project,timespan}_test.go` — `CreateAndGet` / `ScopeIsolation` gain `UserId` assertions (Task 1); `CannotAttachAnotherUsersTag` gains "list is empty after the failed op" assertions (Task 3).
- `internal/repository/contract_test/repository_project_stats_test.go` — (Task 2, optional) an EXPLAIN-plan test.

---

## Task 1: Postgres returns `user_id` (#2)

**Files:**
- Modify: `internal/repository/postgres/tag.go`, `project.go`, `timespan.go`
- Test: `internal/repository/postgres/{tag,project,timespan}_test.go`, `helpers_test.go`
- Test: `internal/repository/contract_test/repository_{tag,project,timespan}_test.go`

**Interfaces:**
- Consumes: `model.{Tag,Project,Timespan}.UserId *uuid.UUID` (already exists).
- Produces: after this task, a `model.{Tag,Project,Timespan}` returned by any Postgres `Get*` / `List*` / `Create*` / `Update*` has `UserId` populated (nil for an unowned row), matching the memory store. The contract tests now assert this against both stores.

- [ ] **Step 1: Write the failing contract assertions**

In `internal/repository/contract_test/repository_tag_test.go`, replace the `CreateAndGet` subtest body with:

```go
t.Run(repoName+"CreateAndGet", func(t *testing.T) {
	repo := newRepo(t)

	created, err := repo.CreateTag(ctx, testScope, model.Tag{Name: "work", Color: "#ff0000"})
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, created.Id)
	require.NotNil(t, created.UserId)
	require.Equal(t, *testScope.UserID(), *created.UserId)

	got, err := repo.GetTag(ctx, testScope, created.Id)
	require.NoError(t, err)
	require.Equal(t, "work", got.Name)
	require.NotNil(t, got.UserId)
	require.Equal(t, *testScope.UserID(), *got.UserId)
})

t.Run(repoName+"UnownedCreateHasNilUserId", func(t *testing.T) {
	repo := newRepo(t)

	created, err := repo.CreateTag(ctx, model.UnownedScope(), model.Tag{Name: "u", Color: "#ffffff"})
	require.NoError(t, err)
	require.Nil(t, created.UserId)

	got, err := repo.GetTag(ctx, model.UnownedScope(), created.Id)
	require.NoError(t, err)
	require.Nil(t, got.UserId)
})
```

Add the analogous pair to `repository_project_test.go` (`model.Project{Name: "p", Color: "#111111"}`) and `repository_timespan_test.go` (`model.Timespan{Name: "t", StartTime: time.Now().UTC(), EndTime: time.Now().UTC().Add(time.Hour)}`), each asserting `created.UserId` / `got.UserId` for a `testScope` create and nil for an `UnownedScope()` create.

- [ ] **Step 2: Run to verify they fail (postgres only)**

Run: `go test ./internal/repository/contract_test/ -run 'TestTagRepositoryContract/postgres(CreateAndGet|UnownedCreateHasNilUserId)' -v`
Expected: FAIL — `created.UserId` / `got.UserId` is `nil` from the Postgres store (the memory subtests pass; Postgres `SELECT id, name, color` never returns `user_id`).

- [ ] **Step 3: Add `user_id` to the Postgres tag queries**

In `internal/repository/postgres/tag.go`:
- `GetTag`: `SELECT id, name, color, user_id FROM tags WHERE ...` and `.Scan(&t.Id, &t.Name, &t.Color, &t.UserId)`.
- `ListTags` data query only (not the count query): `SELECT id, name, color, user_id FROM tags WHERE ...` and in the row loop `rows.Scan(&t.Id, &t.Name, &t.Color, &t.UserId)`.
- `CreateTag`: `RETURNING id, name, color, user_id` and `.Scan(&created.Id, &created.Name, &created.Color, &created.UserId)`.
- `UpdateTag`: `RETURNING id, name, color, user_id` and `.Scan(&updated.Id, &updated.Name, &updated.Color, &updated.UserId)` (`user_id` stays out of the `SET`).

- [ ] **Step 4: Do the same for `project.go` and `timespan.go`**

- `project.go`: `GetProject` / `ListProjects` data query → `SELECT id, name, color, time_budget, user_id`; `CreateProject` / `UpdateProject` → `RETURNING id, name, color, time_budget, user_id`. Scan `&p.UserId` / `&created.UserId` / `&updated.UserId` as the last target.
- `timespan.go`: `GetTimespan` / `ListTimespans` data query → `SELECT id, name, start_time, end_time, user_id`; `CreateTimespan` / `UpdateTimespan` → `RETURNING id, name, start_time, end_time, user_id`. Scan `&ts.UserId` / `&created.UserId` / `&updated.UserId` as the last target.
- Do **not** touch `GetTotalDurationByTags` / `AggregateTimeSpentByTagsAndBuckets` (no entity rows), the `ListX` **count** queries, or the `*TagIds` helpers.

- [ ] **Step 5: Update the postgres unit-test row builders**

In `internal/repository/postgres/project_test.go`: `var projectCols = []string{"id", "name", "color", "time_budget", "user_id"}`. Every `pgxmock.NewRows(projectCols).AddRow(p.Id, p.Name, p.Color, p.TimeBudget)` becomes `.AddRow(p.Id, p.Name, p.Color, p.TimeBudget, p.UserId)` (or `testScope.UserID()` where the fixture has no `UserId`). The `Get`/`List`-data/`Create`/`Update` query regexes that assert the column list must include `user_id` (e.g. `SELECT id, name, color, time_budget, user_id FROM projects` / `RETURNING id, name, color, time_budget, user_id`). The `ListProjects` **count**-query regex is unchanged.

In `internal/repository/postgres/timespan_test.go`: `var timespanCols = []string{"id", "name", "start_time", "end_time", "user_id"}` and the same `AddRow` / regex updates.

In `internal/repository/postgres/tag_test.go`: introduce `var tagCols = []string{"id", "name", "color", "user_id"}` at the top and convert the inline `pgxmock.NewRows([]string{"id", "name", "color"})` uses for `GetTag` / `ListTags` data / `CreateTag` / `UpdateTag` to `pgxmock.NewRows(tagCols)` with the extra `AddRow` value; update those query regexes to include `user_id`. Leave `ListTags` count-query rows (`[]string{"count"}`) alone.

In `internal/repository/postgres/helpers_test.go`: give `aProject()` / `aTag()` / `aTimespan()` a `UserId: ptr(testScope.UserID())`-style field only if a test needs a concrete non-nil value; otherwise pass `testScope.UserID()` explicitly in the `AddRow`. Keep it minimal — prefer `AddRow(..., testScope.UserID())`.

- [ ] **Step 6: Run the postgres unit tests**

Run: `go test ./internal/repository/postgres/ -v`
Expected: PASS. If a `Scan` count mismatch panics, a row builder is missing its 5th/4th value.

- [ ] **Step 7: Run the contract tests**

Run: `go test ./internal/repository/contract_test/ -run 'TestTagRepositoryContract|TestProjectRepositoryContract|TestTimespanRepositoryContract' -v`
Expected: PASS (memory + postgres; the new `CreateAndGet` / `UnownedCreateHasNilUserId` assertions now pass on both).

- [ ] **Step 8: Full verification + commit**

Run: `go build ./... && go vet ./... && go test ./... && golangci-lint run ./internal/...`
Expected: all green, 0 lint issues.

```bash
git add -A
git commit -m "feat: return user_id from the postgres repository reads

The postgres store now selects user_id back on Get/List/Create/Update so
model.{Tag,Project,Timespan}.UserId is populated to match the memory
store; contract tests assert it against both.

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
Claude-Session: https://claude.ai/code/session_01UzTPnp1FQiGpFJAG1JdbXW"
```

---

## Task 2: Index-friendly owner predicate (#1)

**Files:**
- Modify: `internal/repository/postgres/helpers.go`, `tag.go`, `project.go`, `timespan.go`
- Test: `internal/repository/postgres/{tag,project,timespan}_test.go`
- Test (optional): `internal/repository/contract_test/repository_project_stats_test.go`

**Interfaces:**
- Consumes: `model.OwnerScope.UserID() *uuid.UUID`.
- Produces:
  - `func ownerPredicate(column string, scope model.OwnerScope, n int) (sql string, args []any)` in package `postgres` — returns `("<column> = $n", []any{uuid})` when scoped, `("<column> IS NULL", nil)` when unowned.
  - The 3 `ListX` count queries, 3 `ListX` data queries, and 2 aggregate queries build their `user_id` predicate via `ownerPredicate` with the owner arg placed **last** in the arg list. No behavior change — same rows, same errors.

- [ ] **Step 1: Add the helper**

In `internal/repository/postgres/helpers.go`, add (and add `"fmt"` + the model import):

```go
// ownerPredicate returns a WHERE fragment on the given user_id column and the
// value to bind for it (nil ⇒ the fragment takes no parameter). It uses
// `<column> = $n` / `<column> IS NULL` rather than
// `<column> IS NOT DISTINCT FROM $n` so the btree idx_<table>_user_id index
// from migration 0006 applies. Callers place the fragment last in the WHERE
// clause so $n is the final positional parameter and LIMIT/OFFSET numbering is
// stable across the scoped and unowned variants.
func ownerPredicate(column string, scope model.OwnerScope, n int) (sql string, args []any) {
	if id := scope.UserID(); id != nil {
		return fmt.Sprintf("%s = $%d", column, n), []any{*id}
	}
	return column + " IS NULL", nil
}
```

- [ ] **Step 2: Write the failing unit-test expectations for `ListTags`**

In `internal/repository/postgres/tag_test.go`, for the existing `ListTags` success test (scoped, `testScope`), change the two `mock.ExpectQuery(...)` regexes to expect `user_id = \$1` (count) and `... user_id = \$3 ORDER BY name LIMIT \$1 OFFSET \$2` (data), and `WithArgs` to `(testScope.UserID())` for count and `(25, 0, testScope.UserID())` for data (limit/offset first, owner last). Add a new unowned test: `repo.ListTags(ctx, model.UnownedScope(), ...)` expecting `user_id IS NULL` in both regexes and `WithArgs()` (count, no args) / `WithArgs(25, 0)` (data).

Run: `go test ./internal/repository/postgres/ -run 'ListTag' -v`
Expected: FAIL — the implementation still emits `user_id IS NOT DISTINCT FROM $1`.

- [ ] **Step 3: Rebuild the `ListTags` queries with `ownerPredicate`**

In `internal/repository/postgres/tag.go` `ListTags`:

```go
ownerSQL, ownerArgs := ownerPredicate("user_id", scope, 1)
countQ := `SELECT COUNT(*) FROM tags WHERE deleted_at IS NULL AND ` + ownerSQL
var totalCount int
if err := r.db.QueryRow(ctx, countQ, ownerArgs...).Scan(&totalCount); err != nil {
	return model.Page[model.Tag]{}, fmt.Errorf("ListTags count: %w", err)
}

dataOwnerSQL, _ := ownerPredicate("user_id", scope, 3)
dataQ := `SELECT id, name, color, user_id FROM tags
	WHERE deleted_at IS NULL AND ` + dataOwnerSQL + `
	ORDER BY name
	LIMIT $1 OFFSET $2`
args := append([]any{params.Limit, params.Offset}, ownerArgs...)
rows, err := r.db.Query(ctx, dataQ, args...)
```

(`ownerArgs` is the same slice for both — compute once, reuse. `dataOwnerSQL` re-derives just the SQL fragment with `$3`.) Keep the rest of the method (row loop with `user_id` from Task 1, `tagErr` loop, nil-slice guard) unchanged.

Run: `go test ./internal/repository/postgres/ -run 'ListTag' -v` → PASS.

- [ ] **Step 4: Do `ListProjects` and `ListTimespans`**

`project.go` `ListProjects` and `timespan.go` `ListTimespans` — identical transformation. `ListTimespans` keeps `ORDER BY start_time DESC`. Column lists already carry `user_id` from Task 1. Update the corresponding `project_test.go` / `timespan_test.go` List expectations (scoped regex `user_id = \$1` / `\$3`; add an unowned variant with `user_id IS NULL` and no owner arg).

- [ ] **Step 5: Do the two aggregate queries**

In `internal/repository/postgres/timespan.go`:

`GetTotalDurationByTags` — `tagIds` stays `$1`, owner predicate at `$2`:
```go
ownerSQL, ownerArgs := ownerPredicate("t.user_id", scope, 2)
q := `SELECT SUM(t.end_time - t.start_time) AS total_time
	FROM timespans t
	WHERE t.deleted_at IS NULL
		AND ` + ownerSQL + `
		AND EXISTS (
			SELECT 1 FROM timespan_tags tt
			WHERE tt.timespan_id = t.id AND tt.tag_id = ANY($1)
		)`
args := append([]any{tagIds}, ownerArgs...)
err := r.db.QueryRow(ctx, q, args...).Scan(&duration)
```
Keep the `len(tagIds) == 0 → return 0, nil` short-circuit above it, and the `pgx.ErrNoRows || duration == nil` handling below.

`AggregateTimeSpentByTagsAndBuckets` — args are `tagIds($1)`, `bucketStarts($2)`, `bucketEnds($3)`, owner at `$4`:
```go
ownerSQL, ownerArgs := ownerPredicate("t.user_id", scope, 4)
q := `WITH input_buckets AS ( ... $2 ... $3 ... ),
	bucket_window AS ( ... ),
	matching_timespans AS (
		SELECT t.id, t.start_time, t.end_time
		FROM timespans t
		CROSS JOIN bucket_window bw
		WHERE t.deleted_at IS NULL
			AND ` + ownerSQL + `
			AND EXISTS (
				SELECT 1 FROM timespan_tags tt
				WHERE tt.timespan_id = t.id AND tt.tag_id = ANY($1)
			) AND t.start_time < bw.max_end AND t.end_time > bw.min_start
	)
	SELECT ... (unchanged) ...`
args := append([]any{tagIds, bucketStarts, bucketEnds}, ownerArgs...)
rows, err := r.db.Query(ctx, q, args...)
```
Everything else in the CTE (the `input_buckets`/`bucket_window`/final `SELECT` with the overlap math) is byte-identical.

Update the `timespan_test.go` expectations for both aggregates: scoped regex `t.user_id = \$2` / `t.user_id = \$4`, `WithArgs(tagIds, testScope.UserID())` / `WithArgs(tagIds, starts, ends, testScope.UserID())`; add an unowned variant with `t.user_id IS NULL` and the owner arg dropped.

- [ ] **Step 6: Contract tests unchanged — run them**

Run: `go test ./internal/repository/contract_test/ -run 'TestTagRepositoryContract|TestProjectRepositoryContract|TestTimespanRepositoryContract|TestProjectStatsRepositoryContract' -v`
Expected: PASS — behavior is identical, the isolation and aggregate subtests from the parent branch still hold.

- [ ] **Step 7 (optional): EXPLAIN-plan regression test**

In `internal/repository/contract_test/repository_project_stats_test.go` (or a new `repository_index_pg_test.go`), add a postgres-only test:

```go
func TestPostgres_ScopedListUsesIndex(t *testing.T) {
	ctx := context.Background()
	pool := testutils.StartPostgresContainerWithMigrationsApplied(ctx, t)
	repo := postgres.NewPostgresStoreFromPool(pool)

	a := model.UserScope(uuid.New())
	b := model.UserScope(uuid.New())
	seedScopeUser(t, ctx, repo, a)
	seedScopeUser(t, ctx, repo, b)
	for i := 0; i < 300; i++ {
		_, _ = repo.CreateTag(ctx, a, model.Tag{Name: fmt.Sprintf("a%d", i), Color: "#111111"})
		_, _ = repo.CreateTag(ctx, b, model.Tag{Name: fmt.Sprintf("b%d", i), Color: "#222222"})
	}
	_, _ = pool.Exec(ctx, "ANALYZE tags")

	var plan string
	err := pool.QueryRow(ctx,
		`EXPLAIN (FORMAT TEXT) SELECT id, name, color, user_id FROM tags
		 WHERE deleted_at IS NULL AND user_id = $1 ORDER BY name LIMIT 25 OFFSET 0`,
		*a.UserID()).Scan(&plan)
	require.NoError(t, err)
	require.NotContains(t, plan, "Seq Scan on tags", "scoped list should use idx_tags_user_id, plan was:\n"+plan)
}
```

Run it. If the planner still picks a seq scan at 600 rows (small-table threshold) or it proves flaky in CI, **delete this test** — the pgxmock SQL-shape assertions in Steps 2–5 are the real regression guard. Note the outcome in the report.

- [ ] **Step 8: Full verification + commit**

Run: `go build ./... && go vet ./... && go test ./... && golangci-lint run ./internal/...`

```bash
git add -A
git commit -m "perf: use index-friendly user_id predicate on scoped list queries

Scoped ListX and the two timespan aggregates now emit `user_id = $n` /
`user_id IS NULL` (via ownerPredicate) instead of
`user_id IS NOT DISTINCT FROM $n`, so the 0006 btree indexes apply.
Get/Update/Delete keep IS NOT DISTINCT FROM (PK-driven, no index needed).

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
Claude-Session: https://claude.ai/code/session_01UzTPnp1FQiGpFJAG1JdbXW"
```

---

## Task 3: Transactional create/update (#3)

**Files:**
- Modify: `internal/repository/postgres/postgres.go`, `project.go`, `timespan.go`
- Test: `internal/repository/postgres/{project,timespan}_test.go`
- Test: `internal/repository/contract_test/repository_{project,timespan}_test.go`

**Interfaces:**
- Consumes: `Querier` (extended here), `model.OwnerScope`.
- Produces:
  - `Querier` interface gains `Begin(ctx context.Context) (pgx.Tx, error)`.
  - `func (r *PostgresStore) withTx(ctx context.Context, fn func(q Querier) error) error`.
  - `tagsInScope(ctx context.Context, q Querier, scope model.OwnerScope, tagIds []uuid.UUID) (bool, error)` — was `(ctx, scope, tagIds)`, now takes `q`.
  - `setProjectTags(ctx context.Context, q Querier, projectId uuid.UUID, tagIds []uuid.UUID) error` — was `(ctx, scope, projectId, tagIds)`; **`scope` removed**, no longer calls `tagsInScope`, pure DELETE-then-INSERT against `q`.
  - `setTimespanTags(ctx context.Context, q Querier, timespanId uuid.UUID, tagIds []uuid.UUID) error` — same shape.
  - `CreateProject` / `UpdateProject` / `CreateTimespan` / `UpdateTimespan` run their tag check + parent write + join-row replace inside one `withTx`; a foreign tag → `ErrInvalidReference` and a full rollback (no orphan row).

- [ ] **Step 1: Write the failing contract assertion**

In `internal/repository/contract_test/repository_project_test.go`, extend the `CannotAttachAnotherUsersTag` subtest.

(a) Immediately after the existing create-path `require.ErrorIs(err, model.ErrInvalidReference)` (and before the `projA, err := repo.CreateProject(...)` line), add:

```go
// The rejected create leaves no orphan project row.
page, err := repo.ListProjects(ctx, scopeA, model.DefaultPaginationParams())
require.NoError(t, err)
require.Empty(t, page.Data)
```

(b) The existing update-path block creates `projA` (name `"a"`) then sets `projA.TagIds = []uuid.UUID{tagB}`. Also set `projA.Name = "renamed"` on that line, and after the update-path `require.ErrorIs(err, model.ErrInvalidReference)` add:

```go
// The rejected update rolled back the name change too.
got, err := repo.GetProject(ctx, scopeA, projA.Id)
require.NoError(t, err)
require.Equal(t, "a", got.Name)
require.Empty(t, got.TagIds)
```

Add the analogous assertions to `repository_timespan_test.go`'s `CannotAttachAnotherUsersTag`: `ListTimespans` empty after the failed create; and for the update path, rename the timespan in the failing update call and assert `GetTimespan` still returns the original name with empty `TagIds`.

- [ ] **Step 2: Run to verify they fail (postgres only)**

Run: `go test ./internal/repository/contract_test/ -run 'TestProjectRepositoryContract/postgresCannotAttachAnotherUsersTag' -v`
Expected: FAIL — after the rejected `CreateProject`, `ListProjects` returns 1 orphaned row (the parent INSERT committed before `setProjectTags` rejected the tag). Memory already passes.

- [ ] **Step 3: Extend `Querier` and add `withTx`**

In `internal/repository/postgres/postgres.go`:

```go
type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

// withTx runs fn inside a transaction, committing on success and rolling back
// on any error.
func (r *PostgresStore) withTx(ctx context.Context, fn func(q Querier) error) (err error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()
	if err = fn(tx); err != nil {
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
```

`var _ Querier = (*pgxpool.Pool)(nil)` still holds (`pgxpool.Pool` has `Begin`). `pgx.Tx` satisfies `Querier` (it has `QueryRow`/`Query`/`Exec`/`Begin`), so `fn(tx)` type-checks. `pgxmock.PgxPoolIface` also has `Begin(ctx) (pgx.Tx, error)`, so `newMock` still compiles.

Run: `go build ./...`
Expected: PASS. Do **not** run `golangci-lint` yet — `withTx` is an unused unexported method until Step 5 wires it in and staticcheck U1000 will flag it. Lint runs clean at Step 9 when the task is complete.

- [ ] **Step 4: Change the three helper signatures**

In `internal/repository/postgres/project.go`:

```go
func (r *PostgresStore) tagsInScope(ctx context.Context, q Querier, scope model.OwnerScope, tagIds []uuid.UUID) (bool, error) {
	if len(tagIds) == 0 {
		return true, nil
	}
	const query = `
		SELECT count(*) FROM tags
		WHERE id = ANY($1) AND deleted_at IS NULL AND user_id IS NOT DISTINCT FROM $2`
	var n int
	if err := q.QueryRow(ctx, query, tagIds, scope.UserID()).Scan(&n); err != nil {
		return false, fmt.Errorf("tagsInScope: %w", err)
	}
	return n == len(tagIds), nil
}

func (r *PostgresStore) setProjectTags(ctx context.Context, q Querier, projectId uuid.UUID, tagIds []uuid.UUID) error {
	if _, err := q.Exec(ctx, `DELETE FROM project_tags WHERE project_id = $1`, projectId); err != nil {
		return fmt.Errorf("setProjectTags delete: %w", err)
	}
	for _, tagId := range tagIds {
		if _, err := q.Exec(ctx,
			`INSERT INTO project_tags (project_id, tag_id) VALUES ($1, $2)`,
			projectId, tagId,
		); err != nil {
			return fmt.Errorf("setProjectTags insert: %w", model.ErrInvalidReference)
		}
	}
	return nil
}
```

(`tagsInScope` keeps `IS NOT DISTINCT FROM` — it is an `id = ANY($1)` lookup, the owner clause is a filter on an already-narrowed set, no index concern. Do not route it through `ownerPredicate`.)

In `internal/repository/postgres/timespan.go`, `setTimespanTags` gets the same treatment (drop `scope` + the `tagsInScope` call, take `q Querier`, use `q.Exec`).

- [ ] **Step 5: Wrap `CreateProject` / `UpdateProject`**

`CreateProject`:

```go
func (r *PostgresStore) CreateProject(ctx context.Context, scope model.OwnerScope, project model.Project) (model.Project, error) {
	if project.Name == "" {
		return model.Project{}, fmt.Errorf("CreateProject: name must not be empty: %w", model.ErrInvalidArgument)
	}
	if project.Id == uuid.Nil {
		project.Id = uuid.New()
	}

	var created model.Project
	err := r.withTx(ctx, func(q Querier) error {
		ok, err := r.tagsInScope(ctx, q, scope, project.TagIds)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("CreateProject: %w", model.ErrInvalidReference)
		}

		const insert = `
			INSERT INTO projects (id, name, color, time_budget, user_id)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, name, color, time_budget, user_id`
		if err := q.QueryRow(ctx, insert, project.Id, project.Name, project.Color, project.TimeBudget, scope.UserID()).
			Scan(&created.Id, &created.Name, &created.Color, &created.TimeBudget, &created.UserId); err != nil {
			return fmt.Errorf("CreateProject: %w", err)
		}
		return r.setProjectTags(ctx, q, created.Id, project.TagIds)
	})
	if err != nil {
		return model.Project{}, err
	}
	created.TagIds = project.TagIds
	return created, nil
}
```

`UpdateProject`: same pattern — validation outside, then `withTx(func(q) { tagsInScope → UPDATE ... RETURNING (with pgx.ErrNoRows → ErrNotFound) → setProjectTags })`. Keep the `errors.Is(err, pgx.ErrNoRows)` branch returning `fmt.Errorf("UpdateProject %s: %w", project.Id, model.ErrNotFound)` — return that from inside the closure; `withTx` rolls back and propagates it unchanged.

- [ ] **Step 6: Wrap `CreateTimespan` / `UpdateTimespan`**

Identical structure in `timespan.go` — validation (`StartTime`/`EndTime`) outside the closure, then `withTx(func(q) { tagsInScope → INSERT/UPDATE ... RETURNING id, name, start_time, end_time, user_id → setTimespanTags })`.

- [ ] **Step 7: Update the postgres unit tests**

In `internal/repository/postgres/project_test.go`:
- `expectSetProjectTags(mock, projectId, tagIds)` — **remove** the `SELECT count(*) FROM tags` expectation (it moved to the method). The helper now registers only `DELETE FROM project_tags` + the per-tag `INSERT`s.
- Every `Create*` / `Update*` test: wrap the expectation block in `mock.ExpectBegin()` … `mock.ExpectCommit()`. Order inside: (a) if the fixture has tag ids, `mock.ExpectQuery(\`SELECT count\(\*\) FROM tags\`).WithArgs(p.TagIds, testScope.UserID()).WillReturnRows(...AddRow(len(p.TagIds)))`; (b) `ExpectQuery(\`INSERT INTO projects\`)` / `ExpectQuery(\`UPDATE projects\`)` returning `projectCols` rows; (c) `expectSetProjectTags(...)`.
- The existing "foreign tag rejected" test (`WillReturnRows(...AddRow(len(p.TagIds) - 1))`): now `mock.ExpectBegin()`, the short-count `SELECT count(*)`, `mock.ExpectRollback()`, assert `ErrInvalidReference`. **Do not** register the `INSERT` expectation — its absence + `ExpectationsWereMet` proves the parent write never ran.
- The `UpdateProject` not-found test: `ExpectBegin`, (count query if tags), `ExpectQuery(\`UPDATE projects\`)...WillReturnRows(empty)`, `ExpectRollback`, assert `ErrNotFound`.

Same changes in `internal/repository/postgres/timespan_test.go` for `expectSetTimespanTags` and the timespan create/update tests.

- [ ] **Step 8: Run postgres unit + contract tests**

Run: `go test ./internal/repository/postgres/ ./internal/repository/contract_test/ -v`
Expected: PASS — the `CannotAttachAnotherUsersTag` contract subtests now show an empty list after the rejected op (real rollback), and the unit rollback tests pass.

- [ ] **Step 9: Full verification + commit**

Run: `go build ./... && go vet ./... && go test ./... && golangci-lint run ./internal/...`
Expected: all green, 0 lint issues. `gofmt -l internal/` lists only the known pre-existing entries.

```bash
git add -A
git commit -m "fix: wrap project/timespan create and update in a transaction

Tag-ownership validation now runs as the first statement inside a
transaction that also does the parent INSERT/UPDATE and the join-row
replace; a foreign tag rolls the whole thing back instead of leaving an
orphaned project/timespan row.

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
Claude-Session: https://claude.ai/code/session_01UzTPnp1FQiGpFJAG1JdbXW"
```

- [ ] **Step 10: Update the branch memory note**

Append to `/Users/olars/.claude/projects/-Users-olars-dev-inundated/memory/pr-85-user-support.md`: the three PR #103 follow-ups are done on the same branch — Postgres reads return `user_id`; scoped `ListX`/aggregates use `user_id = $n` / `IS NULL` (index-usable, `Get/Update/Delete` keep `IS NOT DISTINCT FROM`); project/timespan create+update run in a `withTx` with tag validation first so a foreign tag rolls back. (Not committed — outside the repo.)

---

## Self-Review

**1. Spec coverage**

| Spec section | Task |
|---|---|
| #1 `ownerPredicate` helper (§Design 1) | Task 2 Step 1 |
| #1 applied to 3 ListX (count+data) + 2 aggregates, owner arg last (§Design 1) | Task 2 Steps 3–5 |
| #1 `Get/Update/Delete` keep `IS NOT DISTINCT FROM` (§Constraints) | Not touched in any task; stated in Global Constraints |
| #1 pgxmock scoped + unowned variants (§Design 1 Tests) | Task 2 Steps 2, 4, 5 |
| #1 optional EXPLAIN test, drop-if-flaky (§Design 1 Tests) | Task 2 Step 7 |
| #2 `user_id` in Get/List-data/Create/Update SELECT+RETURNING (§Design 2) | Task 1 Steps 3–4 |
| #2 pgxmock rows gain `user_id` (§Design 2 Tests) | Task 1 Step 5 |
| #2 contract `UserId` assertions on both stores (§Design 2 Tests) | Task 1 Steps 1, 7 |
| #3 `Querier.Begin` + `withTx` (§Design 3) | Task 3 Step 3 |
| #3 `tagsInScope`/`setXTags` take `q Querier`; `setXTags` drop the check (§Design 3) | Task 3 Step 4 |
| #3 4 methods: validate → withTx(tagsInScope → write → setXTags) (§Design 3) | Task 3 Steps 5–6 |
| #3 `CreateTag`/`UpdateTag` NOT wrapped (§Design 3) | Not in Task 3's method list; stated in the task's Interfaces block |
| #3 pgxmock Begin/Commit/Rollback + failure tests (§Design 3 Tests) | Task 3 Step 7 |
| #3 contract "list empty after failed attach" (§Design 3 Tests) | Task 3 Steps 1, 8 |
| Sequencing #2 → #1 → #3 (§Sequencing) | Task order 1 → 2 → 3 |
| No migration / API / handler change (§Constraints) | No such files in any task |

No gaps.

**2. Placeholder scan** — Task 2 Step 7 is explicitly optional with a delete condition, not a placeholder. Task 3 Step 3's "FAIL to compile … Move on" describes an expected transient state during a multi-step task, with the fix in Step 4. All code steps carry real code or exact edit rules with worked examples. No "TBD" / "add error handling" / "similar to Task N".

**3. Type consistency** — `ownerPredicate(column string, scope model.OwnerScope, n int) (string, []any)` used identically in Task 2 Steps 3–5. `withTx(ctx, func(q Querier) error) error`, `tagsInScope(ctx, q Querier, scope, tagIds)`, `setProjectTags(ctx, q Querier, projectId, tagIds)` (no `scope`), `setTimespanTags(ctx, q Querier, timespanId, tagIds)` (no `scope`) consistent across Task 3 Steps 3–6. `projectCols` / `timespanCols` / new `tagCols` gain `"user_id"` as the last element in Task 1 Step 5 and are referenced unchanged after. `Querier` gains exactly `Begin(ctx context.Context) (pgx.Tx, error)` in Task 3 Step 3 and nothing else.
