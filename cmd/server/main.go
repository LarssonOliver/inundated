package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/larssonoliver/inundated/internal/api"
	"github.com/larssonoliver/inundated/internal/api/handlers"
	"github.com/larssonoliver/inundated/internal/api/middleware"
	"github.com/larssonoliver/inundated/internal/auth"
	"github.com/larssonoliver/inundated/internal/config"
	postgresdb "github.com/larssonoliver/inundated/internal/db/postgres"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/repository/memory"
	"github.com/larssonoliver/inundated/internal/repository/postgres"
	"github.com/larssonoliver/inundated/internal/service"
)

var Version = "dev"

func setupRepositories(ctx context.Context, databaseUrl string) (repository.Repository, repository.LoginStateRepository, repository.SessionRepository) {
	if databaseUrl == "in-memory" {
		log.Println("Using in-memory repository (not recommended for production)")
		memoryStore := memory.NewMemoryStore()
		return memoryStore, memoryStore, memoryStore
	}
	if strings.HasPrefix(databaseUrl, "postgresql://") {
		log.Printf("Using postgres repository")
		log.Printf("Applying database migrations...")
		err := postgresdb.ApplyMigrations(ctx, databaseUrl)
		if err != nil {
			log.Fatalf("failed to apply database migrations: %v", err)
		}

		repository, err := postgres.NewPostgresStore(ctx, databaseUrl)
		if err != nil {
			log.Fatalf("failed to connect to PostgreSQL: %v", err)
		}
		return repository, repository, repository
	}
	log.Fatalf("unsupported database URL: %s", databaseUrl)
	return nil, nil, nil
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

	oidcClient := auth.NewOIDCClient()

	repo, loginStateRepo, sessionRepo := setupRepositories(context.Background(), cfg.DatabaseURL)
	svc := service.NewService(repo)
	authSvc := service.NewAuthService(svc, sessionRepo, loginStateRepo, oidcClient)
	handler := handlers.NewHandler(authSvc, svc)
	server := api.NewServer(handler)

	r := chi.NewMux()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.CSRFHeader)

	r.Handle("/health", handlers.HealthHandler())

	r.Route("/api", func(r chi.Router) {
		logger := log.New(os.Stdout, "[http] ", log.LstdFlags)
		r.Use(middleware.RequestLogger(logger, func(r *http.Request) bool {
			return r.URL.Path == "/health"
		}))
		r.Use(middleware.NoSniffJSON)
		r.Use(middleware.OIDCAuth(svc, sessionRepo))
		r.Use(middleware.RequireAuth())

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

	log.Printf("Starting inundated %s on %s", Version, addrStr)
	log.Fatal(s.ListenAndServe())
}
