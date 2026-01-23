//go:generate go tool oapi-codegen --config oapi-codegen/model.cfg.yaml ../../openapi/inundated.yaml
//go:generate go tool oapi-codegen --config oapi-codegen/server.cfg.yaml ../../openapi/inundated.yaml

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

// HealthCheck implements [StrictServerInterface].
func (s *Server) HealthCheck(ctx context.Context, request HealthCheckRequestObject) (HealthCheckResponseObject, error) {
	return s.handler.HealthCheck(ctx, request)
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

// CreateTimeSpan implements StrictServerInterface.
func (s *Server) CreateTimeSpan(ctx context.Context, request CreateTimeSpanRequestObject) (CreateTimeSpanResponseObject, error) {
	return s.handler.CreateTimeSpan(ctx, request)
}

// DeleteTimeSpan implements StrictServerInterface.
func (s *Server) DeleteTimeSpan(ctx context.Context, request DeleteTimeSpanRequestObject) (DeleteTimeSpanResponseObject, error) {
	return s.handler.DeleteTimeSpan(ctx, request)
}

// GetTimeSpan implements StrictServerInterface.
func (s *Server) GetTimeSpan(ctx context.Context, request GetTimeSpanRequestObject) (GetTimeSpanResponseObject, error) {
	return s.handler.GetTimeSpan(ctx, request)
}

// ListTimeSpans implements StrictServerInterface.
func (s *Server) ListTimeSpans(ctx context.Context, request ListTimeSpansRequestObject) (ListTimeSpansResponseObject, error) {
	return s.handler.ListTimeSpans(ctx, request)
}

// UpdateTimeSpan implements StrictServerInterface.
func (s *Server) UpdateTimeSpan(ctx context.Context, request UpdateTimeSpanRequestObject) (UpdateTimeSpanResponseObject, error) {
	return s.handler.UpdateTimeSpan(ctx, request)
}
