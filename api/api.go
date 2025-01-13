package api

import (
	"github.com/go-chi/chi"
)

func InitializeAPIRouter() chi.Router {
	r := chi.NewRouter()

	// Mount handlers
	r.Get("/", HomeHandler)
	r.Get("/health", HealthCheckHandler)

	r.Mount("/v1", V1Router())

	return r
}

func V1Router() chi.Router {
	r := chi.NewRouter()

	// authentication middleware for all API routes
	r.Use(AuthMiddleware)

	r.Post("/shorten", ShortenHandler)

	// Use AdminOnlyMiddleware for admin only routes
	r.With(AdminOnlyMiddleware).Post("/validate", ValidateKeyHandler)

	return r
}
