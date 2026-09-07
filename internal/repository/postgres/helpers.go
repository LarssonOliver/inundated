package postgres

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/larssonoliver/inundated/internal/model"
)

// isUniqueViolation reports whether err is a Postgres unique-constraint violation.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// ownerPredicate appends the owner bind (if the scope is a user) to args and
// returns a WHERE fragment referencing its placeholder, so the placeholder
// number can never drift from the args slice. Pass the query's other binds as
// args; the owner bind always lands at $(len(args)+1). A scoped call yields
// `<column> = $n`, an unowned call yields `<column> IS NULL` and returns args
// unchanged. `= $n` / `IS NULL` (rather than `IS NOT DISTINCT FROM $n`) lets the
// btree idx_<table>_user_id index from migration 0006 apply.
func ownerPredicate(column string, scope model.OwnerScope, args []any) (sql string, newArgs []any) {
	if id := scope.UserID(); id != nil {
		return fmt.Sprintf("%s = $%d", column, len(args)+1), append(args, *id)
	}
	return column + " IS NULL", args
}
