package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"go-link-shortener/handlers"
)

func main() {
	// Create a new chi router
	r := chi.NewRouter()

	// Add some useful middleware
	r.Use(middleware.Logger)    // Log API requests
	r.Use(middleware.Recoverer) // Recover from panics without crashing server

	// Mount our handlers
	r.Get("/", handlers.HomeHandler)
	r.Get("/health", handlers.HealthCheckHandler)

	// Start the server
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
