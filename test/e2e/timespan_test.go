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
	listResp, err := client.ListTimespansWithResponse(ctx)
	require.NoError(t, err)
	require.Equal(t, 200, listResp.StatusCode())
	require.Len(t, *listResp.JSON200, 1)

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

	resp, err := client.CreateTimespanWithResponse(ctx, CreateTimespanJSONRequestBody{
		Name: nil,
	})
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode())
}
