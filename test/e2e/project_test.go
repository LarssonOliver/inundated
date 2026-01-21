package e2e_test

import (
	"context"
	"slices"
	"testing"

	"github.com/google/uuid"
)

func TestProject_CRUD(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	hours := 5.0
	uuids := []uuid.UUID{
		uuid.MustParse("e53c8903-bb06-45cd-80af-74da154e1472"),
		uuid.MustParse("0b2e62f8-f223-4db2-8a2c-d4188d63760d"),
	}

	// CREATE
	createResp, err := client.CreateProjectWithResponse(ctx, CreateProjectJSONRequestBody{
		Name:            "Test Project",
		Color:           "#FF5733",
		TimeBudgetHours: &hours,
		TagIds:          &uuids,
	})
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}
	if createResp.StatusCode() != 201 {
		t.Fatalf("Expected status code 201, got %d", createResp.StatusCode())
	}

	projectId := createResp.JSON201.Id

	// READ
	getResp, err := client.GetProjectWithResponse(ctx, projectId)
	if err != nil {
		t.Fatalf("Failed to get project: %v", err)
	}
	if getResp.StatusCode() != 200 {
		t.Fatalf("Expected status code 200, got %d", getResp.StatusCode())
	}
	if getResp.JSON200.Name != "Test Project" || getResp.JSON200.Color != "#FF5733" || *getResp.JSON200.TimeBudgetHours != hours || len(*getResp.JSON200.TagIds) != len(uuids) {
		t.Fatalf("Project data does not match created data")
	}
	for _, tagId := range *getResp.JSON200.TagIds {
		found := slices.Contains(uuids, tagId)
		if !found {
			t.Fatalf("Tag ID %s not found in created tag IDs", tagId)
		}

	}

	// UPDATE
	updateResp, err := client.UpdateProjectWithResponse(ctx, projectId, UpdateProjectJSONRequestBody{
		Name: ptr("Updated Test Project"),
	})
	if err != nil {
		t.Fatalf("Failed to update project: %v", err)
	}
	if updateResp.StatusCode() != 200 {
		t.Fatalf("Expected status code 200, got %d", updateResp.StatusCode())
	}

	// LIST
	listResp, err := client.ListProjectsWithResponse(ctx)
	if err != nil {
		t.Fatalf("Failed to list projects: %v", err)
	}
	if listResp.StatusCode() != 200 {
		t.Fatalf("Expected status code 200, got %d", listResp.StatusCode())
	}
	if len(*listResp.JSON200) != 1 {
		t.Fatalf("Expected at least one project in list")
	}

	// DELETE
	deleteResp, err := client.DeleteProjectWithResponse(ctx, projectId)
	if err != nil {
		t.Fatalf("Failed to delete project: %v", err)
	}
	if deleteResp.StatusCode() != 204 {
		t.Fatalf("Expected status code 204, got %d", deleteResp.StatusCode())
	}

	// VERIFY DELETION
	getRespAfterDelete, err := client.GetProjectWithResponse(ctx, projectId)
	if err != nil {
		t.Fatalf("Failed to get project after deletion: %v", err)
	}
	if getRespAfterDelete.StatusCode() != 404 {
		t.Fatalf("Expected status code 404 after deletion, got %d", getRespAfterDelete.StatusCode())
	}
}

func TestProject_Create_InvalidInput(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	resp, err := client.CreateProjectWithResponse(ctx, CreateProjectJSONRequestBody{
		Name: "",
	})
	if err != nil {
		t.Fatalf("Failed to create project with invalid input: %v", err)
	}
	if resp.StatusCode() != 400 {
		t.Fatalf("Expected status code 400 for invalid input, got %d", resp.StatusCode())
	}
}
