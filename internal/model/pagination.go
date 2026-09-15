package model

func DefaultPaginationParams() PaginationParams {
	return PaginationParams{
		Limit:  25, // From OpenAPI spec
		Offset: 0,
	}
}

type PaginationParams struct {
	Limit  int
	Offset int

	// IncludeArchived, when false (the default), excludes archived
	// projects/tags from List results. Ignored by list operations on
	// resources that don't support archiving.
	IncludeArchived bool
}

type Page[T any] struct {
	Data       []T
	TotalCount int
	Limit      int
	Offset     int
}
