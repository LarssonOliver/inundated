package handlers_test

import "github.com/larssonoliver/inundated/internal/api"

func ptrLimit(l int) *api.Limit {
	limit := api.Limit(l)
	return &limit
}

func ptrOffset(o int) *api.Offset {
	offset := api.Offset(o)
	return &offset
}

func ptrInterval(i string) *api.Interval {
	interval := api.Interval(i)
	return &interval
}

func boolPtr(b bool) *bool {
	return &b
}
