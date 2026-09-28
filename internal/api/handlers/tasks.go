package handlers

import (
	"context"
	"errors"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/larssonoliver/inundated/internal/api"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/service"
	"github.com/larssonoliver/inundated/internal/utils"
)

type TaskHandler struct {
	svc service.TaskService
}

var _ api.TaskHandler = (*TaskHandler)(nil)

func NewTaskHandler(svc service.TaskService) *TaskHandler {
	return &TaskHandler{
		svc,
	}
}

// CreateTask implements [api.TaskHandler].
func (h *TaskHandler) CreateTask(ctx context.Context, request api.CreateTaskRequestObject) (api.CreateTaskResponseObject, error) {
	task := model.Task{
		Name:     request.Body.Name,
		ParentId: request.Body.ParentId,
		DueDate:  dateToTime(request.Body.DueDate),
		Estimate: utils.FloatHoursToDuration(request.Body.EstimateHours),
	}
	if request.Body.TagIds != nil {
		task.TagIds = *request.Body.TagIds
	}

	reply, err := h.svc.CreateTask(ctx, task)

	if errors.Is(err, model.ErrInvalidArgument) || errors.Is(err, model.ErrInvalidReference) {
		return api.CreateTask400Response{}, nil
	} else if err != nil {
		return nil, err
	}

	return api.CreateTask201JSONResponse(toAPITask(reply)), nil
}

// DeleteTask implements [api.TaskHandler].
func (h *TaskHandler) DeleteTask(ctx context.Context, request api.DeleteTaskRequestObject) (api.DeleteTaskResponseObject, error) {
	err := h.svc.DeleteTask(ctx, request.TaskId)

	if errors.Is(err, model.ErrNotFound) {
		return api.DeleteTask404Response{}, nil
	} else if errors.Is(err, model.ErrConflict) {
		return api.DeleteTask409Response{}, nil
	} else if err != nil {
		return nil, err
	}

	return api.DeleteTask204Response{}, nil
}

// GetTask implements [api.TaskHandler].
func (h *TaskHandler) GetTask(ctx context.Context, request api.GetTaskRequestObject) (api.GetTaskResponseObject, error) {
	reply, err := h.svc.GetTask(ctx, request.TaskId)

	if errors.Is(err, model.ErrNotFound) {
		return api.GetTask404Response{}, nil
	} else if err != nil {
		return nil, err
	}

	return api.GetTask200JSONResponse(toAPITask(reply)), nil
}

// ListTasks implements [api.TaskHandler].
func (h *TaskHandler) ListTasks(ctx context.Context, request api.ListTasksRequestObject) (api.ListTasksResponseObject, error) {
	params := model.TaskListParams{PaginationParams: model.DefaultPaginationParams()}

	if request.Params.Limit != nil {
		if *request.Params.Limit < 1 || *request.Params.Limit > 100 {
			return api.ListTasks400Response{}, nil
		}
		params.Limit = *request.Params.Limit
	}

	if request.Params.Offset != nil {
		if *request.Params.Offset < 0 {
			return api.ListTasks400Response{}, nil
		}
		params.Offset = *request.Params.Offset
	}

	if request.Params.IncludeClosed != nil {
		params.IncludeClosed = *request.Params.IncludeClosed
	}
	params.ParentId = request.Params.ParentId
	params.TagId = request.Params.TagId
	params.DueFrom = dateToTime(request.Params.DueFrom)
	params.DueTo = dateToTime(request.Params.DueTo)

	page, err := h.svc.ListTasks(ctx, params)

	if errors.Is(err, model.ErrInvalidArgument) {
		return api.ListTasks400Response{}, nil
	} else if err != nil {
		return nil, err
	}

	apiTasks := make([]api.Task, 0, len(page.Data))
	for _, task := range page.Data {
		apiTasks = append(apiTasks, toAPITask(task))
	}

	return api.ListTasks200JSONResponse(api.PaginatedTasks{
		Data: apiTasks,
		Pagination: api.PaginationDetails{
			Limit:  page.Limit,
			Offset: page.Offset,
			Total:  page.TotalCount,
		},
	}), nil
}

// UpdateTask implements [api.TaskHandler].
func (h *TaskHandler) UpdateTask(ctx context.Context, request api.UpdateTaskRequestObject) (api.UpdateTaskResponseObject, error) {
	body := request.Body
	patch := model.TaskPatch{
		Name:          body.Name,
		TagIds:        body.TagIds,
		DueDate:       dateToTime(body.DueDate),
		ClearDueDate:  body.ClearDueDate != nil && *body.ClearDueDate,
		Estimate:      utils.FloatHoursToDuration(body.EstimateHours),
		ClearEstimate: body.ClearEstimate != nil && *body.ClearEstimate,
		Closed:        body.Closed,
	}
	if body.CloseReason != nil {
		reason := model.CloseReason(*body.CloseReason)
		patch.CloseReason = &reason
	}

	reply, err := h.svc.UpdateTask(ctx, request.TaskId, patch)

	if errors.Is(err, model.ErrInvalidArgument) || errors.Is(err, model.ErrInvalidReference) {
		return api.UpdateTask400Response{}, nil
	} else if errors.Is(err, model.ErrNotFound) {
		return api.UpdateTask404Response{}, nil
	} else if err != nil {
		return nil, err
	}

	return api.UpdateTask200JSONResponse(toAPITask(reply)), nil
}

// MoveTask implements [api.TaskHandler].
func (h *TaskHandler) MoveTask(ctx context.Context, request api.MoveTaskRequestObject) (api.MoveTaskResponseObject, error) {
	reply, err := h.svc.MoveTask(ctx, request.TaskId, request.Body.ParentId, request.Body.AfterTaskId)

	if errors.Is(err, model.ErrInvalidArgument) {
		return api.MoveTask400Response{}, nil
	} else if errors.Is(err, model.ErrNotFound) {
		return api.MoveTask404Response{}, nil
	} else if err != nil {
		return nil, err
	}

	return api.MoveTask200JSONResponse(toAPITask(reply)), nil
}

func toAPITask(task model.Task) api.Task {
	apiTask := api.Task{
		Id:            task.Id,
		Name:          task.Name,
		TagId:         task.TagId,
		ParentId:      task.ParentId,
		EstimateHours: utils.DurationToFloatHours(task.Estimate),
		Rank:          task.Rank,
		Closed:        task.Closed(),
		ClosedAt:      task.ClosedAt,
	}
	if len(task.TagIds) > 0 {
		apiTask.TagIds = &task.TagIds
	}
	if task.DueDate != nil {
		apiTask.DueDate = &openapi_types.Date{Time: *task.DueDate}
	}
	if task.CloseReason != nil {
		reason := api.CloseReason(*task.CloseReason)
		apiTask.CloseReason = &reason
	}
	return apiTask
}

func dateToTime(d *openapi_types.Date) *time.Time {
	if d == nil {
		return nil
	}
	t := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
	return &t
}
