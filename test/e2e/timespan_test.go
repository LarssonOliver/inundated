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
		resp, err := client.CreateTagWithResponse(ctx, &CreateTagParams{XXSRFTOKEN: "token"}, CreateTagJSONRequestBody{
			Name:  fmt.Sprintf("Tag %d", i+1),
			Color: "#123456",
		})
		require.NoError(t, err)
		require.Equal(t, 201, resp.StatusCode())

		uuids = append(uuids, resp.JSON201.Id)
	}

	name := "Test Timespan"

	// CREATE
	createResp, err := client.CreateTimespanWithResponse(ctx, &CreateTimespanParams{XXSRFTOKEN: "token"}, CreateTimespanJSONRequestBody{
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
	updateResp, err := client.UpdateTimespanWithResponse(ctx, timespanId, &UpdateTimespanParams{XXSRFTOKEN: "token"}, UpdateTimespanJSONRequestBody{
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
	deleteResp, err := client.DeleteTimespanWithResponse(ctx, timespanId, &DeleteTimespanParams{XXSRFTOKEN: "token"})
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

	resp, err := client.CreateTimespanWithResponse(ctx, &CreateTimespanParams{XXSRFTOKEN: "token"}, CreateTimespanJSONRequestBody{})
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
