//go:generate go tool oapi-codegen --config oapi-codegen/model.cfg.yaml ../../openapi/inundated.yaml
//go:generate go tool oapi-codegen --config oapi-codegen/server.cfg.yaml ../../openapi/inundated.yaml

package api

import "context"

var _ StrictServerInterface = (*Server)(nil)

type Server struct {
	tags      TagHandler
	projects  ProjectHandler
	timeSpans TimeSpanHandler
}

func NewServer(tags TagHandler, projects ProjectHandler, timeSpans TimeSpanHandler) *Server {
	return &Server{
		tags,
		projects,
		timeSpans,
	}
}

// CreateProject implements StrictServerInterface.
func (s *Server) CreateProject(ctx context.Context, request CreateProjectRequestObject) (CreateProjectResponseObject, error) {
	return s.projects.CreateProject(ctx, request)
}

// CreateTag implements StrictServerInterface.
func (s *Server) CreateTag(ctx context.Context, request CreateTagRequestObject) (CreateTagResponseObject, error) {
	return s.tags.CreateTag(ctx, request)
}

// CreateTimeSpan implements StrictServerInterface.
func (s *Server) CreateTimeSpan(ctx context.Context, request CreateTimeSpanRequestObject) (CreateTimeSpanResponseObject, error) {
	return s.timeSpans.CreateTimeSpan(ctx, request)
}

// DeleteProject implements StrictServerInterface.
func (s *Server) DeleteProject(ctx context.Context, request DeleteProjectRequestObject) (DeleteProjectResponseObject, error) {
	return s.projects.DeleteProject(ctx, request)
}

// DeleteTag implements StrictServerInterface.
func (s *Server) DeleteTag(ctx context.Context, request DeleteTagRequestObject) (DeleteTagResponseObject, error) {
	return s.tags.DeleteTag(ctx, request)
}

// DeleteTimeSpan implements StrictServerInterface.
func (s *Server) DeleteTimeSpan(ctx context.Context, request DeleteTimeSpanRequestObject) (DeleteTimeSpanResponseObject, error) {
	return s.timeSpans.DeleteTimeSpan(ctx, request)
}

// GetProject implements StrictServerInterface.
func (s *Server) GetProject(ctx context.Context, request GetProjectRequestObject) (GetProjectResponseObject, error) {
	return s.projects.GetProject(ctx, request)
}

// GetTag implements StrictServerInterface.
func (s *Server) GetTag(ctx context.Context, request GetTagRequestObject) (GetTagResponseObject, error) {
	return s.tags.GetTag(ctx, request)
}

// GetTimeSpan implements StrictServerInterface.
func (s *Server) GetTimeSpan(ctx context.Context, request GetTimeSpanRequestObject) (GetTimeSpanResponseObject, error) {
	return s.timeSpans.GetTimeSpan(ctx, request)
}

// ListProjects implements StrictServerInterface.
func (s *Server) ListProjects(ctx context.Context, request ListProjectsRequestObject) (ListProjectsResponseObject, error) {
	return s.projects.ListProjects(ctx, request)
}

// ListTags implements StrictServerInterface.
func (s *Server) ListTags(ctx context.Context, request ListTagsRequestObject) (ListTagsResponseObject, error) {
	return s.tags.ListTags(ctx, request)
}

// ListTimeSpans implements StrictServerInterface.
func (s *Server) ListTimeSpans(ctx context.Context, request ListTimeSpansRequestObject) (ListTimeSpansResponseObject, error) {
	return s.timeSpans.ListTimeSpans(ctx, request)
}

// UpdateProject implements StrictServerInterface.
func (s *Server) UpdateProject(ctx context.Context, request UpdateProjectRequestObject) (UpdateProjectResponseObject, error) {
	return s.projects.UpdateProject(ctx, request)
}

// UpdateTag implements StrictServerInterface.
func (s *Server) UpdateTag(ctx context.Context, request UpdateTagRequestObject) (UpdateTagResponseObject, error) {
	return s.tags.UpdateTag(ctx, request)
}

// UpdateTimeSpan implements StrictServerInterface.
func (s *Server) UpdateTimeSpan(ctx context.Context, request UpdateTimeSpanRequestObject) (UpdateTimeSpanResponseObject, error) {
	return s.timeSpans.UpdateTimeSpan(ctx, request)
}
