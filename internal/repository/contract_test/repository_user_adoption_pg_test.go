package contract_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository/postgres"
	"github.com/larssonoliver/inundated/test/testutils"
	"github.com/stretchr/testify/require"
)

// countOwnedBy counts rows in table with the given user_id.
func countOwnedBy(t *testing.T, ctx context.Context, pool *pgxpool.Pool, table string, userID uuid.UUID) int {
	t.Helper()
	var n int
	err := pool.QueryRow(ctx, fmt.Sprintf("SELECT count(*) FROM %s WHERE user_id = $1", table), userID).Scan(&n)
	require.NoError(t, err)
	return n
}

func TestPostgres_CreateUserAdoptingOrphans_PopulatesUserIdColumn(t *testing.T) {
	ctx := context.Background()
	pool := testutils.StartPostgresContainerWithMigrationsApplied(ctx, t)
	repo := postgres.NewPostgresStoreFromPool(pool)

	const perType = 3
	seedOrphanResources(t, ctx, repo, perType, perType, perType)

	user := model.User{Id: uuid.New(), Sub: "auth0|first", Email: "first@example.com", Name: "First"}
	_, adoption, err := repo.CreateUserAdoptingOrphans(ctx, user)
	require.NoError(t, err)
	require.Equal(t, perType*len(userScopedModels), adoption.Total())

	for _, table := range userScopedModels {
		require.Equal(t, perType, countOwnedBy(t, ctx, pool, table, user.Id), "table %q", table)
	}
}

// Two logins racing to be the first must not split ownership: row locks in the
// adoption UPDATE serialise them, and whoever commits second updates nothing.
func TestPostgres_CreateUserAdoptingOrphans_ConcurrentFirstLoginsClaimOnce(t *testing.T) {
	ctx := context.Background()
	pool := testutils.StartPostgresContainerWithMigrationsApplied(ctx, t)
	repo := postgres.NewPostgresStoreFromPool(pool)

	const nProjects = 25
	seedOrphanResources(t, ctx, repo, 0, nProjects, 0)

	const nUsers = 6
	type result struct {
		id       uuid.UUID
		adoption model.OrphanAdoption
	}
	results := make(chan result, nUsers)
	var wg sync.WaitGroup
	for i := range nUsers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			u := model.User{
				Id:    uuid.New(),
				Sub:   fmt.Sprintf("auth0|user-%d", i),
				Email: fmt.Sprintf("user-%d@example.com", i),
				Name:  "User",
			}
			created, adoption, err := repo.CreateUserAdoptingOrphans(ctx, u)
			require.NoError(t, err)
			results <- result{id: created.Id, adoption: adoption}
		}(i)
	}
	wg.Wait()
	close(results)

	adopters := 0
	var winner uuid.UUID
	for r := range results {
		if r.adoption.Projects > 0 {
			adopters++
			winner = r.id
			require.Equal(t, nProjects, r.adoption.Projects)
		}
	}
	require.Equal(t, 1, adopters, "exactly one racing user should adopt the orphans")
	require.Equal(t, nProjects, countOwnedBy(t, ctx, pool, "projects", winner))

	var unowned int
	require.NoError(t, pool.QueryRow(ctx, "SELECT count(*) FROM projects WHERE user_id IS NULL").Scan(&unowned))
	require.Equal(t, 0, unowned)
}
