package e2e_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTag_CRUD(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	// CREATE
	createResp, err := client.CreateTagWithResponse(ctx, CreateTagJSONRequestBody{
		Name:  "Test Tag",
		Color: "#FF5733",
	})
	require.NoError(t, err)
	require.Equal(t, 201, createResp.StatusCode())

	tagId := createResp.JSON201.Id

	// READ
	getResp, err := client.GetTagWithResponse(ctx, tagId)
	require.NoError(t, err)
	require.Equal(t, 200, getResp.StatusCode())
	require.Equal(t, "Test Tag", getResp.JSON200.Name)
	require.Equal(t, "#FF5733", getResp.JSON200.Color)
	require.Equal(t, tagId, getResp.JSON200.Id)

	// UPDATE
	updateResp, err := client.UpdateTagWithResponse(ctx, tagId, UpdateTagJSONRequestBody{
		Name: ptr("Updated Test Tag"),
	})
	require.NoError(t, err)
	require.Equal(t, 200, updateResp.StatusCode())

	// LIST
	listResp, err := client.ListTagsWithResponse(ctx)
	require.NoError(t, err)
	require.Equal(t, 200, listResp.StatusCode())
	require.GreaterOrEqual(t, len(*listResp.JSON200), 1)

	// DELETE
	deleteResp, err := client.DeleteTagWithResponse(ctx, tagId)
	require.NoError(t, err)
	require.Equal(t, 204, deleteResp.StatusCode())

	// VERIFY DELETION
	getRespAfterDelete, err := client.GetTagWithResponse(ctx, tagId)
	require.NoError(t, err)
	require.Equal(t, 404, getRespAfterDelete.StatusCode())
}

func TestTag_Create_InvalidInput(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	resp, err := client.CreateTagWithResponse(ctx, CreateTagJSONRequestBody{
		Name: "",
	})
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode())
}
