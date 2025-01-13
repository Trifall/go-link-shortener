package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ShortenRequest struct {
	URL       string `json:"url"`
	Duration  int    `json:"duration"`
	SecretKey string `json:"key"`
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
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}
}
