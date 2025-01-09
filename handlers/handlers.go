package handlers

import (
	"encoding/json"
	"net/http"
)

// HomeHandler responds to the root endpoint
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"message": "Welcome to our Go Web Server",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HealthCheckHandler provides a simple health check endpoint
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "healthy",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
