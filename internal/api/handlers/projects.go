package handlers

import (
	"context"

	"github.com/larssonoliver/inundated/internal/api"
)

type ProjectHandler struct {
	// svc service.ProjectService
}

var _ api.ProjectHandler = (*ProjectHandler)(nil)

func NewProjectHandler() *ProjectHandler {
	return &ProjectHandler{}
}

// func NewProjectHandler(svc service.ProjectService) *ProjectHandler {
// 	return &ProjectHandler{
// 		svc,
// 	}
// }

// CreateProject implements [api.ProjectHandler].
func (p *ProjectHandler) CreateProject(ctx context.Context, request api.CreateProjectRequestObject) (api.CreateProjectResponseObject, error) {
	panic("unimplemented")
}

// DeleteProject implements [api.ProjectHandler].
func (p *ProjectHandler) DeleteProject(ctx context.Context, request api.DeleteProjectRequestObject) (api.DeleteProjectResponseObject, error) {
	panic("unimplemented")
}

// GetProject implements [api.ProjectHandler].
func (p *ProjectHandler) GetProject(ctx context.Context, request api.GetProjectRequestObject) (api.GetProjectResponseObject, error) {
	panic("unimplemented")
}

// ListProjects implements [api.ProjectHandler].
func (p *ProjectHandler) ListProjects(ctx context.Context, request api.ListProjectsRequestObject) (api.ListProjectsResponseObject, error) {
	panic("unimplemented")
}

// UpdateProject implements [api.ProjectHandler].
func (p *ProjectHandler) UpdateProject(ctx context.Context, request api.UpdateProjectRequestObject) (api.UpdateProjectResponseObject, error) {
	panic("unimplemented")
}
