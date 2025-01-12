package api

import (
	"encoding/json"
	"net/http"
)

// GetContextValues extracts the ContextValues from the request context.
func GetContextValues(r *http.Request) (ContextValues, bool) {
	ctxValues, ok := r.Context().Value(secretKeyContextKey).(ContextValues)
	return ctxValues, ok
}

// CheckUnauthorized checks if the context values are valid and writes an unauthorized response if not.
func CheckUnauthorized(w http.ResponseWriter, r *http.Request) bool {
	ctxValues, ok := GetContextValues(r)
	_ = ctxValues
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		errorResponse := ErrorResponse{Message: "Unauthorized"}
		if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
			http.Error(w, "Server Error", http.StatusInternalServerError)
		}
		return true
	}
	return false
}
