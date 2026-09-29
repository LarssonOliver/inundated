package handlers

import (
	"slices"

	"github.com/larssonoliver/inundated/internal/model"
)

// parsePagination applies the optional limit (1 to 100) and offset (not
// negative) query parameters to the default page, and reports false if
// either is out of range.
func parsePagination(limit, offset *int) (model.PaginationParams, bool) {
	params := model.DefaultPaginationParams()
	if limit != nil {
		if *limit < 1 || *limit > 100 {
			return model.PaginationParams{}, false
		}
		params.Limit = *limit
	}
	if offset != nil {
		if *offset < 0 {
			return model.PaginationParams{}, false
		}
		params.Offset = *offset
	}
	return params, true
}

// hasInclude reports whether an include query parameter asks for value,
// one of the endpoint's own include constants.
func hasInclude[T ~string](include *[]string, value T) bool {
	return include != nil && slices.Contains(*include, string(value))
}
