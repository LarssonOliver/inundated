package e2e_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestProject_GetStats_Contract200(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	tagResp, err := client.CreateTagWithResponse(ctx, CreateTagJSONRequestBody{
		Name:  "Stats Tag",
		Color: "#123456",
	})
	require.NoError(t, err)
	require.Equal(t, 201, tagResp.StatusCode())

	tagIDs := []TagIdPath{tagResp.JSON201.Id}

	projectResp, err := client.CreateProjectWithResponse(ctx, CreateProjectJSONRequestBody{
		Name:   "Stats Project",
		Color:  "#654321",
		TagIds: &tagIDs,
	})
	require.NoError(t, err)
	require.Equal(t, 201, projectResp.StatusCode())
	projectID := projectResp.JSON201.Id

	start := time.Now().UTC().Truncate(time.Second).Add(-2 * time.Hour)
	end := start.Add(90 * time.Minute)
	timespanName := "Stats Timespan"

	timespanResp, err := client.CreateTimespanWithResponse(ctx, CreateTimespanJSONRequestBody{
		Name:      &timespanName,
		StartTime: start,
		EndTime:   end,
		TagIds:    &tagIDs,
	})
	require.NoError(t, err)
	require.Equal(t, 201, timespanResp.StatusCode())

	interval := start.Format(time.RFC3339) + "/" + end.Format(time.RFC3339)
	granularity := "PT1H"
	timezone := "UTC"

	statsResp, err := client.GetProjectStatsWithResponse(ctx, projectID, &GetProjectStatsParams{
		Metric:      TimeSpent,
		Interval:    &interval,
		Granularity: &granularity,
		Timezone:    &timezone,
	})
	require.NoError(t, err)
	require.Equal(t, 200, statsResp.StatusCode())
	require.NotNil(t, statsResp.JSON200)

	stats := statsResp.JSON200
	require.Equal(t, projectID, stats.ProjectId)
	require.Equal(t, ProjectStatsMetricTimeSpent, stats.Metric)
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

func TestProject_GetStats_Contract422(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	tagResp, err := client.CreateTagWithResponse(ctx, CreateTagJSONRequestBody{
		Name:  "Stats Tag 422",
		Color: "#123456",
	})
	require.NoError(t, err)
	require.Equal(t, 201, tagResp.StatusCode())

	tagIDs := []TagIdPath{tagResp.JSON201.Id}

	projectResp, err := client.CreateProjectWithResponse(ctx, CreateProjectJSONRequestBody{
		Name:   "Stats Project 422",
		Color:  "#654321",
		TagIds: &tagIDs,
	})
	require.NoError(t, err)
	require.Equal(t, 201, projectResp.StatusCode())
	projectID := projectResp.JSON201.Id

	point := time.Now().UTC().Truncate(time.Second).Add(-2 * time.Hour)
	invalidInterval := point.Format(time.RFC3339) + "/" + point.Format(time.RFC3339)
	granularity := "PT1H"
	timezone := "UTC"

	statsResp, err := client.GetProjectStatsWithResponse(ctx, projectID, &GetProjectStatsParams{
		Metric:      TimeSpent,
		Interval:    &invalidInterval,
		Granularity: &granularity,
		Timezone:    &timezone,
	})
	require.NoError(t, err)
	require.Equal(t, 422, statsResp.StatusCode())
}
