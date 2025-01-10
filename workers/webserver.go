package workers

import (
	"go-link-shortener/handlers"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
)

func InitializeWebserver() error {
	log.Println("⏳ Initializing routes and handlers...")
	// Create a new chi router
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Middleware stack
	r.Use(middleware.Logger)    // Log API requests
	r.Use(middleware.Recoverer) // Recover from panics without crashing server

	// Mount handlers
	r.Get("/", handlers.HomeHandler)
	r.Get("/health", handlers.HealthCheckHandler)
	r.Post("/shorten", handlers.ShortenHandler)

	log.Println("✔️  Routes and handlers initialized successfully.")
	log.Println("✔️  Starting server on :8080")

	// Start the server
	if err := http.ListenAndServe(":8080", r); err != nil {
		return err
	}

	return nil
}
