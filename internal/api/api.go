//go:generate go tool oapi-codegen --config oapi-codegen/model.cfg.yaml ../../openapi/dist/inundated.yaml
//go:generate go tool oapi-codegen --config oapi-codegen/server.cfg.yaml ../../openapi/dist/inundated.yaml

package api

import (
	"context"
)

var _ StrictServerInterface = (*Server)(nil)

type Server struct {
	handler HttpHandler
}

func NewServer(handler HttpHandler) *Server {
	return &Server{
		handler: handler,
	}
}

// GetCurrentUser implements [StrictServerInterface].
func (s *Server) GetCurrentUser(ctx context.Context, request GetCurrentUserRequestObject) (GetCurrentUserResponseObject, error) {
	panic("unimplemented")
}

// UpdateCurrentUser implements [StrictServerInterface].
func (s *Server) UpdateCurrentUser(ctx context.Context, request UpdateCurrentUserRequestObject) (UpdateCurrentUserResponseObject, error) {
	panic("unimplemented")
}

// CreateTag implements StrictServerInterface.
func (s *Server) CreateTag(ctx context.Context, request CreateTagRequestObject) (CreateTagResponseObject, error) {
	return s.handler.CreateTag(ctx, request)
}

// DeleteTag implements StrictServerInterface.
func (s *Server) DeleteTag(ctx context.Context, request DeleteTagRequestObject) (DeleteTagResponseObject, error) {
	return s.handler.DeleteTag(ctx, request)
}

// GetTag implements StrictServerInterface.
func (s *Server) GetTag(ctx context.Context, request GetTagRequestObject) (GetTagResponseObject, error) {
	return s.handler.GetTag(ctx, request)
}

// ListTags implements StrictServerInterface.
func (s *Server) ListTags(ctx context.Context, request ListTagsRequestObject) (ListTagsResponseObject, error) {
	return s.handler.ListTags(ctx, request)
}

// UpdateTag implements StrictServerInterface.
func (s *Server) UpdateTag(ctx context.Context, request UpdateTagRequestObject) (UpdateTagResponseObject, error) {
	return s.handler.UpdateTag(ctx, request)
}

// CreateProject implements StrictServerInterface.
func (s *Server) CreateProject(ctx context.Context, request CreateProjectRequestObject) (CreateProjectResponseObject, error) {
	return s.handler.CreateProject(ctx, request)
}

// DeleteProject implements StrictServerInterface.
func (s *Server) DeleteProject(ctx context.Context, request DeleteProjectRequestObject) (DeleteProjectResponseObject, error) {
	return s.handler.DeleteProject(ctx, request)
}

// GetProject implements StrictServerInterface.
func (s *Server) GetProject(ctx context.Context, request GetProjectRequestObject) (GetProjectResponseObject, error) {
	return s.handler.GetProject(ctx, request)
}

// ListProjects implements StrictServerInterface.
func (s *Server) ListProjects(ctx context.Context, request ListProjectsRequestObject) (ListProjectsResponseObject, error) {
	return s.handler.ListProjects(ctx, request)
}

// UpdateProject implements StrictServerInterface.
func (s *Server) UpdateProject(ctx context.Context, request UpdateProjectRequestObject) (UpdateProjectResponseObject, error) {
	return s.handler.UpdateProject(ctx, request)
}

// GetProjectStats implements StrictServerInterface.
func (s *Server) GetProjectStats(ctx context.Context, request GetProjectStatsRequestObject) (GetProjectStatsResponseObject, error) {
	return s.handler.GetProjectStats(ctx, request)
}

// CreateTimespan implements StrictServerInterface.
func (s *Server) CreateTimespan(ctx context.Context, request CreateTimespanRequestObject) (CreateTimespanResponseObject, error) {
	return s.handler.CreateTimespan(ctx, request)
}

// DeleteTimespan implements StrictServerInterface.
func (s *Server) DeleteTimespan(ctx context.Context, request DeleteTimespanRequestObject) (DeleteTimespanResponseObject, error) {
	return s.handler.DeleteTimespan(ctx, request)
}

// GetTimespan implements StrictServerInterface.
func (s *Server) GetTimespan(ctx context.Context, request GetTimespanRequestObject) (GetTimespanResponseObject, error) {
	return s.handler.GetTimespan(ctx, request)
}

// ListTimespans implements StrictServerInterface.
func (s *Server) ListTimespans(ctx context.Context, request ListTimespansRequestObject) (ListTimespansResponseObject, error) {
	return s.handler.ListTimespans(ctx, request)
}

// UpdateTimespan implements StrictServerInterface.
func (s *Server) UpdateTimespan(ctx context.Context, request UpdateTimespanRequestObject) (UpdateTimespanResponseObject, error) {
	return s.handler.UpdateTimespan(ctx, request)
}
