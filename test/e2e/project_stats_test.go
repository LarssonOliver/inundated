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

	tagResp, err := client.CreateTagWithResponse(ctx, &CreateTagParams{XXSRFTOKEN: "token"}, CreateTagJSONRequestBody{
		Name:  "Stats Tag",
		Color: "#123456",
	})
	require.NoError(t, err)
	require.Equal(t, 201, tagResp.StatusCode())

	tagIDs := []TagIdPath{tagResp.JSON201.Id}

	projectResp, err := client.CreateProjectWithResponse(ctx, &CreateProjectParams{XXSRFTOKEN: "token"}, CreateProjectJSONRequestBody{
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

	timespanResp, err := client.CreateTimespanWithResponse(ctx, &CreateTimespanParams{XXSRFTOKEN: "token"}, CreateTimespanJSONRequestBody{
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

func TestProject_GetStats_Semantics(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	createTag := func(name, color string) TagIdPath {
		resp, err := client.CreateTagWithResponse(ctx, &CreateTagParams{XXSRFTOKEN: "token"}, CreateTagJSONRequestBody{
			Name:  name,
			Color: color,
		})
		require.NoError(t, err)
		require.Equal(t, 201, resp.StatusCode())

		return resp.JSON201.Id
	}

	tagA := createTag("Stats Semantic Tag A", "#aa1111")
	tagB := createTag("Stats Semantic Tag B", "#11aa11")
	tagC := createTag("Stats Semantic Tag C", "#1111aa")

	projectTags := []TagIdPath{tagA, tagB}
	projectResp, err := client.CreateProjectWithResponse(ctx, &CreateProjectParams{XXSRFTOKEN: "token"}, CreateProjectJSONRequestBody{
		Name:   "Stats Semantic Project",
		Color:  "#444444",
		TagIds: &projectTags,
	})
	require.NoError(t, err)
	require.Equal(t, 201, projectResp.StatusCode())
	projectID := projectResp.JSON201.Id

	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	name1 := "semantic-ts1"
	_, err = client.CreateTimespanWithResponse(ctx, &CreateTimespanParams{XXSRFTOKEN: "token"}, CreateTimespanJSONRequestBody{
		Name:      &name1,
		StartTime: base.Add(15 * time.Minute),
		EndTime:   base.Add(75 * time.Minute),
		TagIds:    &[]TagIdPath{tagA, tagB},
	})
	require.NoError(t, err)

	name2 := "semantic-ts2"
	_, err = client.CreateTimespanWithResponse(ctx, &CreateTimespanParams{XXSRFTOKEN: "token"}, CreateTimespanJSONRequestBody{
		Name:      &name2,
		StartTime: base.Add(60 * time.Minute),
		EndTime:   base.Add(120 * time.Minute),
		TagIds:    &[]TagIdPath{tagB},
	})
	require.NoError(t, err)

	name3 := "semantic-ts3-ignored"
	_, err = client.CreateTimespanWithResponse(ctx, &CreateTimespanParams{XXSRFTOKEN: "token"}, CreateTimespanJSONRequestBody{
		Name:      &name3,
		StartTime: base,
		EndTime:   base.Add(120 * time.Minute),
		TagIds:    &[]TagIdPath{tagC},
	})
	require.NoError(t, err)

	interval := base.Format(time.RFC3339) + "/" + base.Add(3*time.Hour).Format(time.RFC3339)
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
	require.Len(t, stats.Series, 3)

	expectedIntervals := []string{
		base.Format(time.RFC3339) + "/" + base.Add(1*time.Hour).Format(time.RFC3339),
		base.Add(1*time.Hour).Format(time.RFC3339) + "/" + base.Add(2*time.Hour).Format(time.RFC3339),
		base.Add(2*time.Hour).Format(time.RFC3339) + "/" + base.Add(3*time.Hour).Format(time.RFC3339),
	}

	expectedValues := []float64{
		45 * 60, // split overlap from ts1
		75 * 60, // ts1 tail + ts2, with ts1 deduped despite two matching tags
		0,       // no overlapping timespans in third bucket
	}

	for i := range stats.Series {
		require.Equal(t, expectedIntervals[i], stats.Series[i].Interval)
		require.InDelta(t, expectedValues[i], float64(stats.Series[i].Value), 0.0001)
	}
}

func TestProject_GetStats_Contract422(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	tagResp, err := client.CreateTagWithResponse(ctx, &CreateTagParams{XXSRFTOKEN: "token"}, CreateTagJSONRequestBody{
		Name:  "Stats Tag 422",
		Color: "#123456",
	})
	require.NoError(t, err)
	require.Equal(t, 201, tagResp.StatusCode())

	tagIDs := []TagIdPath{tagResp.JSON201.Id}

	projectResp, err := client.CreateProjectWithResponse(ctx, &CreateProjectParams{XXSRFTOKEN: "token"}, CreateProjectJSONRequestBody{
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
