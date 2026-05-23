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
	getResp, err := client.GetTagWithResponse(ctx, tagId, nil)
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
	listResp, err := client.ListTagsWithResponse(ctx, nil)
	require.NoError(t, err)
	require.Equal(t, 200, listResp.StatusCode())
	require.GreaterOrEqual(t, len(listResp.JSON200.Data), 1)

	// DELETE
	deleteResp, err := client.DeleteTagWithResponse(ctx, tagId)
	require.NoError(t, err)
	require.Equal(t, 204, deleteResp.StatusCode())

	// VERIFY DELETION
	getRespAfterDelete, err := client.GetTagWithResponse(ctx, tagId, nil)
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

func TestTag_List_Pagination(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	// Test pagination with default parameters (no params passed)
	defaultResp, err := client.ListTagsWithResponse(ctx, nil)
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
	paramsResp, err := client.ListTagsWithResponse(ctx, &ListTagsParams{
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
