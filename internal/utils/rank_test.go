package utils_test

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/larssonoliver/inundated/internal/utils"
	"github.com/stretchr/testify/require"
)

func TestRankBetween_Basics(t *testing.T) {
	first, err := utils.RankBetween("", "")
	require.NoError(t, err)
	require.NotEmpty(t, first)

	after, err := utils.RankBetween(first, "")
	require.NoError(t, err)
	require.Greater(t, after, first)

	before, err := utils.RankBetween("", first)
	require.NoError(t, err)
	require.Less(t, before, first)

	mid, err := utils.RankBetween(before, first)
	require.NoError(t, err)
	require.Greater(t, mid, before)
	require.Less(t, mid, first)
}

func TestRankBetween_AdjacentDigits(t *testing.T) {
	got, err := utils.RankBetween("a", "b")
	require.NoError(t, err)
	require.Greater(t, got, "a")
	require.Less(t, got, "b")
}

func TestRankBetween_RejectsBadInput(t *testing.T) {
	_, err := utils.RankBetween("b", "a")
	require.Error(t, err)

	_, err = utils.RankBetween("a", "a")
	require.Error(t, err)

	_, err = utils.RankBetween("a0", "")
	require.Error(t, err)

	_, err = utils.RankBetween("a-", "")
	require.Error(t, err)
}

// Random inserts into a list must keep it strictly ordered and never
// produce a rank ending in '0', which RankBetween would later reject.
func TestRankBetween_RandomInsertsStayOrdered(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	var ranks []string
	for range 2000 {
		i := rng.IntN(len(ranks) + 1)
		var lo, hi string
		if i > 0 {
			lo = ranks[i-1]
		}
		if i < len(ranks) {
			hi = ranks[i]
		}
		r, err := utils.RankBetween(lo, hi)
		require.NoError(t, err)
		require.NotEqual(t, byte('0'), r[len(r)-1])
		ranks = slices.Insert(ranks, i, r)
	}
	require.True(t, slices.IsSorted(ranks))
	for i := 1; i < len(ranks); i++ {
		require.NotEqual(t, ranks[i-1], ranks[i])
	}
}

func TestRankBetween_AppendingGrowsSlowly(t *testing.T) {
	last := ""
	for range 1000 {
		r, err := utils.RankBetween(last, "")
		require.NoError(t, err)
		last = r
	}
	require.LessOrEqual(t, len(last), 200)
}

func TestEvenRanks(t *testing.T) {
	require.Nil(t, utils.EvenRanks(0))

	for _, n := range []int{1, 2, 61, 62, 500, 5000} {
		ranks := utils.EvenRanks(n)
		require.Len(t, ranks, n)
		require.True(t, slices.IsSorted(ranks))
		for i, r := range ranks {
			require.NotEmpty(t, r)
			require.NotEqual(t, byte('0'), r[len(r)-1])
			require.LessOrEqual(t, len(r), 3)
			if i > 0 {
				require.NotEqual(t, ranks[i-1], r)
			}
		}
		// Room is left on both ends for later inserts.
		_, err := utils.RankBetween("", ranks[0])
		require.NoError(t, err)
		_, err = utils.RankBetween(ranks[n-1], "")
		require.NoError(t, err)
	}
}
