package handlers

import (
	"io/fs"
	"net/http"

	"github.com/larssonoliver/inundated/frontend"
)

const prefix = "dist"

func spaHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to serve static file
		_, err := frontend.FrontendFS.Open(prefix + r.URL.Path)
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
	sub, err := fs.Sub(frontend.FrontendFS, prefix)

	if err != nil {
		panic(err)
	}

	handler := http.FileServer(http.FS(sub))
	return spaHandler(handler)
}
