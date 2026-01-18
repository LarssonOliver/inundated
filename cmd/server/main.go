package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/larssonoliver/inundated/internal/api"
	genapi "github.com/larssonoliver/inundated/internal/api/gen"
)

func main() {
	server := api.NewServer()

	r := chi.NewMux()

	m := []genapi.StrictMiddlewareFunc{}
	
	h := genapi.HandlerFromMux(genapi.NewStrictHandler(server, m), r)

	s := &http.Server{
		Handler: h,
		Addr:    "0.0.0.0:8080",
	}

	log.Fatal(s.ListenAndServe())
}
