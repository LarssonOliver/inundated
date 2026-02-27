package postgres_test

// helpers_test.go – shared test utilities used across tag/project/timespan tests.

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/repository/postgres"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

// newMock creates a fresh pgxmock pool and a Repository wired to it.
func newMock(t *testing.T) (repository.Repository, pgxmock.PgxPoolIface) {
	t.Helper()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled pgxmock expectations: %v", err)
		}
	})
	return postgres.NewPostgresStoreWithQuerier(mock), mock
}

// dur is a convenience helper for building *time.Duration literals.
func dur(d time.Duration) *time.Duration { return &d }

// ── fixture factories ────────────────────────────────────────────────────────

func aTag() model.Tag {
	return model.Tag{
		Id:    uuid.New(),
		Name:  "backend",
		Color: "#ff0000",
	}
}

func aProject() model.Project {
	return model.Project{
		Id:         uuid.New(),
		Name:       "Alpha",
		Color:      "#00ff00",
		TimeBudget: dur(8 * time.Hour),
		TagIds:     []uuid.UUID{uuid.New(), uuid.New()},
	}
}

func aTimeSpan() model.TimeSpan {
	now := time.Now().UTC().Truncate(time.Millisecond)
	return model.TimeSpan{
		Id:        uuid.New(),
		Name:      "morning session",
		StartTime: now,
		EndTime:   now.Add(2 * time.Hour),
		TagIds:    []uuid.UUID{uuid.New()},
	}
}


