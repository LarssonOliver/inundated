package handlers

import (
	"net/http"

	"github.com/larssonoliver/inundated/internal/api"
	"github.com/larssonoliver/inundated/internal/service"
)

type Handler struct {
	AuthHandler
	UserHandler
	TagHandler
	ProjectHandler
	TimespanHandler
}

var _ api.HttpHandler = (*Handler)(nil)

func NewHandler(authSvc service.AuthService, svc service.Service, secureCookies bool) *Handler {
	return &Handler{
		UserHandler:     *NewUserHandler(svc),
		TagHandler:      *NewTagHandler(svc),
		ProjectHandler:  *NewProjectHandler(svc),
		TimespanHandler: *NewTimespanHandler(svc),

		AuthHandler: *NewAuthHandler(authSvc, secureCookies),
	}
}

func HealthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message":"OK"}`))
	})
}

func NotFoundHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	})
}
