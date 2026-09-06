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
