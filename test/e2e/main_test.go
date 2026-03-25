//go:generate go tool oapi-codegen --config oapi-codegen/client.cfg.yaml ../../openapi/dist/inundated.yaml

package e2e_test

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/larssonoliver/inundated/internal/api"
	"github.com/larssonoliver/inundated/internal/api/handlers"
	"github.com/larssonoliver/inundated/internal/repository/memory"
	"github.com/larssonoliver/inundated/internal/service"
)

const timeout = 2 * time.Second

var port int

func startServer() *http.Server {
	repo := memory.NewMemoryStore()
	svc := service.NewService(repo)
	handler := handlers.NewHandler(svc)
	server := api.NewServer(handler)

	r := chi.NewMux()
	h := api.HandlerFromMux(api.NewStrictHandler(server, nil), r)

	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("failed to start listener: %v", err)
	}

	port = l.Addr().(*net.TCPAddr).Port

	s := &http.Server{
		Handler: h,
	}

	go func() {
		if err := s.Serve(l); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	return s
}

func newClient() *ClientWithResponses {
	c, err := NewClientWithResponses(
		fmt.Sprintf("http://localhost:%d", port),
		WithHTTPClient(&http.Client{
			Timeout: timeout,
		}),
	)

	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	return c
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	s := startServer()

	code := m.Run()

	_ = s.Shutdown(ctx)

	os.Exit(code)
}
