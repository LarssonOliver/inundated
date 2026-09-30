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
	require.Equal(t, &TagOwner{Kind: TagOwnerKindTask, Id: task.Id}, tagResp.JSON200.Owner)
	require.Equal(t, "#A3BE8C", tagResp.JSON200.Color)

	// READ
	getResp, err := client.GetTaskWithResponse(ctx, task.Id, nil)
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

	getResp, err = client.GetTaskWithResponse(ctx, task.Id, nil)
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
	getResp, err := client.GetTaskWithResponse(ctx, b.Id, nil)
	require.NoError(t, err)
	require.True(t, getResp.JSON200.Closed)

	openResp, err := client.ListTasksWithResponse(ctx, &ListTasksParams{ParentId: &parent.Id})
	require.NoError(t, err)
	require.Empty(t, openResp.JSON200.Data)

	// Deleting the parent takes its subtasks with it.
	deleteResp, err := client.DeleteTaskWithResponse(ctx, parent.Id)
	require.NoError(t, err)
	require.Equal(t, 204, deleteResp.StatusCode())
	getResp, err = client.GetTaskWithResponse(ctx, a.Id, nil)
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

	getResp, err := client.GetTaskWithResponse(ctx, task.Id, nil)
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
	kind := func(kinds ...ListTagsParamsKind) *TagKindQuery {
		q := TagKindQuery{}
		for _, k := range kinds {
			q = append(q, string(k))
		}
		return &q
	}

	// Regular tags only by default.
	require.NotContains(t, listed(&ListTagsParams{Limit: new(100)}), taskTagId)
	require.Contains(t, listed(&ListTagsParams{Limit: new(100), Kind: kind(ListTagsParamsKindAll)}), taskTagId)
	require.Equal(t, []openapi_types.UUID{taskTagId}, listed(&ListTagsParams{Kind: kind(ListTagsParamsKindTask), Q: ptr("zebra CROSSING")}))

	// Several kinds at once.
	labelResp, err := client.CreateTagWithResponse(ctx, CreateTagJSONRequestBody{Name: "Zebra stripes", Color: "#88c0d0"})
	require.NoError(t, err)
	require.ElementsMatch(t, []openapi_types.UUID{labelResp.JSON201.Id, taskTagId},
		listed(&ListTagsParams{Kind: kind(ListTagsParamsKindTask, ListTagsParamsKindLabel), Q: ptr("zebra")}))

	resp, err := client.ListTagsWithResponse(ctx, &ListTagsParams{Kind: kind(ListTagsParamsKindLabel, "everything")})
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

	getResp, err := client.GetTaskWithResponse(ctx, task.Id, nil)
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

func TestTask_NilIdWritesAreNotFound(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	updateResp, err := client.UpdateTaskWithResponse(ctx, openapi_types.UUID{}, UpdateTaskJSONRequestBody{Name: new("x")})
	require.NoError(t, err)
	require.Equal(t, 404, updateResp.StatusCode())

	moveResp, err := client.MoveTaskWithResponse(ctx, openapi_types.UUID{}, MoveTaskJSONRequestBody{})
	require.NoError(t, err)
	require.Equal(t, 404, moveResp.StatusCode())

	tagResp, err := client.DeleteTagWithResponse(ctx, openapi_types.UUID{})
	require.NoError(t, err)
	require.Equal(t, 404, tagResp.StatusCode())
}

func TestTask_HugeEstimateIsRejected(t *testing.T) {
	ctx := context.Background()
	client := newClient()
	huge := 1e7

	createResp, err := client.CreateTaskWithResponse(ctx, CreateTaskJSONRequestBody{Name: "huge", EstimateHours: &huge})
	require.NoError(t, err)
	require.Equal(t, 400, createResp.StatusCode())

	createResp, err = client.CreateTaskWithResponse(ctx, CreateTaskJSONRequestBody{Name: "fine"})
	require.NoError(t, err)
	require.Equal(t, 201, createResp.StatusCode())
	updateResp, err := client.UpdateTaskWithResponse(ctx, createResp.JSON201.Id, UpdateTaskJSONRequestBody{EstimateHours: &huge})
	require.NoError(t, err)
	require.Equal(t, 400, updateResp.StatusCode())
}

func TestTask_TimeAndProjects(t *testing.T) {
	ctx := context.Background()
	client := newClient()

	parentResp, err := client.CreateTaskWithResponse(ctx, CreateTaskJSONRequestBody{Name: "attribution-parent"})
	require.NoError(t, err)
	require.Equal(t, 201, parentResp.StatusCode())
	parent := *parentResp.JSON201
	childResp, err := client.CreateTaskWithResponse(ctx, CreateTaskJSONRequestBody{Name: "attribution-child", ParentId: &parent.Id})
	require.NoError(t, err)
	require.Equal(t, 201, childResp.StatusCode())
	child := *childResp.JSON201

	// Assign the parent to a project and log time on the child.
	projectResp, err := client.CreateProjectWithResponse(ctx, CreateProjectJSONRequestBody{
		Name: "attribution-project", Color: "#5E81AC", TagIds: &[]TagIdPath{parent.TagId},
	})
	require.NoError(t, err)
	require.Equal(t, 201, projectResp.StatusCode())
	projectId := projectResp.JSON201.Id

	start := time.Date(2026, 9, 1, 9, 0, 0, 0, time.UTC)
	spanResp, err := client.CreateTimespanWithResponse(ctx, CreateTimespanJSONRequestBody{
		StartTime: start, EndTime: start.Add(90 * time.Minute), TagIds: &[]TagIdPath{child.TagId},
	})
	require.NoError(t, err)
	require.Equal(t, 201, spanResp.StatusCode())

	include := TaskInclude{"totalTimeMs", "projectIds"}
	getResp, err := client.GetTaskWithResponse(ctx, parent.Id, &GetTaskParams{Include: &include})
	require.NoError(t, err)
	require.Equal(t, 200, getResp.StatusCode())
	require.Equal(t, 90*60*1000, *getResp.JSON200.TotalTimeMs, "subtask time rolls up")
	require.Equal(t, []openapi_types.UUID{projectId}, *getResp.JSON200.ProjectIds)

	// Without include the computed fields are left out.
	plainResp, err := client.GetTaskWithResponse(ctx, parent.Id, nil)
	require.NoError(t, err)
	require.Nil(t, plainResp.JSON200.TotalTimeMs)
	require.Nil(t, plainResp.JSON200.ProjectIds)

	listResp, err := client.ListTasksWithResponse(ctx, &ListTasksParams{ProjectId: &projectId, Include: &include})
	require.NoError(t, err)
	require.Equal(t, 200, listResp.StatusCode())
	require.Len(t, listResp.JSON200.Data, 2)
	for _, task := range listResp.JSON200.Data {
		require.Equal(t, 90*60*1000, *task.TotalTimeMs)
		require.Equal(t, []openapi_types.UUID{projectId}, *task.ProjectIds)
	}

	// The project total counts the subtask's time, all of it on tasks.
	projectInclude := ProjectIncludeQuery{"totalTimeMs", "taskTimeMs"}
	projectGet, err := client.GetProjectWithResponse(ctx, projectId, &GetProjectParams{Include: &projectInclude})
	require.NoError(t, err)
	require.Equal(t, 90*60*1000, *projectGet.JSON200.TotalTimeMs)
	require.Equal(t, 90*60*1000, *projectGet.JSON200.TaskTimeMs)

	// Time logged on the project's regular tag is not on any task.
	tagResp, err := client.CreateTagWithResponse(ctx, CreateTagJSONRequestBody{Name: "attribution-tag", Color: "#88C0D0"})
	require.NoError(t, err)
	require.Equal(t, 201, tagResp.StatusCode())
	tagIds := []TagIdPath{parent.TagId, tagResp.JSON201.Id}
	updateResp, err := client.UpdateProjectWithResponse(ctx, projectId, UpdateProjectJSONRequestBody{TagIds: &tagIds})
	require.NoError(t, err)
	require.Equal(t, 200, updateResp.StatusCode())
	spanResp, err = client.CreateTimespanWithResponse(ctx, CreateTimespanJSONRequestBody{
		StartTime: start.Add(2 * time.Hour), EndTime: start.Add(150 * time.Minute), TagIds: &[]TagIdPath{tagResp.JSON201.Id},
	})
	require.NoError(t, err)
	require.Equal(t, 201, spanResp.StatusCode())

	projectGet, err = client.GetProjectWithResponse(ctx, projectId, &GetProjectParams{Include: &projectInclude})
	require.NoError(t, err)
	require.Equal(t, 120*60*1000, *projectGet.JSON200.TotalTimeMs)
	require.Equal(t, 90*60*1000, *projectGet.JSON200.TaskTimeMs)
}
