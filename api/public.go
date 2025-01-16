package api

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

// HomeHandler responds to the root endpoint
// @Summary Home endpoint
// @Description Returns a welcome message
// @Tags public
// @Produce json
// @Success 200 {object} map[string]string
// @Router / [get]
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"message": "Jerren's Link Shortener - there's nothing on this page!",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}
}

// HealthCheckHandler provides a simple health check endpoint
// @Summary Health check endpoint
// @Description Returns the health status of the service
// @Tags public
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "healthy",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}
}
