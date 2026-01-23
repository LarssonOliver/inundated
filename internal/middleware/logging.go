package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func RequestLogger(
	logger *log.Logger,
	skipFn func(r *http.Request) bool,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if skipFn != nil && skipFn(r) {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			logger.Printf(
				"%s %s %d %dB %s",
				r.Method,
				r.URL.Path,
				ww.Status(),
				ww.BytesWritten(),
				time.Since(start),
			)
		})
	}
}
