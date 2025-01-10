package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

// HomeHandler responds to the root endpoint
// HomeHandler responds to the root endpoint
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

type ShortenRequest struct {
	URL       string `json:"url"`
	Duration  int    `json:"duration"`
	SecretKey string `json:"secret_key"`
}

type ShortenResponse struct {
	ShortenedURL string `json:"shortened_url"`
}

func ShortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var request ShortenRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := ErrorResponse{Message: "Invalid request body"}
		if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
			http.Error(w, "Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	response := ShortenResponse{
		ShortenedURL: fmt.Sprintf("Received: %s, Duration: %d", request.URL, request.Duration),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
