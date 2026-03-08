package postgres

import (
	"context"
	"database/sql"
	"embed"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func ApplyMigrations(ctx context.Context, dsn string) error {
	m, cleanup, err := newMigrator(dsn)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return cleanup()
}

func ApplyMigrationsUpTo(ctx context.Context, dsn string, version uint) error {
	if version == 0 {
		return nil
	}

	m, cleanup, err := newMigrator(dsn)
	if err != nil {
		return err
	}

	if err := m.Migrate(version); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return cleanup()
}

func ApplyMigrationsSteps(ctx context.Context, dsn string, steps int) error {
	m, cleanup, err := newMigrator(dsn)
	if err != nil {
		return err
	}

	if err := m.Steps(steps); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return cleanup()
}

func newMigrator(dsn string) (*migrate.Migrate, func() error, error) {
	fs, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return nil, nil, err
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, nil, err
	}

	driver, err := pgx.WithInstance(db, &pgx.Config{})
	if err != nil {
		_ = db.Close()
		return nil, nil, err
	}

	m, err := migrate.NewWithInstance("iofs", fs, "pgx", driver)
	if err != nil {
		_ = db.Close()
		return nil, nil, err
	}

	cleanup := func() error {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			return srcErr
		}
		return dbErr
	}

	return m, cleanup, nil
}
