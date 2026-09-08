package postgres

import (
	"errors"
	"fmt"

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
