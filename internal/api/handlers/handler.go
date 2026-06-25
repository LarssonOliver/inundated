package handlers

import (
	"github.com/larssonoliver/inundated/internal/api"
	"github.com/larssonoliver/inundated/internal/service"
)

type Handler struct {
	UserHandler
	TagHandler
	ProjectHandler
	TimespanHandler
}

var _ api.HttpHandler = (*Handler)(nil)

func NewHandler(svc service.Service) *Handler {
	return &Handler{
		UserHandler:     *NewUserHandler(svc),
		TagHandler:      *NewTagHandler(svc),
		ProjectHandler:  *NewProjectHandler(svc),
		TimespanHandler: *NewTimespanHandler(svc),
	}
}
