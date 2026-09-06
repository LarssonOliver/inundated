package contract_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/stretchr/testify/require"
)

var testScope = model.UserScope(uuid.MustParse("11111111-1111-1111-1111-111111111111"))

// seedScopeUser inserts the users row that a UserScope's user_id foreign key
// references, so owned resources can be created under that scope against a real
// database. It is a no-op for the unowned scope and idempotent for a scope whose
// user already exists.
func seedScopeUser(t *testing.T, ctx context.Context, repo repository.Repository, scope model.OwnerScope) {
	t.Helper()
	id := scope.UserID()
	if id == nil {
		return
	}
	_, err := repo.CreateUser(ctx, model.User{
		Id:    *id,
		Sub:   "scope|" + id.String(),
		Email: id.String() + "@scope.test",
		Name:  "scope user",
	})
	if errors.Is(err, model.ErrAlreadyExists) {
		return
	}
	require.NoError(t, err)
}

func seedTags(
	t *testing.T,
	ctx context.Context,
	repo repository.TagRepository,
	scope model.OwnerScope,
	n int,
) []uuid.UUID {
	t.Helper()

	ids := make([]uuid.UUID, n)
	for i := range n {
		tag := model.Tag{
			Name:  fmt.Sprintf("tag-%d", i),
			Color: "#123456",
		}
		created, err := repo.CreateTag(ctx, scope, tag)
		require.NoError(t, err)
		ids[i] = created.Id
	}
	return ids
}

// seedOrphanResources creates the given number of tags, projects and timespans,
// all with a nil UserId (the create path does not assign ownership yet).
func seedOrphanResources(
	t *testing.T,
	ctx context.Context,
	repo repository.Repository,
	nTags, nProjects, nTimespans int,
) {
	t.Helper()

	seedTags(t, ctx, repo, model.UnownedScope(), nTags)

	for i := range nProjects {
		_, err := repo.CreateProject(ctx, model.UnownedScope(), model.Project{
			Name:  fmt.Sprintf("project-%d", i),
			Color: "#123456",
		})
		require.NoError(t, err)
	}

	base := time.Now().UTC().Truncate(time.Millisecond)
	for i := range nTimespans {
		start := base.Add(time.Duration(i) * time.Hour)
		_, err := repo.CreateTimespan(ctx, model.UnownedScope(), model.Timespan{
			Name:      fmt.Sprintf("timespan-%d", i),
			StartTime: start,
			EndTime:   start.Add(30 * time.Minute),
		})
		require.NoError(t, err)
	}
}

func assertPage[T any](t *testing.T, page model.Page[T], wantLen, wantTotal int) {
	t.Helper()
	require.Len(t, page.Data, wantLen)
	require.Equal(t, wantTotal, page.TotalCount)
}
