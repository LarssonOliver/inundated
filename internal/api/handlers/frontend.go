package handlers

import (
	"net/http"
	"strings"

	"github.com/larssonoliver/inundated/frontend"
)

func spaHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")

		// Try to serve static file
		_, err := frontend.FrontendFS.Open(path)
		if err == nil {
			handler.ServeHTTP(w, r)
			return
		}

		// Otherwise serve index.html (Vue router)
		r.URL.Path = "/"
		handler.ServeHTTP(w, r)
	})
}

func FrontendHandler() http.Handler {
	handler := http.FileServer(http.FS(frontend.FrontendFS))
	return spaHandler(handler)
}
