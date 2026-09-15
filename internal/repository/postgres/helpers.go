package postgres

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/larssonoliver/inundated/internal/model"
)

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func ownerPredicate(column string, scope model.OwnerScope, args []any) (sql string, newArgs []any) {
	if id := scope.UserID(); id != nil {
		return fmt.Sprintf("%s = $%d", column, len(args)+1), append(args, *id)
	}
	return column + " IS NULL", args
}

// archivedFilter returns the SQL fragment (with a trailing "AND ") that
// excludes archived rows, or an empty string when includeArchived is true.
// This is baked into the query text rather than bound as a parameter: an
// `archived_at IS NULL OR $n` predicate can't use an index on archived_at,
// and folding it into a query-shaped string avoids hand-syncing its
// placeholder number against the rest of the positional args.
func archivedFilter(includeArchived bool) string {
	if includeArchived {
		return ""
	}
	return "archived_at IS NULL AND "
}

// noAssociatedTags is a tagsInScope alreadyAssociated getter for a brand new
// project/timespan, which by definition has no pre-existing associations.
func noAssociatedTags() ([]uuid.UUID, error) {
	return nil, nil
}
