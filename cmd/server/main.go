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

// resolveCSRFKey returns the configured CSRF key, or a random ephemeral one when
// none is set. An ephemeral key is fine for local use but means CSRF tokens
// (and thus in-flight form submissions) don't survive a restart.
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

// secureCookies reports whether cookies should carry the Secure attribute and
// whether gorilla/csrf should enforce its HTTPS Origin rules. It is on only when
// the app's public origin is an https:// URL; a plain-HTTP dev server (no
// PUBLIC_BASE_URL, or an http:// one) needs it off or the browser drops the
// cookies and the CSRF check rejects the http:// Origin.
func secureCookies(cfg *config.Config) bool {
	return strings.HasPrefix(cfg.PublicBaseURL, "https://")
}

// newRouter wires the HTTP middleware stack and routes. API routes are served at
// the paths declared in the OpenAPI spec (/api/...), with the generated handler
// mounted at the root. When OIDC is configured, everything under /api except
// PublicAPIPaths requires a session; in userless mode the /api/auth/* routes are
// hidden entirely.
func newRouter(cfg *config.Config, svc service.Service, sessionRepo repository.SessionRepository, server api.StrictServerInterface, csrfKey []byte) http.Handler {
	r := chi.NewMux()

	secure := secureCookies(cfg)

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.CSRF(csrfKey, secure))
	r.Use(middleware.ExposeCSRFToken(secure))

	r.Handle("/health", handlers.HealthHandler())

	r.Group(func(r chi.Router) {
		logger := log.New(os.Stdout, "[http] ", log.LstdFlags)
		r.Use(middleware.RequestLogger(logger, func(r *http.Request) bool {
			return r.URL.Path == "/health"
		}))
		r.Use(middleware.NoSniffJSON)

		if cfg.OIDC.Enabled() {
			r.Use(middleware.OIDCAuth(svc, sessionRepo, secure))
			r.Use(middleware.RequireAuth(middleware.PublicAPIPaths...))
		} else {
			// No provider to talk to: the OIDC routes don't exist here.
			r.Use(middleware.RejectPathPrefixes("/api/auth/"))
		}

		api.HandlerFromMux(api.NewStrictHandler(server, nil), r)
	})

	// Any /api/* path the generated handler does not claim is a real 404, not a
	// route into the SPA. Without this it falls through to the frontend handler
	// below and returns index.html with a 200, which API clients (and the SPA's
	// own error handling) can misread as success.
	r.Handle("/api/*", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	}))

	r.Group(func(r chi.Router) {
		r.Handle("/*", handlers.FrontendHandler())
	})

	return r
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

	ctx := context.Background()

	repo, loginStateRepo, sessionRepo := setupRepositories(ctx, cfg.DatabaseURL)
	svc := service.NewService(repo)

	if err := service.EnsureAuthConfigConsistent(ctx, repo, cfg.OIDC.Enabled()); err != nil {
		log.Fatalf("auth configuration: %v", err)
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
	handler := handlers.NewHandler(authSvc, svc, secureCookies(cfg))
	server := api.NewServer(handler)

	r := newRouter(cfg, svc, sessionRepo, server, resolveCSRFKey(cfg.CSRFAuthKey))

	addrStr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	s := &http.Server{
		Handler: r,
		Addr:    addrStr,
	}

	if cfg.OIDC.Enabled() {
		log.Printf("OIDC authentication enabled (issuer: %s)", cfg.OIDC.IssuerURL)
	} else {
		log.Printf("OIDC not configured; running in userless mode")
	}

	if !secureCookies(cfg) {
		log.Printf("warning: insecure cookies (Secure attribute off, CSRF HTTPS-origin checks relaxed); set PUBLIC_BASE_URL to an https:// origin in production")
	}

	log.Printf("Starting inundated %s on %s", Version, addrStr)
	log.Fatal(s.ListenAndServe())
}
