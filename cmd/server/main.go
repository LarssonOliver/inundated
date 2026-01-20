package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/larssonoliver/inundated/internal/api"
	"github.com/larssonoliver/inundated/internal/api/handlers"
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
	
	h := api.HandlerFromMux(api.NewStrictHandler(server, nil), r)

	s := &http.Server{
		Handler: h,
		Addr:    "0.0.0.0:8080",
	}

	log.Fatal(s.ListenAndServe())
}
