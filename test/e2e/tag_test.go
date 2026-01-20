package e2e_test

import (
	"context"
	"testing"
)

func TestTag_CRUD(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	// CREATE
	createResp, err := client.CreateTagWithResponse(ctx, CreateTagJSONRequestBody{
		Name:  "Test Tag",
		Color: "#FF5733",
	})
	if err != nil {
		t.Fatalf("Failed to create tag: %v", err)
	}
	if createResp.StatusCode() != 201 {
		t.Fatalf("Expected status code 201, got %d", createResp.StatusCode())
	}

	tagId := createResp.JSON201.Id

	// READ
	getResp, err := client.GetTagWithResponse(ctx, tagId)
	if err != nil {
		t.Fatalf("Failed to get tag: %v", err)
	}
	if getResp.StatusCode() != 200 {
		t.Fatalf("Expected status code 200, got %d", getResp.StatusCode())
	}
	if getResp.JSON200.Name != "Test Tag" || getResp.JSON200.Color != "#FF5733" {
		t.Fatalf("Tag data does not match created data")
	}

	// UPDATE
	updateResp, err := client.UpdateTagWithResponse(ctx, tagId, UpdateTagJSONRequestBody{
		Name: ptr("Updated Test Tag"),
	})
	if err != nil {
		t.Fatalf("Failed to update tag: %v", err)
	}
	if updateResp.StatusCode() != 200 {
		t.Fatalf("Expected status code 200, got %d", updateResp.StatusCode())
	}

	// LIST
	listResp, err := client.ListTagsWithResponse(ctx)
	if err != nil {
		t.Fatalf("Failed to list tags: %v", err)
	}
	if listResp.StatusCode() != 200 {
		t.Fatalf("Expected status code 200, got %d", listResp.StatusCode())
	}
	if len(*listResp.JSON200) != 1 {
		t.Fatalf("Expected at least one tag in list")
	}

	// DELETE
	deleteResp, err := client.DeleteTagWithResponse(ctx, tagId)
	if err != nil {
		t.Fatalf("Failed to delete tag: %v", err)
	}
	if deleteResp.StatusCode() != 204 {
		t.Fatalf("Expected status code 204, got %d", deleteResp.StatusCode())
	}

	// VERIFY DELETION
	getRespAfterDelete, err := client.GetTagWithResponse(ctx, tagId)
	if err != nil {
		t.Fatalf("Failed to get tag after deletion: %v", err)
	}
	if getRespAfterDelete.StatusCode() != 404 {
		t.Fatalf("Expected status code 404 after deletion, got %d", getRespAfterDelete.StatusCode())
	}
}

func TestTag_Create_InvalidInput(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	resp, err := client.CreateTagWithResponse(ctx, CreateTagJSONRequestBody{
		Name: "",
	})
	if err != nil {
		t.Fatalf("Failed to create tag with invalid input: %v", err)
	}
	if resp.StatusCode() != 400 {
		t.Fatalf("Expected status code 400 for invalid input, got %d", resp.StatusCode())
	}
}
