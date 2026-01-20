package api

import "context"

type HttpHandler interface {
	TagHandler
	ProjectHandler
	TimeSpanHandler
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
}

type TimeSpanHandler interface {
	GetTimeSpan(ctx context.Context, request GetTimeSpanRequestObject) (GetTimeSpanResponseObject, error)
	ListTimeSpans(ctx context.Context, request ListTimeSpansRequestObject) (ListTimeSpansResponseObject, error)
	CreateTimeSpan(ctx context.Context, request CreateTimeSpanRequestObject) (CreateTimeSpanResponseObject, error)
	UpdateTimeSpan(ctx context.Context, request UpdateTimeSpanRequestObject) (UpdateTimeSpanResponseObject, error)
	DeleteTimeSpan(ctx context.Context, request DeleteTimeSpanRequestObject) (DeleteTimeSpanResponseObject, error)
}
