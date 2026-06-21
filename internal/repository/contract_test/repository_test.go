package contract_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/stretchr/testify/require"
)

func seedTags(
	t *testing.T,
	ctx context.Context,
	repo repository.TagRepository,
	n int,
) []uuid.UUID {
	t.Helper()

	ids := make([]uuid.UUID, n)
	for i := range n {
		tag := model.Tag{
			Name:  fmt.Sprintf("tag-%d", i),
			Color: "#123456",
		}
		created, err := repo.CreateTag(ctx, tag)
		require.NoError(t, err)
		ids[i] = created.Id
	}
	return ids
}

func assertPage[T any](t *testing.T, page model.Page[T], wantLen, wantTotal int) {
	t.Helper()
	require.Len(t, page.Data, wantLen)
	require.Equal(t, wantTotal, page.TotalCount)
}
