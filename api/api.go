package api

import (
	"encoding/json"
	"go-link-shortener/lib"
	"net/http"

	"github.com/go-chi/chi"
)

func InitializeAPIRouter() chi.Router {
	r := chi.NewRouter()

	// !Public Routes Below!
	r.Use(LogMiddleware)

	// Mount handlers
	r.Get("/", HomeHandler)
	r.Get(lib.ROUTES.Health, HealthCheckHandler)

	// Mount the V1 router
	r.Mount(lib.ROUTES.V1, V1Router())

	return r
}

func V1Router() chi.Router {
	r := chi.NewRouter()

	// !Auth Routes Below!

	r.Group(func(r chi.Router) {
		// Authentication middleware for all API routes
		r.Use(AuthMiddleware)

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			response := map[string]string{
				"message": "Welcome to the V1 API!",
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(response); err != nil {
				http.Error(w, "Server Error", http.StatusInternalServerError)
				return
			}
		})

		r.Route(lib.ROUTES.Links.Base, func(r chi.Router) {
			r.Post(lib.ROUTES.Links.Shorten, ShortenHandler)
			r.Post(lib.ROUTES.Links.Retrieve, RetrieveLinkHandler)
		})

		// !Admin Routes Below!
		r.Group(func(r chi.Router) {
			// Use AdminOnlyMiddleware for admin only routes
			r.Use(AdminOnlyMiddleware)
			r.Route(lib.ROUTES.Keys.Base, func(r chi.Router) {
				r.Post(lib.ROUTES.Keys.Validate, ValidateKeyHandler)
				r.Post(lib.ROUTES.Keys.Generate, GenerateKeyHandler)
				r.Post(lib.ROUTES.Keys.Update, UpdateKeyHandler)
				r.Post(lib.ROUTES.Keys.Delete, DeleteKeyHandler)
			})
		})
	})
	return r
}
