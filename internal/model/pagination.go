package model

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
