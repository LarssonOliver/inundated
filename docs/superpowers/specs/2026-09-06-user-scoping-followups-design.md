# User-scoping follow-ups — design

**Branch:** `data-model-user-scoping` (commits added to the open PR #103)
**Follows:** `docs/superpowers/specs/2026-09-06-data-model-user-scoping-design.md`
**Status:** design approved, pending implementation plan

## Problem

The data-model user-scoping work (PR #103) shipped with three documented
limitations:

1. **Index defeat.** The scoped-list and aggregate queries use
   `user_id IS NOT DISTINCT FROM $N`, which has no btree strategy, so Postgres
   seq-scans `tags` / `projects` / `timespans` even though migration `0006`
   created `idx_{tags,projects,timespans}_user_id`.
2. **Store divergence.** The Postgres store never selects `user_id` back, so
   `model.{Tag,Project,Timespan}.UserId` is always `nil` from Postgres reads but
   populated from the memory store. A contract test asserting `UserId` would fail
   only on Postgres.
3. **Non-transactional writes.** `CreateProject` / `CreateTimespan` insert the row
   and *then* attach tags; a foreign-tag rejection (`ErrInvalidReference`) leaves
   an orphaned parent row the user sees in their own list after a 4xx.

This change fixes all three. No schema change, no API change.

## Constraints

- No new migration. The existing `0006` indexes serve both `= $N` and `IS NULL`.
- No API / OpenAPI / handler change. Cross-user access still returns
  `ErrNotFound`; no owner field is added to any response.
- The memory store is already correct for #1 (in-memory filter) and #3 (atomic
  under its mutex). Only #2 requires a memory-side assertion (it already returns
  `UserId`); the memory store is otherwise untouched.
- Keep existing repository patterns: `const q` static SQL where practical,
  `model.ErrNotFound` / `ErrInvalidReference` / `ErrInvalidArgument`, pgxmock unit
  tests, memory+postgres contract tests.

## Design

### 1. Index-friendly owner predicate (#1)

New helper in `internal/repository/postgres/helpers.go`:

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

**Applied to the 5 multi-row query sites only:**

| File | Method | Sites |
|---|---|---|
| `postgres/tag.go` | `ListTags` | count query, data query |
| `postgres/project.go` | `ListProjects` | count query, data query |
| `postgres/timespan.go` | `ListTimespans` | count query, data query |
| `postgres/timespan.go` | `GetTotalDurationByTags` | 1 |
| `postgres/timespan.go` | `AggregateTimeSpentByTagsAndBuckets` | 1 (the `matching_timespans` CTE) |

Each site changes from a `const q` string to a string built with
`ownerPredicate`, with the owner arg moved to the **end** of the arg list.
Concrete shapes:

- **`ListX` count query:** `SELECT COUNT(*) FROM <table> WHERE deleted_at IS NULL
  AND <ownerPredicate(col, scope, 1)>`; args `= ownerArgs`.
- **`ListX` data query:** `... WHERE deleted_at IS NULL AND <ownerPredicate(col,
  scope, 3)> ORDER BY <existing> LIMIT $1 OFFSET $2`; args `= [Limit, Offset,
  ownerArgs...]`.
- **`GetTotalDurationByTags`:** owner predicate at `$2`
  (`ownerPredicate("t.user_id", scope, 2)`), `tag_id = ANY($1)` unchanged; args
  `= [tagIds, ownerArgs...]`.
- **`AggregateTimeSpentByTagsAndBuckets`:** owner predicate at `$4` in the
  `matching_timespans` CTE (`ownerPredicate("t.user_id", scope, 4)`); args
  `= [tagIds, bucketStarts, bucketEnds, ownerArgs...]`.

**`GetX` / `UpdateX` / `DeleteX` keep `IS NOT DISTINCT FROM $N`** — they are
driven by `WHERE id = $1` (primary key), so the owner clause filters a single
already-fetched row; no index is involved. The helper comment and a note in this
spec record why the two paths differ.

**Tests:**
- `postgres/{tag,project,timespan}_test.go`: at each of the 5 sites, update the
  `mock.ExpectQuery` regex and `WithArgs` for BOTH the scoped variant (regex
  contains `user_id = \$N`, args include the uuid) and the unowned variant (regex
  contains `user_id IS NULL`, args omit it). Where a test previously used one
  `testScope`, add an unowned-scope counterpart so both branches are covered.
- Optional single contract test (`postgres` only): seed ~500 rows split across
  two `UserScope`s, run `EXPLAIN (FORMAT JSON)` on the scoped `ListTags` data
  query, assert the plan node is an Index/Bitmap Index Scan on
  `idx_tags_user_id`, not a `Seq Scan`. If it proves planner-flaky in CI, delete
  it — the pgxmock SQL-shape assertions are the regression guard.

### 2. Postgres returns `user_id` (#2)

`model.{Tag,Project,Timespan}.UserId *uuid.UUID` already exists (used by the
memory store's `matchesScope`). Add `user_id` to the Postgres store's read/return
lists:

| Method | Change |
|---|---|
| `GetTag` / `GetProject` / `GetTimespan` | `SELECT … , user_id` + `.Scan(…, &x.UserId)` |
| `ListTags` / `ListProjects` / `ListTimespans` data query | `SELECT … , user_id` + scan into each row's `UserId` |
| `CreateTag` / `CreateProject` / `CreateTimespan` | `RETURNING … , user_id` + scan |
| `UpdateTag` / `UpdateProject` / `UpdateTimespan` | `RETURNING … , user_id` + scan (`user_id` stays out of `SET`) |

`*uuid.UUID` scans a nullable `uuid` column directly (NULL → nil pointer). The
aggregate methods return no entity rows and are unaffected.

**Tests:**
- Every affected pgxmock `NewRows` gains a `"user_id"` column and its `AddRow`
  gains a value (`uuid.New()` or `nil`).
- Contract tests: the `CreateAndGet` and `ScopeIsolation` subtests for each
  resource gain `require.Equal(t, *scope.UserID(), *got.UserId)` for a scoped
  create/get and `require.Nil(t, got.UserId)` for an unowned one. These now
  execute against **both** stores, turning `UserId` into a real cross-store
  contract.

### 3. Transactional create/update (#3)

**`Querier` gains `Begin`:**

```go
type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}
```

`*pgxpool.Pool` and `pgxmock.PgxPoolIface` both already implement `Begin`, so the
`var _ Querier = (*pgxpool.Pool)(nil)` assertion and `NewPostgresStoreWithQuerier`
still compile. `pgx.Tx` itself satisfies `Querier` (it has `QueryRow`/`Query`/
`Exec`/`Begin`), so the transaction handle can be passed straight into the
tx-scoped helpers.

**`withTx` helper** (`internal/repository/postgres/postgres.go` or `helpers.go`):

```go
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

**Helper signature changes** (write path only; read helpers untouched):

- `tagsInScope(ctx, tagIds)` → `tagsInScope(ctx, q Querier, scope, tagIds)` — uses
  `q` instead of `r.db`.
- `setProjectTags(ctx, scope, projectId, tagIds)` →
  `setProjectTags(ctx, q Querier, projectId, tagIds)` — **drops** its internal
  `tagsInScope` call (the caller now runs it first inside the same tx); becomes a
  pure DELETE-then-INSERT of the join rows against `q`.
- `setTimespanTags` — same shape as `setProjectTags`.
- `projectTagIds` / `timespanTagIds` — **unchanged**; only called from read paths
  (`GetX` / `ListX`), never inside a tx.

Only these four methods are wrapped — they each run two or more statements
(parent write + join-row replace, and the tag check when tags are supplied).
`CreateTag` / `UpdateTag` are single-statement and stay as they are.

**`CreateProject` / `UpdateProject` / `CreateTimespan` / `UpdateTimespan`
restructure:**

```
1. validate inputs (name/time-range checks, id generation) — before any DB work
2. r.withTx(ctx, func(q Querier) error {
       ok, err := r.tagsInScope(ctx, q, scope, project.TagIds)   // first: validate tag ownership
       if err != nil { return err }
       if !ok { return fmt.Errorf("<Method>: %w", model.ErrInvalidReference) }

       // INSERT / UPDATE the parent row via q  (existing SQL + the #1/#2 changes)
       //   → scan RETURNING (incl. user_id from #2)
       //   → ErrNotFound on pgx.ErrNoRows for Update

       return r.setProjectTags(ctx, q, created.Id, project.TagIds)
   })
```

Tag validation is the **first statement inside the transaction**, so a foreign
tag returns `ErrInvalidReference` before the parent INSERT/UPDATE runs and the
transaction rolls back with nothing written. (A tag deleted or reassigned in the
narrow window between the check and the join-row INSERT is still caught by
`project_tags_tag_id_fkey` if it is a hard delete; a soft delete or reassignment
in that window would leave a join row to an unreachable tag — harmless, `GetTag`
returns 404 for it. A `FOR UPDATE` lock would close this and is explicitly out of
scope.)

**Tests:**
- `postgres/{project,timespan}_test.go`: existing `Create*` / `Update*` tests
  gain `mock.ExpectBegin()` before and `mock.ExpectCommit()` after their query
  expectations. For tests whose input supplies tag ids, the `tagsInScope`
  `SELECT count(*)` expectation moves to immediately after `ExpectBegin`; for
  tests with no tags, `tagsInScope` short-circuits (`len(tagIds) == 0`) so no
  count query is expected.
- New per-method failure test: `ExpectBegin` → `tagsInScope` `SELECT count(*)`
  returns a short count → `ExpectRollback` → assert `ErrInvalidReference` and
  that no parent INSERT/UPDATE was expected (pgxmock `ExpectationsWereMet` fails
  if the code runs an unexpected query, so omitting the INSERT expectation
  proves it never ran).
- Contract tests: the existing `CannotAttachAnotherUsersTag` subtests (project +
  timespan, both create and update paths) gain a trailing assertion that
  `ListProjects` / `ListTimespans` under `scopeA` returns zero rows after the
  failed operation — proving the real Postgres transaction rolled back (and that
  the memory store, already atomic, behaves the same).

## Implementation sequencing

Bottom-up; each phase compiles and its tests pass before the next.

1. **#2 — Postgres returns `user_id`.** Isolated; adds a column to read/return
   lists + the cross-store contract assertions.
2. **#1 — `ownerPredicate` helper + the 5 query sites.** Depends on #2 only in
   that the `ListX` data query SELECT list is already `… , user_id` by then.
3. **#3 — `Querier.Begin` + `withTx` + the 4 create/update methods + helper
   signature changes.** Goes last: it wraps the INSERT/UPDATE + `RETURNING`
   (shaped by #2) and reuses `tagsInScope` (unchanged by #1).

Each phase is one commit on `data-model-user-scoping`.

## Out of scope

- Fix #4 (the `TestUserScopedModelsHaveIsolationCoverage` textual/AST guard) — it
  is a deliberate drift tripwire, not a proof; leaving it.
- `FOR UPDATE` / fully TOCTOU-free tag validation.
- Any migration, API, OpenAPI, or handler change.
- Threading `Begin` through the read helpers or the memory store.
