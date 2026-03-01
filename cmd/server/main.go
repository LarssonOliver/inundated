package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/larssonoliver/inundated/internal/api"
	"github.com/larssonoliver/inundated/internal/api/handlers"
	"github.com/larssonoliver/inundated/internal/api/middleware"
	"github.com/larssonoliver/inundated/internal/config"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/repository/memory"
	"github.com/larssonoliver/inundated/internal/service"
)

func setupRepository() repository.Repository {
	return memory.NewMemoryStore()
}

func main() {
	cfg, err := config.Load()
	if err != nil {
        if errors.Is(err, flag.ErrHelp) {
            os.Exit(0) // -help is not an error
        }
        fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
        os.Exit(1)
    }

	repo := setupRepository()
	svc := service.NewService(repo)
	handler := handlers.NewHandler(svc)
	server := api.NewServer(handler)

	r := chi.NewMux()

	r.Handle("/health", handlers.HealthHandler())

	r.Route("/api", func(r chi.Router) {
		logger := log.New(os.Stdout, "[http] ", log.LstdFlags)
		r.Use(middleware.RequestLogger(logger, func(r *http.Request) bool {
			return r.URL.Path == "/health"
		}))

		api.HandlerFromMux(api.NewStrictHandler(server, nil), r)
	})

	r.Group(func(r chi.Router) {
		r.Handle("/*", handlers.FrontendHandler())
	})

	addrStr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	s := &http.Server{
		Handler: r,
		Addr:    addrStr,
	}

	log.Println("Starting server on", addrStr)
	log.Fatal(s.ListenAndServe())
}
