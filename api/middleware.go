package api

import (
	"context"
	"go-link-shortener/auth"
	"go-link-shortener/lib"
	"go-link-shortener/utils"
	"net/http"
)

type ContextKey string

// Define a constant for the key
const secretKeyContextKey ContextKey = "secret_key"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the secret key from the request (assuming it's in the header)
		secretKey := r.Header.Get("Authorization")
		if secretKey == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Validate the key and get the key object
		keyObj, err := auth.ValidateKey(secretKey)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Save the updated key object back to the database
		db := utils.GetDB()
		if db == nil {
			http.Error(w, lib.ERRORS.Database, http.StatusInternalServerError)
			return
		}

		db.Save(&keyObj)

		// Attach the key object to the request context as a ContextValues struct
		ctxValues := ContextValues{
			SecretKey: keyObj.Key,
			IsAdmin:   keyObj.IsAdmin,
		}
		ctx := context.WithValue(r.Context(), secretKeyContextKey, ctxValues)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AdminOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the context values from the context
		ctxValues, ok := r.Context().Value(secretKeyContextKey).(ContextValues)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check if the key has admin privileges
		if !ctxValues.IsAdmin {
			http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
			return
		}

		// Call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}
