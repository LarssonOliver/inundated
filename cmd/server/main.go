package main

import (
	"context"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

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

func setupRepositories(ctx context.Context, databaseUrl string) (
	repository.Repository,
	repository.LoginStateRepository,
	repository.SessionRepository,
) {
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

func resolveCSRFKey(configured string) []byte {
	if configured != "" {
		return []byte(configured)
	}
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		log.Fatalf("failed to generate ephemeral CSRF key: %v", err)
	}
	log.Printf("warning: CSRF_AUTH_KEY not set; generated an ephemeral key")
	return key
}

func shouldUseSecureCookies(cfg *config.Config) bool {
	return strings.HasPrefix(cfg.PublicBaseURL, "https://")
}

func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}

func newRouter(
	cfg *config.Config,
	svc service.Service,
	sessionRepo repository.SessionRepository,
	server api.StrictServerInterface,
	csrfKey []byte,
) http.Handler {

	r := chi.NewMux()

	isSecure := shouldUseSecureCookies(cfg)

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.SecurityHeaders)

	r.Handle("/health", handlers.HealthHandler())

	r.Group(func(r chi.Router) {
		logger := log.New(os.Stdout, "[http] ", log.LstdFlags)
		r.Use(middleware.RequestLogger(logger, func(r *http.Request) bool {
			return r.URL.Path == "/health"
		}))
		r.Use(middleware.RateLimitByIP(middleware.APIRateLimitRequests, middleware.APIRateLimitWindow))
		r.Use(middleware.RateLimitByIPForPrefixes(middleware.AuthRateLimitRequests, middleware.AuthRateLimitWindow, "/api/auth/"))
		r.Use(middleware.MaxBodyBytes(middleware.MaxAPIBodyBytes))
		r.Use(middleware.NoSniffJSON)
		r.Use(middleware.CSRF(csrfKey, isSecure))
		r.Use(middleware.ExposeCSRFToken(isSecure))

		if cfg.OIDC.Enabled() {
			r.Use(middleware.OIDCAuth(svc, sessionRepo, isSecure))
			r.Use(middleware.RequireAuth(middleware.PublicAPIPaths...))
		} else {
			r.Use(middleware.RejectPathPrefixes("/api/auth/"))
		}

		api.HandlerFromMux(api.NewStrictHandler(server, nil), r)
	})

	r.Handle("/api/*", handlers.NotFoundHandler())

	r.Group(func(r chi.Router) {
		r.Handle("/*", handlers.FrontendHandler())
	})

	return r
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	repo, loginStateRepo, sessionRepo := setupRepositories(ctx, cfg.DatabaseURL)
	svc := service.NewService(repo, service.WithRegistrationDisabled(cfg.DisableUserRegistration))

	if err := service.EnsureAuthConfigConsistent(ctx, repo, cfg.OIDC.Enabled()); err != nil {
		log.Fatalf("auth configuration: %v", err)
	}

	if cfg.DisableUserRegistration {
		log.Printf("user registration disabled; only OIDC identities with an existing account can log in")
		if hasUsers, err := repo.HasUsers(ctx); err == nil && !hasUsers {
			log.Printf("warning: registration is disabled and no users exist yet -- nobody can log in until you enable it once")
		}
	}

	oidcClient := auth.NewOIDCClientWithConfig(auth.OIDCClientConfig{
		IssuerURL:    cfg.OIDC.IssuerURL,
		ClientID:     cfg.OIDC.ClientID,
		ClientSecret: cfg.OIDC.ClientSecret,
		RedirectURL:  cfg.OIDC.RedirectURL,
		Scopes:       cfg.OIDC.Scopes,
		HTTPTimeout:  cfg.OIDC.HTTPTimeout,
	})

	cleanupSvc := service.NewCleanupService(sessionRepo, loginStateRepo, 5*time.Minute)
	go cleanupSvc.Run(ctx)

	authSvc := service.NewAuthService(svc, sessionRepo, loginStateRepo, oidcClient)
	handler := handlers.NewHandler(authSvc, svc, shouldUseSecureCookies(cfg))
	server := api.NewServer(handler)

	r := newRouter(cfg, svc, sessionRepo, server, resolveCSRFKey(cfg.CSRFAuthKey))

	addrStr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	s := newHTTPServer(addrStr, r)

	if cfg.OIDC.Enabled() {
		log.Printf("OIDC authentication enabled (issuer: %s)", cfg.OIDC.IssuerURL)
	} else {
		log.Printf("OIDC not configured; running in userless mode")
	}

	if !shouldUseSecureCookies(cfg) {
		log.Printf("warning: insecure cookies (Secure attribute off, CSRF " +
			"HTTPS-origin checks relaxed); set PUBLIC_BASE_URL to an https://" +
			"origin in production")
	}

	log.Printf("Starting inundated %s on %s", Version, addrStr)
	log.Fatal(s.ListenAndServe())
}
