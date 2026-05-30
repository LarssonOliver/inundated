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
}

type Page[T any] struct {
	Data       []T
	TotalCount int
	Limit      int
	Offset     int
}
