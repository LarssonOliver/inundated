package e2e_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestTimespan_CRUD(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	baseTime := time.Now()
	uuids := []uuid.UUID{}

	for i := range 2 {
		resp, err := client.CreateTagWithResponse(ctx, CreateTagJSONRequestBody{
			Name:  fmt.Sprintf("Tag %d", i+1),
			Color: "#123456",
		})
		require.NoError(t, err)
		require.Equal(t, 201, resp.StatusCode())

		uuids = append(uuids, resp.JSON201.Id)
	}

	name := "Test Timespan"

	// CREATE
	createResp, err := client.CreateTimespanWithResponse(ctx, CreateTimespanJSONRequestBody{
		Name:      &name,
		StartTime: baseTime,
		EndTime:   baseTime.Add(2 * time.Hour),
		TagIds:    &uuids,
	})
	require.NoError(t, err)
	require.Equal(t, 201, createResp.StatusCode())

	timespanId := createResp.JSON201.Id

	// READ
	getResp, err := client.GetTimespanWithResponse(ctx, timespanId)
	require.NoError(t, err)
	require.Equal(t, 200, getResp.StatusCode())

	require.Equal(t, "Test Timespan", *getResp.JSON200.Name)
	require.True(t, getResp.JSON200.StartTime.Equal(baseTime))
	require.True(t, getResp.JSON200.EndTime.Equal(baseTime.Add(2*time.Hour)))
	require.Len(t, *getResp.JSON200.TagIds, len(uuids))

	for _, tagId := range *getResp.JSON200.TagIds {
		require.Contains(t, uuids, tagId)
	}

	// UPDATE
	updateResp, err := client.UpdateTimespanWithResponse(ctx, timespanId, UpdateTimespanJSONRequestBody{
		Name: ptr("Updated Test Timespan"),
	})
	require.NoError(t, err)
	require.Equal(t, 200, updateResp.StatusCode())

	// LIST
	listResp, err := client.ListTimespansWithResponse(ctx, nil)
	require.NoError(t, err)
	require.Equal(t, 200, listResp.StatusCode())
	require.GreaterOrEqual(t, len(listResp.JSON200.Data), 1)

	// DELETE
	deleteResp, err := client.DeleteTimespanWithResponse(ctx, timespanId)
	require.NoError(t, err)
	require.Equal(t, 204, deleteResp.StatusCode())

	// VERIFY DELETION
	getRespAfterDelete, err := client.GetTimespanWithResponse(ctx, timespanId)
	require.NoError(t, err)
	require.Equal(t, 404, getRespAfterDelete.StatusCode())
}

func TestTimespan_Create_InvalidInput(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	resp, err := client.CreateTimespanWithResponse(ctx, CreateTimespanJSONRequestBody{})
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode())
}

func TestTimespan_List_Pagination(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	// Test pagination with default parameters (no params passed)
	defaultResp, err := client.ListTimespansWithResponse(ctx, nil)
	require.NoError(t, err)
	require.Equal(t, 200, defaultResp.StatusCode())
	require.NotNil(t, defaultResp.JSON200)

	// Verify response structure
	resp := defaultResp.JSON200
	require.NotNil(t, resp.Data)
	require.NotNil(t, resp.Pagination)

	// Verify pagination metadata exists
	require.GreaterOrEqual(t, resp.Pagination.Limit, 1)
	require.GreaterOrEqual(t, resp.Pagination.Offset, 0)
	require.GreaterOrEqual(t, resp.Pagination.Total, 0)

	// Test pagination with explicit limit and offset parameters
	limit := Limit(10)
	offset := Offset(0)
	paramsResp, err := client.ListTimespansWithResponse(ctx, &ListTimespansParams{
		Limit:  &limit,
		Offset: &offset,
	})
	require.NoError(t, err)
	require.Equal(t, 200, paramsResp.StatusCode())
	require.NotNil(t, paramsResp.JSON200)

	// Verify pagination parameters are respected
	paramsData := paramsResp.JSON200
	require.Equal(t, int(limit), paramsData.Pagination.Limit)
	require.Equal(t, int(offset), paramsData.Pagination.Offset)
	require.LessOrEqual(t, len(paramsData.Data), int(limit))
}

func TestTimespan_List_IntervalFilter(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	// Far in the past so this window can't collide with timespans other
	// tests create around time.Now().
	base := time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)
	name := "Interval filter test"

	createResp, err := client.CreateTimespanWithResponse(ctx, CreateTimespanJSONRequestBody{
		Name:      &name,
		StartTime: base,
		EndTime:   base.Add(time.Hour),
	})
	require.NoError(t, err)
	require.Equal(t, 201, createResp.StatusCode())

	overlapping := Interval(fmt.Sprintf("%s/%s", base.Add(-time.Hour).Format(time.RFC3339), base.Add(2*time.Hour).Format(time.RFC3339)))
	overlapResp, err := client.ListTimespansWithResponse(ctx, &ListTimespansParams{Interval: &overlapping})
	require.NoError(t, err)
	require.Equal(t, 200, overlapResp.StatusCode())

	found := false
	for _, ts := range overlapResp.JSON200.Data {
		if ts.Id == createResp.JSON201.Id {
			found = true
		}
	}
	require.True(t, found, "timespan should be returned by an overlapping interval")

	nonOverlapping := Interval(fmt.Sprintf("%s/%s", base.Add(-24*time.Hour).Format(time.RFC3339), base.Add(-2*time.Hour).Format(time.RFC3339)))
	nonOverlapResp, err := client.ListTimespansWithResponse(ctx, &ListTimespansParams{Interval: &nonOverlapping})
	require.NoError(t, err)
	require.Equal(t, 200, nonOverlapResp.StatusCode())

	for _, ts := range nonOverlapResp.JSON200.Data {
		require.NotEqual(t, createResp.JSON201.Id, ts.Id, "timespan should not be returned by a non-overlapping interval")
	}

	invalid := Interval(fmt.Sprintf("%s/%s", base.Add(time.Hour).Format(time.RFC3339), base.Format(time.RFC3339)))
	invalidResp, err := client.ListTimespansWithResponse(ctx, &ListTimespansParams{Interval: &invalid})
	require.NoError(t, err)
	require.Equal(t, 422, invalidResp.StatusCode())
}
