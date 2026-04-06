package handlers

import (
	"context"
	"slices"

	"github.com/google/uuid"
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

// GetProjectStats implements [api.ProjectHandler].
func (p *ProjectHandler) GetProjectStats(ctx context.Context, request api.GetProjectStatsRequestObject) (api.GetProjectStatsResponseObject, error) {
	// Extract and validate parameters
	metric := string(request.Params.Metric)

	intervalStr := ""
	if request.Params.Interval != nil {
		intervalStr = string(*request.Params.Interval)
	}

	granularity := "P1D" // default
	if request.Params.Granularity != nil {
		granularity = string(*request.Params.Granularity)
	}

	timezone := "UTC" // default
	if request.Params.Timezone != nil {
		timezone = string(*request.Params.Timezone)
	}

	// Call service
	stats, err := p.svc.GetProjectStats(ctx, request.ProjectId, metric, intervalStr, granularity, timezone)

	// Handle errors
	if err == model.ErrNotFound {
		return api.GetProjectStats404Response{}, nil
	}

	// Check for validation errors (400 or 422)
	if err != nil {
		errMsg := err.Error()
		// Semantic errors (interval validation) return 422
		if containsAny(errMsg, "start must be before end", "after", "before") {
			return api.GetProjectStats422Response{}, nil
		}
		// Format/parse errors return 400
		if containsAny(errMsg, "invalid", "unsupported", "failed to parse") {
			return api.GetProjectStats400Response{}, nil
		}
		// Other errors
		return nil, err
	}

	// Convert to API response
	series := make([]api.SeriesPoint, len(stats.Series))
	for i, point := range stats.Series {
		series[i] = api.SeriesPoint{
			Interval: point.Interval,
			Value:    float32(point.Value),
		}
	}

	projectUUID, err := uuid.Parse(stats.ProjectID)
	if err != nil {
		return nil, err
	}

	response := api.ProjectStats{
		ProjectId:   projectUUID,
		Metric:      api.ProjectStatsMetric(stats.Metric),
		Interval:    stats.Interval,
		Granularity: stats.Granularity,
		Unit:        stats.Unit,
		Series:      series,
	}

	return api.GetProjectStats200JSONResponse(response), nil
}

// containsAny checks if s contains any of the substrings
func containsAny(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if len(substr) > 0 && len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}
