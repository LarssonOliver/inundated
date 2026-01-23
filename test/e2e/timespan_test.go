package e2e_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestTimeSpan_CRUD(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	baseTime := time.Now()
	uuids := []uuid.UUID{
		uuid.MustParse("e53c8903-bb06-45cd-80af-74da154e1472"),
		uuid.MustParse("0b2e62f8-f223-4db2-8a2c-d4188d63760d"),
	}

	// CREATE
	createResp, err := client.CreateTimeSpanWithResponse(ctx, CreateTimeSpanJSONRequestBody{
		Name:      "Test TimeSpan",
		StartTime: baseTime,
		EndTime:   baseTime.Add(2 * time.Hour),
		TagIds:    &uuids,
	})
	if err != nil {
		t.Fatalf("Failed to create timespan: %v", err)
	}
	if createResp.StatusCode() != 201 {
		t.Fatalf("Expected status code 201, got %d", createResp.StatusCode())
	}

	timespanId := createResp.JSON201.Id

	// READ
	getResp, err := client.GetTimeSpanWithResponse(ctx, timespanId)
	if err != nil {
		t.Fatalf("Failed to get timespan: %v", err)
	}
	if getResp.StatusCode() != 200 {
		t.Fatalf("Expected status code 200, got %d", getResp.StatusCode())
	}
	if getResp.JSON200.Name != "Test TimeSpan" || !getResp.JSON200.StartTime.Equal(baseTime) || !getResp.JSON200.EndTime.Equal(baseTime.Add(2*time.Hour)) || len(*getResp.JSON200.TagIds) != len(uuids) {
		t.Fatalf("TimeSpan data does not match created data, gotr%+v", getResp.JSON200)
	}
	for _, tagId := range *getResp.JSON200.TagIds {
		found := slices.Contains(uuids, tagId)
		if !found {
			t.Fatalf("Tag ID %s not found in created tag IDs", tagId)
		}
	}

	// UPDATE
	updateResp, err := client.UpdateTimeSpanWithResponse(ctx, timespanId, UpdateTimeSpanJSONRequestBody{
		Name: ptr("Updated Test TimeSpan"),
	})
	if err != nil {
		t.Fatalf("Failed to update timespan: %v", err)
	}
	if updateResp.StatusCode() != 200 {
		t.Fatalf("Expected status code 200, got %d", updateResp.StatusCode())
	}

	// LIST
	listResp, err := client.ListTimeSpansWithResponse(ctx)
	if err != nil {
		t.Fatalf("Failed to list timespans: %v", err)
	}
	if listResp.StatusCode() != 200 {
		t.Fatalf("Expected status code 200, got %d", listResp.StatusCode())
	}
	if len(*listResp.JSON200) != 1 {
		t.Fatalf("Expected at least one timespan in list")
	}

	// DELETE
	deleteResp, err := client.DeleteTimeSpanWithResponse(ctx, timespanId)
	if err != nil {
		t.Fatalf("Failed to delete timespan: %v", err)
	}
	if deleteResp.StatusCode() != 204 {
		t.Fatalf("Expected status code 204, got %d", deleteResp.StatusCode())
	}

	// VERIFY DELETION
	getRespAfterDelete, err := client.GetTimeSpanWithResponse(ctx, timespanId)
	if err != nil {
		t.Fatalf("Failed to get timespan after deletion: %v", err)
	}
	if getRespAfterDelete.StatusCode() != 404 {
		t.Fatalf("Expected status code 404 after deletion, got %d", getRespAfterDelete.StatusCode())
	}
}

func TestTimeSpan_Create_InvalidInput(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	resp, err := client.CreateTimeSpanWithResponse(ctx, CreateTimeSpanJSONRequestBody{
		Name: "",
	})
	if err != nil {
		t.Fatalf("Failed to create timespan with invalid input: %v", err)
	}
	if resp.StatusCode() != 400 {
		t.Fatalf("Expected status code 400 for invalid input, got %d", resp.StatusCode())
	}
}
