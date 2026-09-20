package postgres

import (
	"errors"
	"fmt"
	"strings"

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

// sqlConditionBuilder accumulates "AND <fragment> $N" clauses alongside
// their bound values, numbering each placeholder from args already added
// (optionally seeded via newSQLConditionBuilder) - so a caller adding a
// condition never has to compute $N by hand or keep it in sync with a
// sibling query built the same way.
//
// ownerPredicate isn't built on this type (it's used throughout this
// package with its own append-and-return-args convention), so it must
// still be called last, passed this builder's own Args(), to keep its
// placeholder numbering correct.
type sqlConditionBuilder struct {
	sql  strings.Builder
	args []any
}

// newSQLConditionBuilder seeds the builder with args already bound ahead of
// any condition it will add (e.g. LIMIT/OFFSET), so its own placeholder
// numbers continue that sequence instead of starting over at $1.
func newSQLConditionBuilder(seedArgs []any) *sqlConditionBuilder {
	return &sqlConditionBuilder{args: seedArgs}
}

// add appends " AND <fragment> $N" to the builder's SQL, bound to value.
// fragment is everything before the placeholder, e.g. "end_time >".
func (b *sqlConditionBuilder) add(fragment string, value any) {
	b.args = append(b.args, value)
	fmt.Fprintf(&b.sql, " AND %s $%d", fragment, len(b.args))
}

func (b *sqlConditionBuilder) SQL() string { return b.sql.String() }
func (b *sqlConditionBuilder) Args() []any { return b.args }
