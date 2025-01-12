package api

import (
	"encoding/json"
	"go-link-shortener/auth"
	"go-link-shortener/utils"
	"log"
	"net/http"
)

type ValidateKeyResponse struct {
	Key       string `json:"key"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Active    bool   `json:"active"`
	IsAdmin   bool   `json:"is_admin"`
}

type ContextValues struct {
	SecretKey string
	IsAdmin   bool
}

func ValidateKeyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Use the helper function to check if the request is unauthorized
	if CheckUnauthorized(w, r) {
		return
	}

	// Retrieve the context values from the context
	ctxValues, _ := GetContextValues(r)

	log.Println("Validating - Key:", ctxValues.SecretKey, ", IsAdmin:", ctxValues.IsAdmin)

	keyObj, err := auth.ValidateKey(ctxValues.SecretKey)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		errorResponse := ErrorResponse{Message: err.Error()}
		if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
			http.Error(w, "Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	response := ValidateKeyResponse{
		Key:       keyObj.Key,
		Name:      keyObj.Name,
		CreatedAt: utils.SafeString(&keyObj.CreatedAt),
		UpdatedAt: utils.SafeString(&keyObj.UpdatedAt),
		Active:    keyObj.Active,
		IsAdmin:   keyObj.IsAdmin,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}
}
