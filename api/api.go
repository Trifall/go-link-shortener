package api

import (
	"github.com/go-chi/chi"
)

func InitializeAPIRouter() chi.Router {
	r := chi.NewRouter()

	// !Public Routes Below!
	r.Use(LogMiddleware)

	// Mount handlers
	r.Get("/", HomeHandler)
	r.Get("/health", HealthCheckHandler)

	// Mount the V1 router
	r.Mount("/v1", V1Router())

	return r
}

func V1Router() chi.Router {
	r := chi.NewRouter()

	// !Auth Routes Below!
	// Authentication middleware for all API routes
	r.Use(AuthMiddleware)

	r.Post("/shorten", ShortenHandler)

	// !Admin Routes Below!
	// Use AdminOnlyMiddleware for admin only routes
	r.With(AdminOnlyMiddleware).Post("/validate", ValidateKeyHandler)
	r.With(AdminOnlyMiddleware).Post("/generate", GenerateKeyHandler)
	r.With(AdminOnlyMiddleware).Post("/delete", DeleteKeyHandler)

	return r
}
