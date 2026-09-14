package e2e_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestTag_GetStats_Contract200(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	tagResp, err := client.CreateTagWithResponse(ctx, CreateTagJSONRequestBody{
		Name:  "Stats Tag",
		Color: "#123456",
	})
	require.NoError(t, err)
	require.Equal(t, 201, tagResp.StatusCode())
	tagID := tagResp.JSON201.Id

	start := time.Now().UTC().Truncate(time.Second).Add(-2 * time.Hour)
	end := start.Add(90 * time.Minute)
	timespanName := "Stats Timespan"

	timespanResp, err := client.CreateTimespanWithResponse(ctx, CreateTimespanJSONRequestBody{
		Name:      &timespanName,
		StartTime: start,
		EndTime:   end,
		TagIds:    &[]TagIdPath{tagID},
	})
	require.NoError(t, err)
	require.Equal(t, 201, timespanResp.StatusCode())

	interval := start.Format(time.RFC3339) + "/" + end.Format(time.RFC3339)
	granularity := "PT1H"
	timezone := "UTC"

	statsResp, err := client.GetTagStatsWithResponse(ctx, tagID, &GetTagStatsParams{
		Metric:      TimeSpent,
		Interval:    &interval,
		Granularity: &granularity,
		Timezone:    &timezone,
	})
	require.NoError(t, err)
	require.Equal(t, 200, statsResp.StatusCode())
	require.NotNil(t, statsResp.JSON200)

	stats := statsResp.JSON200
	require.Equal(t, tagID, stats.TagId)
	require.Equal(t, StatsMetric(TimeSpent), stats.Metric)
	require.Equal(t, interval, stats.Interval)
	require.Equal(t, granularity, stats.Granularity)
	require.NotEmpty(t, stats.Unit)
	require.NotEmpty(t, stats.Series)

	for _, point := range stats.Series {
		require.NotEmpty(t, point.Interval)
		require.GreaterOrEqual(t, point.Value, float32(0))
		parts := strings.Split(point.Interval, "/")
		require.Len(t, parts, 2)
		_, startErr := time.Parse(time.RFC3339, parts[0])
		require.NoError(t, startErr)
		_, endErr := time.Parse(time.RFC3339, parts[1])
		require.NoError(t, endErr)
	}
}

func TestTag_GetStats_Semantics(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	createTag := func(name, color string) TagIdPath {
		resp, err := client.CreateTagWithResponse(ctx, CreateTagJSONRequestBody{
			Name:  name,
			Color: color,
		})
		require.NoError(t, err)
		require.Equal(t, 201, resp.StatusCode())

		return resp.JSON201.Id
	}

	tagA := createTag("Tag Stats Semantic Tag A", "#aa1111")
	tagB := createTag("Tag Stats Semantic Tag B", "#11aa11")

	base := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)

	name1 := "tag-semantic-ts1"
	_, err := client.CreateTimespanWithResponse(ctx, CreateTimespanJSONRequestBody{
		Name:      &name1,
		StartTime: base.Add(15 * time.Minute),
		EndTime:   base.Add(75 * time.Minute),
		TagIds:    &[]TagIdPath{tagA},
	})
	require.NoError(t, err)

	name2 := "tag-semantic-ts2-ignored"
	_, err = client.CreateTimespanWithResponse(ctx, CreateTimespanJSONRequestBody{
		Name:      &name2,
		StartTime: base,
		EndTime:   base.Add(2 * time.Hour),
		TagIds:    &[]TagIdPath{tagB},
	})
	require.NoError(t, err)

	interval := base.Format(time.RFC3339) + "/" + base.Add(2*time.Hour).Format(time.RFC3339)
	granularity := "PT1H"
	timezone := "UTC"

	statsResp, err := client.GetTagStatsWithResponse(ctx, tagA, &GetTagStatsParams{
		Metric:      TimeSpent,
		Interval:    &interval,
		Granularity: &granularity,
		Timezone:    &timezone,
	})
	require.NoError(t, err)
	require.Equal(t, 200, statsResp.StatusCode())
	require.NotNil(t, statsResp.JSON200)

	stats := statsResp.JSON200
	require.Len(t, stats.Series, 2)

	expectedIntervals := []string{
		base.Format(time.RFC3339) + "/" + base.Add(1*time.Hour).Format(time.RFC3339),
		base.Add(1*time.Hour).Format(time.RFC3339) + "/" + base.Add(2*time.Hour).Format(time.RFC3339),
	}

	expectedValues := []float64{
		45 * 60, // only ts1's overlap with the first bucket; ts2 (tagB) is not requested
		15 * 60, // ts1's tail in the second bucket
	}

	for i := range stats.Series {
		require.Equal(t, expectedIntervals[i], stats.Series[i].Interval)
		require.InDelta(t, expectedValues[i], float64(stats.Series[i].Value), 0.0001)
	}
}

func TestTag_GetStats_Contract422(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	tagResp, err := client.CreateTagWithResponse(ctx, CreateTagJSONRequestBody{
		Name:  "Stats Tag 422",
		Color: "#123456",
	})
	require.NoError(t, err)
	require.Equal(t, 201, tagResp.StatusCode())
	tagID := tagResp.JSON201.Id

	point := time.Now().UTC().Truncate(time.Second).Add(-2 * time.Hour)
	invalidInterval := point.Format(time.RFC3339) + "/" + point.Format(time.RFC3339)
	granularity := "PT1H"
	timezone := "UTC"

	statsResp, err := client.GetTagStatsWithResponse(ctx, tagID, &GetTagStatsParams{
		Metric:      TimeSpent,
		Interval:    &invalidInterval,
		Granularity: &granularity,
		Timezone:    &timezone,
	})
	require.NoError(t, err)
	require.Equal(t, 422, statsResp.StatusCode())
}

func TestTag_GetStats_Contract404(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	granularity := "PT1H"
	timezone := "UTC"

	statsResp, err := client.GetTagStatsWithResponse(ctx, uuid.New(), &GetTagStatsParams{
		Metric:      TimeSpent,
		Granularity: &granularity,
		Timezone:    &timezone,
	})
	require.NoError(t, err)
	require.Equal(t, 404, statsResp.StatusCode())
}
