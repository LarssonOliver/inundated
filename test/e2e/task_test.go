package e2e_test

import (
	"context"
	"testing"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/require"
)

func TestTask_CRUD(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	labelResp, err := client.CreateTagWithResponse(ctx, CreateTagJSONRequestBody{Name: "task-crud-label", Color: "#A3BE8C"})
	require.NoError(t, err)
	require.Equal(t, 201, labelResp.StatusCode())
	labelId := labelResp.JSON201.Id

	// CREATE
	due := openapi_types.Date{Time: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)}
	estimate := 1.5
	createResp, err := client.CreateTaskWithResponse(ctx, CreateTaskJSONRequestBody{
		Name:          "Write report",
		TagIds:        &[]openapi_types.UUID{labelId},
		DueDate:       &due,
		EstimateHours: &estimate,
	})
	require.NoError(t, err)
	require.Equal(t, 201, createResp.StatusCode())
	task := *createResp.JSON201
	require.False(t, task.Closed)
	require.NotEmpty(t, task.Rank)
	require.Equal(t, "2026-10-01", task.DueDate.String())
	require.InDelta(t, 1.5, *task.EstimateHours, 1e-9)

	// The task tag is a tag like any other, marked with its task and
	// colored like the task's first regular tag.
	tagResp, err := client.GetTagWithResponse(ctx, task.TagId, nil)
	require.NoError(t, err)
	require.Equal(t, 200, tagResp.StatusCode())
	require.Equal(t, "Write report", tagResp.JSON200.Name)
	require.Equal(t, task.Id, *tagResp.JSON200.TaskId)
	require.Equal(t, "#A3BE8C", tagResp.JSON200.Color)

	// READ
	getResp, err := client.GetTaskWithResponse(ctx, task.Id)
	require.NoError(t, err)
	require.Equal(t, 200, getResp.StatusCode())
	require.Equal(t, task, *getResp.JSON200)

	// UPDATE: rename, clear the due date, close as ignored.
	reason := Ignored
	updateResp, err := client.UpdateTaskWithResponse(ctx, task.Id, UpdateTaskJSONRequestBody{
		Name:         ptr("Write final report"),
		ClearDueDate: new(true),
		CloseReason:  &reason,
	})
	require.NoError(t, err)
	require.Equal(t, 200, updateResp.StatusCode())
	require.Equal(t, "Write final report", updateResp.JSON200.Name)
	require.Nil(t, updateResp.JSON200.DueDate)
	require.True(t, updateResp.JSON200.Closed)
	require.Equal(t, Ignored, *updateResp.JSON200.CloseReason)
	require.NotNil(t, updateResp.JSON200.ClosedAt)

	tagResp, err = client.GetTagWithResponse(ctx, task.TagId, nil)
	require.NoError(t, err)
	require.Equal(t, "Write final report", tagResp.JSON200.Name)
	require.True(t, tagResp.JSON200.Archived)

	// Reopen.
	updateResp, err = client.UpdateTaskWithResponse(ctx, task.Id, UpdateTaskJSONRequestBody{Closed: new(false)})
	require.NoError(t, err)
	require.Equal(t, 200, updateResp.StatusCode())
	require.False(t, updateResp.JSON200.Closed)
	require.Nil(t, updateResp.JSON200.CloseReason)

	// DELETE
	deleteResp, err := client.DeleteTaskWithResponse(ctx, task.Id)
	require.NoError(t, err)
	require.Equal(t, 204, deleteResp.StatusCode())

	getResp, err = client.GetTaskWithResponse(ctx, task.Id)
	require.NoError(t, err)
	require.Equal(t, 404, getResp.StatusCode())
	tagResp, err = client.GetTagWithResponse(ctx, task.TagId, nil)
	require.NoError(t, err)
	require.Equal(t, 404, tagResp.StatusCode())
}

func TestTask_Update_InvalidInput(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	createResp, err := client.CreateTaskWithResponse(ctx, CreateTaskJSONRequestBody{Name: "update-invalid"})
	require.NoError(t, err)
	id := createResp.JSON201.Id

	due := openapi_types.Date{Time: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	reason := Done
	negative := -1.0
	for name, body := range map[string]UpdateTaskJSONRequestBody{
		"due date and clear":   {DueDate: &due, ClearDueDate: new(true)},
		"reopen with a reason": {Closed: new(false), CloseReason: &reason},
		"empty name":           {Name: ptr("")},
		"negative estimate":    {EstimateHours: &negative},
		"task tag as a tag":    {TagIds: &[]openapi_types.UUID{createResp.JSON201.TagId}},
	} {
		t.Run(name, func(t *testing.T) {
			resp, err := client.UpdateTaskWithResponse(ctx, id, body)
			require.NoError(t, err)
			require.Equal(t, 400, resp.StatusCode())
		})
	}

	resp, err := client.UpdateTaskWithResponse(ctx, openapi_types.UUID{1}, UpdateTaskJSONRequestBody{Name: ptr("x")})
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode())
}

func TestTask_SubtasksAndMove(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	create := func(name string, parent *openapi_types.UUID) Task {
		resp, err := client.CreateTaskWithResponse(ctx, CreateTaskJSONRequestBody{Name: name, ParentId: parent})
		require.NoError(t, err)
		require.Equal(t, 201, resp.StatusCode())
		return *resp.JSON201
	}
	children := func(parent openapi_types.UUID) []string {
		resp, err := client.ListTasksWithResponse(ctx, &ListTasksParams{ParentId: &parent, IncludeClosed: new(true)})
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode())
		var out []string
		for _, task := range resp.JSON200.Data {
			out = append(out, task.Name)
		}
		return out
	}

	parent := create("move-parent", nil)
	a := create("move-a", &parent.Id)
	b := create("move-b", &parent.Id)
	c := create("move-c", &parent.Id)
	require.Equal(t, []string{"move-a", "move-b", "move-c"}, children(parent.Id))

	moveResp, err := client.MoveTaskWithResponse(ctx, c.Id, MoveTaskJSONRequestBody{ParentId: &parent.Id, AfterTaskId: &a.Id})
	require.NoError(t, err)
	require.Equal(t, 200, moveResp.StatusCode())
	require.Equal(t, []string{"move-a", "move-c", "move-b"}, children(parent.Id))

	// Moving a task under its own subtask is refused.
	moveResp, err = client.MoveTaskWithResponse(ctx, parent.Id, MoveTaskJSONRequestBody{ParentId: &a.Id})
	require.NoError(t, err)
	require.Equal(t, 400, moveResp.StatusCode())

	// Closing the parent closes its subtasks, and they drop out of the
	// default list.
	reason := Done
	updateResp, err := client.UpdateTaskWithResponse(ctx, parent.Id, UpdateTaskJSONRequestBody{CloseReason: &reason})
	require.NoError(t, err)
	require.Equal(t, 200, updateResp.StatusCode())
	getResp, err := client.GetTaskWithResponse(ctx, b.Id)
	require.NoError(t, err)
	require.True(t, getResp.JSON200.Closed)

	openResp, err := client.ListTasksWithResponse(ctx, &ListTasksParams{ParentId: &parent.Id})
	require.NoError(t, err)
	require.Empty(t, openResp.JSON200.Data)

	// Deleting the parent takes its subtasks with it.
	deleteResp, err := client.DeleteTaskWithResponse(ctx, parent.Id)
	require.NoError(t, err)
	require.Equal(t, 204, deleteResp.StatusCode())
	getResp, err = client.GetTaskWithResponse(ctx, a.Id)
	require.NoError(t, err)
	require.Equal(t, 404, getResp.StatusCode())
}

func TestTask_DeleteWithLoggedTimeIsRefused(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	createResp, err := client.CreateTaskWithResponse(ctx, CreateTaskJSONRequestBody{Name: "logged"})
	require.NoError(t, err)
	task := *createResp.JSON201

	start := time.Date(2026, 9, 1, 9, 0, 0, 0, time.UTC)
	spanResp, err := client.CreateTimespanWithResponse(ctx, CreateTimespanJSONRequestBody{
		StartTime: start,
		EndTime:   start.Add(time.Hour),
		TagIds:    &[]openapi_types.UUID{task.TagId},
	})
	require.NoError(t, err)
	require.Equal(t, 201, spanResp.StatusCode())

	deleteResp, err := client.DeleteTaskWithResponse(ctx, task.Id)
	require.NoError(t, err)
	require.Equal(t, 409, deleteResp.StatusCode())

	getResp, err := client.GetTaskWithResponse(ctx, task.Id)
	require.NoError(t, err)
	require.Equal(t, 200, getResp.StatusCode())
}

func TestTag_TaskTagsInList(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	createResp, err := client.CreateTaskWithResponse(ctx, CreateTaskJSONRequestBody{Name: "Zebra crossing survey"})
	require.NoError(t, err)
	taskTagId := createResp.JSON201.TagId

	listed := func(params *ListTagsParams) []openapi_types.UUID {
		resp, err := client.ListTagsWithResponse(ctx, params)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode())
		var ids []openapi_types.UUID
		for _, tag := range resp.JSON200.Data {
			ids = append(ids, tag.Id)
		}
		return ids
	}
	kind := func(k ListTagsParamsKind) *ListTagsParamsKind { return &k }

	// Regular tags only by default.
	require.NotContains(t, listed(&ListTagsParams{Limit: new(100)}), taskTagId)
	require.Contains(t, listed(&ListTagsParams{Limit: new(100), Kind: kind(ListTagsParamsKindAll)}), taskTagId)
	require.Equal(t, []openapi_types.UUID{taskTagId}, listed(&ListTagsParams{Kind: kind(ListTagsParamsKindTask), Q: ptr("zebra CROSSING")}))

	resp, err := client.ListTagsWithResponse(ctx, &ListTagsParams{Kind: kind("everything")})
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode())

	// Task tags follow their task and can't be changed directly.
	updateResp, err := client.UpdateTagWithResponse(ctx, taskTagId, UpdateTagJSONRequestBody{Name: ptr("renamed")})
	require.NoError(t, err)
	require.Equal(t, 400, updateResp.StatusCode())
	deleteResp, err := client.DeleteTagWithResponse(ctx, taskTagId)
	require.NoError(t, err)
	require.Equal(t, 400, deleteResp.StatusCode())
}

func TestTask_SurvivesDeletedTag(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	labelResp, err := client.CreateTagWithResponse(ctx, CreateTagJSONRequestBody{Name: "task-deleted-label", Color: "#A3BE8C"})
	require.NoError(t, err)
	require.Equal(t, 201, labelResp.StatusCode())
	labelId := labelResp.JSON201.Id

	createResp, err := client.CreateTaskWithResponse(ctx, CreateTaskJSONRequestBody{
		Name:   "Tagged then untagged",
		TagIds: &[]openapi_types.UUID{labelId},
	})
	require.NoError(t, err)
	require.Equal(t, 201, createResp.StatusCode())
	task := createResp.JSON201

	deleteResp, err := client.DeleteTagWithResponse(ctx, labelId)
	require.NoError(t, err)
	require.Equal(t, 204, deleteResp.StatusCode())

	getResp, err := client.GetTaskWithResponse(ctx, task.Id)
	require.NoError(t, err)
	require.Equal(t, 200, getResp.StatusCode())
	require.Nil(t, getResp.JSON200.TagIds)

	// A partial update that leaves tagIds alone must not trip over the
	// deleted tag.
	updateResp, err := client.UpdateTaskWithResponse(ctx, task.Id, UpdateTaskJSONRequestBody{Closed: new(true)})
	require.NoError(t, err)
	require.Equal(t, 200, updateResp.StatusCode())
	require.True(t, updateResp.JSON200.Closed)
}

func TestTask_DeleteNilIdIsNotFound(t *testing.T) {
	resp, err := newClient().DeleteTaskWithResponse(context.Background(), openapi_types.UUID{})
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode())
}
