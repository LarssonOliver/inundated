package handlers

import (
	"context"
	"github.com/larssonoliver/inundated/internal/api"
)

type HealthHandler struct{}

var _ api.HealthHandler = (*HealthHandler)(nil)

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthCheck implements [api.HealthHandler].
func (h *HealthHandler) HealthCheck(ctx context.Context, request api.HealthCheckRequestObject) (api.HealthCheckResponseObject, error) {
	return api.HealthCheck200Response{}, nil
}
