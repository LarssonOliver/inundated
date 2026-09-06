# Data model user-scoping — design

**Branch:** `data-model-user-scoping` (off `users`)
**PR #85 checklist item:** "Data model migration"
**Status:** design approved, pending implementation plan

## Problem

`user_id` exists on `projects`, `tags`, and `timespans` (migration `0006`), and the
first user to log in adopts every pre-existing (`user_id IS NULL`) row
(`CreateUserAdoptingOrphans`). But no read or write path filters on `user_id`
yet: any authenticated user can list, fetch, mutate, and delete every other
user's projects, tags, and timespans, and can attach another user's tags to
their own resources.

This change threads an ownership scope through the repository layer so that every
resource operation is confined to the current user — or, when the server runs
without authentication, to the pool of unowned (`user_id IS NULL`) rows.

## Constraints

- **Userless mode must keep working.** The app can run with no OIDC configured
  (the e2e server mounts no auth middleware). In that mode there is no current
  user; operations act on the `user_id IS NULL` pool. This is symmetric with
  authenticated mode, where the scope is a specific user id.
- **No schema change.** `user_id` stays nullable (NULL is the unowned scope).
  The `idx_{projects,tags,timespans}_user_id` indexes from `0006` are adequate.
- **No API surface change.** OpenAPI schemas and HTTP handlers are untouched. The
  scope is derived server-side from the request context, never sent by the
  client. No owner/`user_id` field is added to any response.
- **Cross-user access returns 404**, indistinguishable from a non-existent id —
  falls out of the existing `ErrNotFound` handling.
- Follow existing repository patterns (soft-delete predicate, `ErrNotFound` /
  `ErrInvalidReference` / `ErrInvalidArgument`, pgxmock unit tests, memory +
  postgres contract tests).

## Design

### 1. `model.OwnerScope`

A small opaque value type identifying whose resources an operation may touch.

```go
package model

// OwnerScope identifies whose resources an operation may see or modify. The zero
// value is the unowned scope, used when the server runs without authentication.
type OwnerScope struct {
	userID *uuid.UUID
}

// UserScope scopes operations to a single authenticated user.
func UserScope(id uuid.UUID) OwnerScope { return OwnerScope{userID: &id} }

// UnownedScope scopes operations to rows with no owner (user_id IS NULL).
func UnownedScope() OwnerScope { return OwnerScope{} }

// UserID returns the scoped user id, or nil for the unowned scope.
func (s OwnerScope) UserID() *uuid.UUID { return s.userID }
```

The pointer is fully encapsulated; callers only ever see `UserScope(id)` /
`UnownedScope()` and `UserID()`. It is deliberately not a bare `uuid.UUID` with a
`uuid.Nil` sentinel — the compiler tracks the type through every call site, and
an accidental zero value is the safe (unowned) scope rather than an ambiguous
one.

### 2. Scope resolution (service package)

A single helper is the one place that reads the current user:

```go
func ownerScope(ctx context.Context) model.OwnerScope {
	if u, ok := model.GetCurrentUserFromContext(ctx); ok && u.Id != uuid.Nil {
		return model.UserScope(u.Id)
	}
	return model.UnownedScope()
}
```

Every project / tag / timespan / stats service method calls `ownerScope(ctx)` and
passes the result to its repository calls. `GetProjectStats` passes the same
scope to both `GetProject` and `AggregateTimeSpentByTagsAndBuckets`.

Service interface signatures do not change — they already take `ctx`. Handlers
and OpenAPI are untouched.

### 3. Repository interfaces

`scope model.OwnerScope` becomes the parameter immediately after `ctx` on every
method of:

- `TagRepository` — `GetTag`, `ListTags`, `CreateTag`, `UpdateTag`, `DeleteTag`
- `ProjectRepository` — `GetProject`, `ListProjects`, `CreateProject`, `UpdateProject`, `DeleteProject`
- `TimespanRepository` — `GetTimespan`, `ListTimespans`, `CreateTimespan`, `UpdateTimespan`, `DeleteTimespan`, `GetTotalDurationByTags`
- `ProjectStatsRepository` — `AggregateTimeSpentByTagsAndBuckets`

**Not** scoped: `UserRepository` (users are not user-owned; `CreateUserAdoptingOrphans`
is inherently cross-scope), `SessionRepository`, `LoginStateRepository`.

`RepoMock` (`internal/repository/repository_mock.go`) and `MockRepository`
(`internal/repository/postgres/mock.go`) are updated by hand: the `...Fn` function
fields and method signatures gain `scope`, bodies pass it through.

### 4. Postgres implementation

**Single predicate for both modes.** Every scoped query gains:

```sql
AND user_id IS NOT DISTINCT FROM $N
```

where `$N` binds `scope.UserID()` (a `*uuid.UUID`; pgx sends `nil` as SQL `NULL`).
`IS NOT DISTINCT FROM NULL` matches the unowned rows; `IS NOT DISTINCT FROM
'<uuid>'` matches that user's rows. No branching in the query builders.

- **`GetX`**: add the predicate to `WHERE id = $1 AND deleted_at IS NULL`. Not
  owned → `pgx.ErrNoRows` → `ErrNotFound` (unchanged path).
- **`ListX`**: add the predicate to both the `COUNT(*)` query and the data query.
- **`CreateX`**: add `user_id` to the column list, binding `scope.UserID()`.
- **`UpdateX` / `DeleteX`**: add the predicate to the `WHERE`. 0 rows affected →
  `ErrNotFound` (unchanged path).
- **Tag attachment** (`setProjectTags`, `setTimespanTags`, called on create and
  update of projects and timespans): before writing join rows, verify every tag
  id is in scope:

  ```sql
  SELECT count(*) FROM tags
  WHERE id = ANY($1) AND deleted_at IS NULL AND user_id IS NOT DISTINCT FROM $2
  ```

  `count != len(tagIds)` → `ErrInvalidReference` (the error these paths already
  return). This replaces reliance on the FK violation, which a real
  other-user tag would not trigger.
- **Aggregates** (`GetTotalDurationByTags`, `AggregateTimeSpentByTagsAndBuckets`):
  add `AND t.user_id IS NOT DISTINCT FROM $N` to the `timespans t` filter.
  Defense in depth — incoming `tagIds` already come from a scoped lookup — but it
  makes each method's isolation guarantee self-contained.

The join-table read helpers (`projectTagIds`, `timespanTagIds`) are unchanged:
the parent row they are called for has already passed the scope check, so their
results are already confined to the current scope.

No transaction/atomicity changes: scoping adds a predicate and a bind arg, nothing
that needs new locking.

### 5. Memory implementation

A shared helper:

```go
func matchesScope(userID *uuid.UUID, scope model.OwnerScope) bool {
	want := scope.UserID()
	if want == nil {
		return userID == nil
	}
	return userID != nil && *userID == *want
}
```

- **`GetX`**: match id **and** `matchesScope` → `ErrNotFound` on mismatch.
- **`ListX`**: filter the backing slice by `matchesScope` before pagination /
  total count.
- **`CreateX`**: set `row.UserId = scope.UserID()` on the stored struct.
- **`UpdateX` / `DeleteX`**: locate by id **and** `matchesScope` → `ErrNotFound`.
- **`tagsExist`** becomes `tagsExistInScope(scope, tagIds)`: every tag must exist
  **and** match the scope.
- **Aggregates**: skip timespans that fail `matchesScope` in the accumulation
  loop.

### 6. First-user adoption interaction

`CreateUserAdoptingOrphans` is unchanged — it still claims `user_id IS NULL` rows
for the first user. Afterward that user's `ownerScope(ctx)` returns
`UserScope(theirID)`, so they see exactly the rows they just adopted. A userless
instance always resolves to `UnownedScope()` and sees the same NULL rows it
always did.

## Testing

### Contract tests (`internal/repository/contract_test`, memory + postgres)

- The shared `run(repoName, newRepo)` harness threads a fixed test scope
  (`model.UserScope(uuidA)`) through every existing repository call (~100 sites,
  mechanical).
- New **isolation subtests** per resource (tags, projects, timespans, stats),
  with a second owner B:
  - `ListX` as A excludes B's rows; `TotalCount` reflects only A's.
  - `GetX` / `UpdateX` / `DeleteX` of B's row as A → `ErrNotFound`.
  - `CreateProject` / `CreateTimespan` referencing B's tag as A →
    `ErrInvalidReference`.
  - `model.UnownedScope()` sees only `user_id IS NULL` rows; `UserScope` never
    sees them.
  - `GetTotalDurationByTags` / `AggregateTimeSpentByTagsAndBuckets` as A never
    count B's timespans.
- `user_scoped_models_test.go` gains a table-driven assertion: for every entry in
  `userScopedModels`, a named isolation subtest must exist for that resource.
  Keeps "added a user-scoped resource, forgot to scope its queries" loud. This is
  a checklist-style assertion (test names present), not SQL introspection.

### Unit tests

- **Postgres (pgxmock)**: `tag_test.go`, `project_test.go`, `timespan_test.go` —
  update expected query patterns and `WithArgs` to include the new predicate and
  the scope bind arg; add an "other user's row → ErrNotFound" case per method.
- **Memory**: add `scope` args to existing calls; add direct cross-scope cases.

### Service tests

- `RepoMock` `...Fn` fields gain the `scope` parameter.
- New assertions: a `ctx` carrying a user resolves to `model.UserScope(thatID)`
  at the repo boundary; a bare `ctx` resolves to `model.UnownedScope()`.
- `project_stats` test: the same scope value reaches both `GetProject` and
  `AggregateTimeSpentByTagsAndBuckets`.

### e2e

The e2e server mounts no auth middleware, so it exercises the **unowned** path.
Existing e2e tests pass unchanged and now actively cover userless mode.
Authenticated-mode e2e is out of scope here — it depends on the separate
"Configuration of OIDC client" PR item — and is noted as follow-up.

### Mocks

`repository_mock.go` (`RepoMock`) and `postgres/mock.go` (`MockRepository`) are
updated by hand to the new signatures.

## Implementation sequencing

Bottom-up; each phase compiles and its tests pass before the next.

1. **`model.OwnerScope`** + type unit tests. No other changes.
2. **Repository interfaces + mocks** — add `scope` to the four interfaces and both
   mock types; bodies pass through. Build stays green (nothing consumes it yet).
3. **Tags** — postgres + memory scope enforcement; `tagsExistInScope`; contract
   isolation subtests; postgres/memory unit tests. First, because projects and
   timespans validate tag ownership.
4. **Projects** — repo methods + `setProjectTags` ownership guard; contract + unit
   tests.
5. **Timespans** — repo methods + `setTimespanTags` guard; contract + unit tests.
6. **Aggregates** — `GetTotalDurationByTags`,
   `AggregateTimeSpentByTagsAndBuckets`, `ProjectStatsRepository`; contract + unit
   tests.
7. **Service layer** — `ownerScope(ctx)` helper; thread through every
   project/tag/timespan/stats service method; service-test updates + ctx→scope
   assertions.
8. **Coverage guard** — extend `user_scoped_models_test.go`; final full suite +
   `golangci-lint` + e2e.

Phases 3–6 share a near-identical shape; phase 3 establishes the pattern.

## Out of scope

- Making `user_id` `NOT NULL`.
- Any OpenAPI / HTTP handler change.
- Authenticated-mode e2e coverage (blocked on OIDC config).
- Route protection / `RequireAuth` wiring in `main.go` (separate PR item).
- Sharing resources between users, roles, or org-level ownership.
