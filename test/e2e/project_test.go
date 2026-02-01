package e2e_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestProject_CRUD(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	hours := 5.0

	uuids := []uuid.UUID{}
	for i := range 2 {
		resp, err := client.CreateTagWithResponse(ctx, CreateTagJSONRequestBody{
			Name:  fmt.Sprintf("Tag %d", i+1),
			Color: "#123456",
		})
		require.NoError(t, err)
		uuids = append(uuids, resp.JSON201.Id)
	}

	// CREATE
	createResp, err := client.CreateProjectWithResponse(ctx, CreateProjectJSONRequestBody{
		Name:            "Test Project",
		Color:           "#FF5733",
		TimeBudgetHours: &hours,
		TagIds:          &uuids,
	})
	require.NoError(t, err)
	require.Equal(t, 201, createResp.StatusCode())

	projectId := createResp.JSON201.Id

	// READ
	getResp, err := client.GetProjectWithResponse(ctx, projectId)
	require.NoError(t, err)
	require.Equal(t, 200, getResp.StatusCode())
	require.Equal(t, "Test Project", getResp.JSON200.Name)
	require.Equal(t, "#FF5733", getResp.JSON200.Color)
	require.Equal(t, hours, *getResp.JSON200.TimeBudgetHours)
	require.Len(t, *getResp.JSON200.TagIds, len(uuids))
	require.ElementsMatch(t, uuids, *getResp.JSON200.TagIds)

	// UPDATE
	updateResp, err := client.UpdateProjectWithResponse(ctx, projectId, UpdateProjectJSONRequestBody{
		Name: ptr("Updated Test Project"),
	})
	require.NoError(t, err)
	require.Equal(t, 200, updateResp.StatusCode())

	// LIST
	listResp, err := client.ListProjectsWithResponse(ctx)
	require.NoError(t, err)
	require.Equal(t, 200, listResp.StatusCode())
	require.GreaterOrEqual(t, len(*listResp.JSON200), 1)

	// DELETE
	deleteResp, err := client.DeleteProjectWithResponse(ctx, projectId)
	require.NoError(t, err)
	require.Equal(t, 204, deleteResp.StatusCode())

	// VERIFY DELETION
	getRespAfterDelete, err := client.GetProjectWithResponse(ctx, projectId)
	require.NoError(t, err)
	require.Equal(t, 404, getRespAfterDelete.StatusCode())
}

func TestProject_Create_InvalidInput(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	resp, err := client.CreateProjectWithResponse(ctx, CreateProjectJSONRequestBody{
		Name: "",
	})
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode())
}
