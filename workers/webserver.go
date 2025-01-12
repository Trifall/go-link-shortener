package workers

import (
	"go-link-shortener/api"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
)

func InitializeWebserver() error {
	log.Println("⏳ Initializing API...")
	// Create a new chi router
	r := chi.NewRouter()

	// TODO: might need to update this to allow for the frontend to connect
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

	r.Mount("/api", api.InitializeAPIRouter())

	log.Println("✔️  API initialized successfully.")
	log.Println("✔️  Starting server on :8080")

	// Start the server
	if err := http.ListenAndServe(":8080", r); err != nil {
		return err
	}

	return nil
}
