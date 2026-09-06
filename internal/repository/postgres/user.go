package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/larssonoliver/inundated/internal/model"
)

// CreateUser implements [repository.Repository].
func (r *PostgresStore) CreateUser(ctx context.Context, user model.User) (model.User, error) {
	if user.Sub == "" {
		return model.User{}, fmt.Errorf("CreateUser: sub must not be empty: %w", model.ErrInvalidArgument)
	}
	if user.Email == "" {
		return model.User{}, fmt.Errorf("CreateUser: email must not be empty: %w", model.ErrInvalidArgument)
	}
	if user.Id == uuid.Nil {
		user.Id = uuid.New()
	}

	const q = `
		INSERT INTO users (id, sub, email, name)
		VALUES ($1, $2, $3, $4)
		RETURNING id, sub, email, name`

	var created model.User
	err := r.db.QueryRow(ctx, q, user.Id, user.Sub, user.Email, user.Name).
		Scan(&created.Id, &created.Sub, &created.Email, &created.Name)
	if err != nil {
		if isUniqueViolation(err) {
			return model.User{}, fmt.Errorf("CreateUser: sub already exists: %w", model.ErrAlreadyExists)
		}
		return model.User{}, fmt.Errorf("CreateUser: %w", err)
	}
	return created, nil
}

// CreateUserAdoptingOrphans implements [repository.UserRepository].
//
// The whole operation runs as a single statement so that the "is this the first
// user" check and the ownership hand-over are atomic. Sibling CTEs do not see
// new_user's insert, so is_first reflects the table state before this statement.
// Two racing first logins therefore both compute is_first = true and both issue
// the UPDATE; row locks serialise them, and the loser re-evaluates
// "user_id IS NULL" and updates nothing. Whoever commits first adopts everything.
func (r *PostgresStore) CreateUserAdoptingOrphans(ctx context.Context, user model.User) (model.User, model.OrphanAdoption, error) {
	if user.Sub == "" {
		return model.User{}, model.OrphanAdoption{}, fmt.Errorf("CreateUserAdoptingOrphans: sub must not be empty: %w", model.ErrInvalidArgument)
	}
	if user.Email == "" {
		return model.User{}, model.OrphanAdoption{}, fmt.Errorf("CreateUserAdoptingOrphans: email must not be empty: %w", model.ErrInvalidArgument)
	}
	if user.Id == uuid.Nil {
		user.Id = uuid.New()
	}

	const q = `
		WITH is_first AS (
			SELECT NOT EXISTS (SELECT 1 FROM users) AS v
		),
		new_user AS (
			INSERT INTO users (id, sub, email, name)
			VALUES ($1, $2, $3, $4)
			RETURNING id, sub, email, name
		),
		adopted_projects AS (
			UPDATE projects SET user_id = $1
			WHERE (SELECT v FROM is_first) AND user_id IS NULL
			RETURNING 1
		),
		adopted_tags AS (
			UPDATE tags SET user_id = $1
			WHERE (SELECT v FROM is_first) AND user_id IS NULL
			RETURNING 1
		),
		adopted_timespans AS (
			UPDATE timespans SET user_id = $1
			WHERE (SELECT v FROM is_first) AND user_id IS NULL
			RETURNING 1
		)
		SELECT
			nu.id, nu.sub, nu.email, nu.name,
			(SELECT count(*) FROM adopted_projects)::int,
			(SELECT count(*) FROM adopted_tags)::int,
			(SELECT count(*) FROM adopted_timespans)::int
		FROM new_user nu`

	var created model.User
	var adoption model.OrphanAdoption
	err := r.db.QueryRow(ctx, q, user.Id, user.Sub, user.Email, user.Name).
		Scan(&created.Id, &created.Sub, &created.Email, &created.Name,
			&adoption.Projects, &adoption.Tags, &adoption.Timespans)
	if err != nil {
		if isUniqueViolation(err) {
			return model.User{}, model.OrphanAdoption{}, fmt.Errorf("CreateUserAdoptingOrphans: sub already exists: %w", model.ErrAlreadyExists)
		}
		return model.User{}, model.OrphanAdoption{}, fmt.Errorf("CreateUserAdoptingOrphans: %w", err)
	}
	return created, adoption, nil
}

// GetUser implements [repository.Repository].
func (r *PostgresStore) GetUser(ctx context.Context, id uuid.UUID) (model.User, error) {
	if id == uuid.Nil {
		return model.User{}, fmt.Errorf("GetUser: id: %w", model.ErrInvalidArgument)
	}

	const q = `
		SELECT id, sub, email, name
		FROM users
		WHERE id = $1`

	var u model.User
	err := r.db.QueryRow(ctx, q, id).Scan(&u.Id, &u.Sub, &u.Email, &u.Name)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, fmt.Errorf("GetUser %s: %w", id, model.ErrNotFound)
	}
	if err != nil {
		return model.User{}, fmt.Errorf("GetUser: %w", err)
	}
	return u, nil
}

// GetUserBySub implements [repository.Repository].
func (r *PostgresStore) GetUserBySub(ctx context.Context, sub string) (model.User, error) {
	if sub == "" {
		return model.User{}, fmt.Errorf("GetUserBySub: sub must not be empty: %w", model.ErrInvalidArgument)
	}

	const q = `
		SELECT id, sub, email, name
		FROM users
		WHERE sub = $1`

	var u model.User
	err := r.db.QueryRow(ctx, q, sub).Scan(&u.Id, &u.Sub, &u.Email, &u.Name)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, fmt.Errorf("GetUserBySub %s: %w", sub, model.ErrNotFound)
	}
	if err != nil {
		return model.User{}, fmt.Errorf("GetUserBySub: %w", err)
	}
	return u, nil
}

// UpdateUser implements [repository.Repository].
func (r *PostgresStore) UpdateUser(ctx context.Context, user model.User) (model.User, error) {
	if user.Id == uuid.Nil {
		return model.User{}, fmt.Errorf("UpdateUser: id: %w", model.ErrInvalidArgument)
	}
	if user.Email == "" {
		return model.User{}, fmt.Errorf("UpdateUser: email must not be empty: %w", model.ErrInvalidArgument)
	}

	const q = `
		UPDATE users
		SET email = $2, name = $3
		WHERE id = $1
		RETURNING id, sub, email, name`

	var updated model.User
	err := r.db.QueryRow(ctx, q, user.Id, user.Email, user.Name).
		Scan(&updated.Id, &updated.Sub, &updated.Email, &updated.Name)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, fmt.Errorf("UpdateUser %s: %w", user.Id, model.ErrNotFound)
	}
	if err != nil {
		return model.User{}, fmt.Errorf("UpdateUser: %w", err)
	}
	return updated, nil
}
