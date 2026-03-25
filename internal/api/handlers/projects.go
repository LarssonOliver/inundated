package handlers

import (
	"context"
	"slices"

	"github.com/larssonoliver/inundated/internal/api"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/service"
	"github.com/larssonoliver/inundated/internal/utils"
)

type ProjectHandler struct {
	svc service.ProjectService
}

var _ api.ProjectHandler = (*ProjectHandler)(nil)

func NewProjectHandler(svc service.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		svc,
	}
}

// CreateProject implements [api.ProjectHandler].
func (p *ProjectHandler) CreateProject(ctx context.Context, request api.CreateProjectRequestObject) (api.CreateProjectResponseObject, error) {
	project := model.Project{
		Name:  request.Body.Name,
		Color: request.Body.Color,
	}

	if request.Body.TimeBudgetHours != nil {
		project.TimeBudget = utils.FloatHoursToDuration(request.Body.TimeBudgetHours)
	}

	if request.Body.TagIds != nil {
		project.TagIds = *request.Body.TagIds
	}

	reply, err := p.svc.CreateProject(ctx, project)

	if err == model.ErrInvalidArgument {
		return api.CreateProject400Response{}, nil
	} else if err != nil {
		return nil, err
	}

	apiProject := api.Project{
		Id:              reply.Id,
		Name:            reply.Name,
		Color:           reply.Color,
		TimeBudgetHours: utils.DurationToFloatHours(reply.TimeBudget),
	}

	if len(reply.TagIds) > 0 {
		apiProject.TagIds = &reply.TagIds
	}

	return api.CreateProject201JSONResponse(apiProject), nil
}

// DeleteProject implements [api.ProjectHandler].
func (p *ProjectHandler) DeleteProject(ctx context.Context, request api.DeleteProjectRequestObject) (api.DeleteProjectResponseObject, error) {
	err := p.svc.DeleteProject(ctx, request.ProjectId)

	if err == model.ErrNotFound {
		return api.DeleteProject404Response{}, nil
	} else if err != nil {
		return nil, err
	}

	return api.DeleteProject204Response{}, nil
}

// GetProject implements [api.ProjectHandler].
func (p *ProjectHandler) GetProject(ctx context.Context, request api.GetProjectRequestObject) (api.GetProjectResponseObject, error) {
	includes := service.ProjectServiceGetIncludes{}

	if request.Params.Include != nil {
		includes.TotalTime = slices.Contains(*request.Params.Include, string(api.GetProjectParamsIncludeTotalTimeMs))
	}

	reply, err := p.svc.GetProject(ctx, request.ProjectId, &includes)

	if err == model.ErrNotFound {
		return api.GetProject404Response{}, nil
	} else if err != nil {
		return nil, err
	}

	apiProject := api.Project{
		Id:              reply.Id,
		Name:            reply.Name,
		Color:           reply.Color,
		TimeBudgetHours: utils.DurationToFloatHours(reply.TimeBudget),
	}

	if includes.TotalTime && reply.TotalTime != nil {
		TotalTimeMs := int(reply.TotalTime.Milliseconds())
		apiProject.TotalTimeMs = &TotalTimeMs
	}

	if len(reply.TagIds) > 0 {
		apiProject.TagIds = &reply.TagIds
	}

	return api.GetProject200JSONResponse(apiProject), nil
}

// ListProjects implements [api.ProjectHandler].
func (p *ProjectHandler) ListProjects(ctx context.Context, request api.ListProjectsRequestObject) (api.ListProjectsResponseObject, error) {
	reply, err := p.svc.ListProjects(ctx)

	if err != nil {
		return nil, err
	}

	apiProjects := make([]api.Project, 0, len(reply))
	for _, project := range reply {
		apiProject := api.Project{
			Id:              project.Id,
			Name:            project.Name,
			Color:           project.Color,
			TimeBudgetHours: utils.DurationToFloatHours(project.TimeBudget),
		}
		if len(project.TagIds) > 0 {
			apiProject.TagIds = &project.TagIds
		}
		apiProjects = append(apiProjects, apiProject)
	}

	return api.ListProjects200JSONResponse(apiProjects), nil
}

// UpdateProject implements [api.ProjectHandler].
func (p *ProjectHandler) UpdateProject(ctx context.Context, request api.UpdateProjectRequestObject) (api.UpdateProjectResponseObject, error) {
	project, err := p.svc.GetProject(ctx, request.ProjectId, nil)

	if err == model.ErrNotFound {
		return api.UpdateProject404Response{}, nil
	} else if err != nil {
		return nil, err
	}

	if request.Body.Name != nil {
		project.Name = *request.Body.Name
	}

	if request.Body.Color != nil {
		project.Color = *request.Body.Color
	}

	if request.Body.TimeBudgetHours != nil {
		project.TimeBudget = utils.FloatHoursToDuration(request.Body.TimeBudgetHours)
	}

	if request.Body.TagIds != nil {
		project.TagIds = *request.Body.TagIds
	}

	reply, err := p.svc.UpdateProject(ctx, project)

	if err == model.ErrInvalidArgument {
		return api.UpdateProject400Response{}, nil
	} else if err != nil {
		return nil, err
	}

	apiProject := api.Project{
		Id:              reply.Id,
		Name:            reply.Name,
		Color:           reply.Color,
		TimeBudgetHours: utils.DurationToFloatHours(reply.TimeBudget),
	}

	if len(reply.TagIds) > 0 {
		apiProject.TagIds = &reply.TagIds
	}

	return api.UpdateProject200JSONResponse(apiProject), nil
}
