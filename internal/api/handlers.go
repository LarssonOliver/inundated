package api

import "context"

type HttpHandler interface {
	UserHandler
	TagHandler
	ProjectHandler
	TimespanHandler
}

type UserHandler interface {
	GetCurrentUser(ctx context.Context, request GetCurrentUserRequestObject) (GetCurrentUserResponseObject, error)
	UpdateCurrentUser(ctx context.Context, request UpdateCurrentUserRequestObject) (UpdateCurrentUserResponseObject, error)
}

type TagHandler interface {
	GetTag(ctx context.Context, request GetTagRequestObject) (GetTagResponseObject, error)
	ListTags(ctx context.Context, request ListTagsRequestObject) (ListTagsResponseObject, error)
	CreateTag(ctx context.Context, request CreateTagRequestObject) (CreateTagResponseObject, error)
	UpdateTag(ctx context.Context, request UpdateTagRequestObject) (UpdateTagResponseObject, error)
	DeleteTag(ctx context.Context, request DeleteTagRequestObject) (DeleteTagResponseObject, error)
}

type ProjectHandler interface {
	GetProject(ctx context.Context, request GetProjectRequestObject) (GetProjectResponseObject, error)
	ListProjects(ctx context.Context, request ListProjectsRequestObject) (ListProjectsResponseObject, error)
	CreateProject(ctx context.Context, request CreateProjectRequestObject) (CreateProjectResponseObject, error)
	UpdateProject(ctx context.Context, request UpdateProjectRequestObject) (UpdateProjectResponseObject, error)
	DeleteProject(ctx context.Context, request DeleteProjectRequestObject) (DeleteProjectResponseObject, error)
	GetProjectStats(ctx context.Context, request GetProjectStatsRequestObject) (GetProjectStatsResponseObject, error)
}

type TimespanHandler interface {
	GetTimespan(ctx context.Context, request GetTimespanRequestObject) (GetTimespanResponseObject, error)
	ListTimespans(ctx context.Context, request ListTimespansRequestObject) (ListTimespansResponseObject, error)
	CreateTimespan(ctx context.Context, request CreateTimespanRequestObject) (CreateTimespanResponseObject, error)
	UpdateTimespan(ctx context.Context, request UpdateTimespanRequestObject) (UpdateTimespanResponseObject, error)
	DeleteTimespan(ctx context.Context, request DeleteTimespanRequestObject) (DeleteTimespanResponseObject, error)
}
