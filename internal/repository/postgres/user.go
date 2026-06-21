package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
)

// CreateUser implements [repository.Repository].
func (r *PostgresStore) CreateUser(ctx context.Context, user model.User) error {
	panic("unimplemented")
}

// GetUser implements [repository.Repository].
func (r *PostgresStore) GetUser(ctx context.Context, id uuid.UUID) (model.User, error) {
	panic("unimplemented")
}

// GetUserBySub implements [repository.Repository].
func (r *PostgresStore) GetUserBySub(ctx context.Context, sub string) (model.User, error) {
	panic("unimplemented")
}

// UpdateUser implements [repository.Repository].
func (r *PostgresStore) UpdateUser(ctx context.Context, user model.User) (model.User, error) {
	panic("unimplemented")
}
