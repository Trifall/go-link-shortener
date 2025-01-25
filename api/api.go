package api

import (
	"encoding/json"
	"go-link-shortener/database"
	"go-link-shortener/lib"
	"go-link-shortener/models"
	"go-link-shortener/utils"
	"net/http"
	"time"

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
			// validates self link
			r.Post(lib.ROUTES.Links.Delete, DeleteLinkHandler)
			// validates self link
			r.Post(lib.ROUTES.Links.Update, UpdateLinkHandler)
			// validates self link
			r.Post(lib.ROUTES.Links.RetrieveAllByKey, RetrieveAllLinksByKeyHandler)

			r.Group(func(r chi.Router) {
				// Use AdminOnlyMiddleware for admin only routes
				r.Use(AdminOnlyMiddleware)
				r.Get(lib.ROUTES.Links.RetrieveAll, RetrieveAllLinksHandler)
			})
		})

		// !Admin Routes Below!
		r.Group(func(r chi.Router) {
			// Use AdminOnlyMiddleware for admin only routes
			r.Use(AdminOnlyMiddleware)
			r.Route(lib.ROUTES.Keys.Base, func(r chi.Router) {
				r.Post(lib.ROUTES.Keys.Validate, ValidateKeyHandler)
				r.Get(lib.ROUTES.Keys.RetrieveAll, RetrieveAllKeysHandler)
				r.Post(lib.ROUTES.Keys.Generate, GenerateKeyHandler)
				r.Post(lib.ROUTES.Keys.Update, UpdateKeyHandler)
				r.Post(lib.ROUTES.Keys.Delete, DeleteKeyHandler)
			})
		})

	})

	return r
}

func RedirectRouter() chi.Router {
	r := chi.NewRouter()

	env := utils.LoadEnv()

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		// blank path, redirect to docs
		if r.URL.Path == "/" {
			if env.ENABLE_DOCS == "false" {
				http.Redirect(w, r, lib.ROUTES.NotFound, http.StatusFound)
				return
			}
			http.Redirect(w, r, lib.ROUTES.Docs+"/", http.StatusFound)
			return
		}

		db := database.GetDB()

		// remove leading slash
		fixedPath := r.URL.Path[1:]
		linkObj, err := RetrieveRedirectURL(db, fixedPath)

		if err != nil {
			// check if err is "record not found"
			if err.Error() == "record not found" {
				http.Error(w, "Link with shortened string '"+fixedPath+"' not found", http.StatusNotFound)
				return
			}
			// if not, return internal server error
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// update the last visited time, increment visits, and add a new record to the link_visits table
		now := time.Now()
		linkObj.Visits += 1
		linkObj.LastVisitedAt = &now
		db.Save(linkObj)

		userAgent := r.Header.Get("User-Agent")
		ipAddress := r.RemoteAddr
		referrer := r.Header.Get("Referer")

		db.Model(&linkObj).Association("Visits").Append(&models.LinkVisit{
			VisitedAt: now,
			UserAgent: &userAgent,
			IPAddress: &ipAddress,
			Referrer:  &referrer,
		})

		http.Redirect(w, r, linkObj.RedirectTo, http.StatusMovedPermanently)
	})

	return r
}
