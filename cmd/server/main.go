package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
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
	"github.com/larssonoliver/inundated/internal/logging"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/repository/memory"
	"github.com/larssonoliver/inundated/internal/repository/postgres"
	"github.com/larssonoliver/inundated/internal/service"
)

var Version = "dev"

// flushLogs drains the async log writer. Set once the logger is built; called
// before any os.Exit so a final error line is not lost in the buffer.
var flushLogs = func() {}

// fatal logs msg at error level with the given key/value args, then exits 1.
func fatal(msg string, args ...any) {
	slog.Error(msg, args...)
	flushLogs()
	os.Exit(1)
}

func setupRepositories(ctx context.Context, databaseUrl string) (
	repository.Repository,
	repository.LoginStateRepository,
	repository.SessionRepository,
) {
	if databaseUrl == "in-memory" {
		slog.Warn("using in-memory repository", "note", "not recommended for production")
		memoryStore := memory.NewMemoryStore()
		return memoryStore, memoryStore, memoryStore
	}
	if strings.HasPrefix(databaseUrl, "postgresql://") {
		slog.Info("using postgres repository")
		slog.Info("applying database migrations")
		err := postgresdb.ApplyMigrations(ctx, databaseUrl)
		if err != nil {
			fatal("failed to apply database migrations", "error", err)
		}

		repository, err := postgres.NewPostgresStore(ctx, databaseUrl)
		if err != nil {
			fatal("failed to connect to postgresql", "error", err)
		}
		return repository, repository, repository
	}
	fatal("unsupported database url", "url", databaseUrl)
	return nil, nil, nil
}

func shouldUseSecureCookies(cfg *config.Config) bool {
	return strings.HasPrefix(cfg.PublicBaseURL, "https://")
}

func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: handler,
		// Built after slog.SetDefault (see main) so it binds the configured
		// handler, not the bootstrap default.
		ErrorLog:          slog.NewLogLogger(slog.Default().Handler(), slog.LevelError),
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
) http.Handler {

	r := chi.NewMux()

	isSecure := shouldUseSecureCookies(cfg)

	r.Use(chimiddleware.RequestID)
	r.Use(middleware.RequestLogContext)
	r.Use(middleware.RealIP(cfg.TrustedProxies, cfg.TrustedProxyHeaders))
	r.Use(middleware.RequestLogger(func(r *http.Request) bool {
		return r.URL.Path == "/health" || !strings.HasPrefix(r.URL.Path, "/api/")
	}))
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.SecurityHeaders)

	r.Handle("/health", handlers.HealthHandler())

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimitByIP(middleware.APIRateLimitRequests, middleware.APIRateLimitWindow))
		r.Use(middleware.RateLimitByIPForPrefixes(middleware.AuthRateLimitRequests, middleware.AuthRateLimitWindow, "/api/auth/"))
		r.Use(middleware.MaxBodyBytes(middleware.MaxAPIBodyBytes))
		r.Use(middleware.NoSniffJSON)
		r.Use(middleware.CrossOriginProtection(cfg.PublicBaseURL))

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

	logger, flush, err := logging.New(logging.Options{
		Level:  cfg.LogLevel,
		Format: cfg.LogFormat,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "logging setup: %v\n", err)
		os.Exit(1)
	}
	flushLogs = flush
	defer flushLogs()
	slog.SetDefault(logger)

	ctx := context.Background()

	repo, loginStateRepo, sessionRepo := setupRepositories(ctx, cfg.DatabaseURL)
	svc := service.NewService(repo, service.WithRegistrationDisabled(cfg.DisableUserRegistration))

	if err := service.EnsureAuthConfigConsistent(ctx, repo, cfg.OIDC.Enabled()); err != nil {
		fatal("auth configuration", "error", err)
	}

	if cfg.DisableUserRegistration {
		slog.Info("user registration disabled; only OIDC identities with an existing account can log in")
		if hasUsers, err := repo.HasUsers(ctx); err == nil && !hasUsers {
			slog.Warn("registration is disabled and no users exist yet -- nobody can log in until you enable it once")
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

	r := newRouter(cfg, svc, sessionRepo, server)

	addrStr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	s := newHTTPServer(addrStr, r)

	if cfg.OIDC.Enabled() {
		slog.Info("OIDC authentication enabled", "issuer", cfg.OIDC.IssuerURL)
	} else {
		slog.Info("OIDC not configured; running in userless mode")
	}

	if !shouldUseSecureCookies(cfg) {
		slog.Warn("insecure cookies (Secure attribute off); set PUBLIC_BASE_URL to an https:// origin in production")
	}

	if len(cfg.TrustedProxies) == 0 {
		slog.Warn("TRUSTED_PROXIES not set; forwarded-for headers are ignored and rate " +
			"limiting keys off the direct connection address. Behind a reverse proxy that " +
			"means every client shares one bucket -- set TRUSTED_PROXIES to the proxy's address(es)")
	}

	slog.Info("starting inundated", "version", Version, "addr", addrStr)
	if err := s.ListenAndServe(); err != nil {
		fatal("http server stopped", "error", err)
	}
}
