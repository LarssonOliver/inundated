package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/larssonoliver/inundated/internal/api"
	"github.com/larssonoliver/inundated/internal/api/handlers"
	"github.com/larssonoliver/inundated/internal/api/middleware"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/repository/memory"
	"github.com/larssonoliver/inundated/internal/service"
)

func setupRepository() repository.Repository {
	return memory.NewMemoryStore()
}

func main() {
	repo := setupRepository()
	svc := service.NewService(repo)
	handler := handlers.NewHandler(svc)
	server := api.NewServer(handler)

	r := chi.NewMux()

	logger := log.New(os.Stdout, "[http] ", log.LstdFlags)
	r.Use(middleware.RequestLogger(logger, func(r *http.Request) bool {
		return r.URL.Path == "/health"
	}))

	r.Group(func(r chi.Router) {
		r.Use(middleware.OpenApiRequestValidator())
		_ = api.HandlerFromMux(api.NewStrictHandler(server, nil), r)
	})

	s := &http.Server{
		Handler: r,
		Addr:    "0.0.0.0:8080",
	}

	log.Println("Starting server on :8080")
	log.Fatal(s.ListenAndServe())
}
