package api

import (
	"encoding/json"
	"go-link-shortener/models"
	"log"
	"net/http"
)

type ErrorResponseConfig struct {
	Status    int
	Message   string
	LogType   models.LogType
	LogSource models.LogSource
	Request   *http.Request
	CtxValues *ContextValues
	Addendum  string
}

// helper func to log and write error response
func writeErrorResponse(w http.ResponseWriter, config ErrorResponseConfig) {
	w.WriteHeader(config.Status)
	errorResponse := ErrorResponse{Message: config.Message}

	// Log the error
	logMessage := config.Message + " | IP: " + config.Request.RemoteAddr
	if config.CtxValues != nil && config.CtxValues.SecretKey != "" {
		logMessage += " | Requested by: '" + config.CtxValues.SecretKey + "'"
	}
	if config.Addendum != "" {
		logMessage += " | " + config.Addendum
	}
	// add addendum to log message
	models.CreateLog(config.LogType, config.LogSource, logMessage)

	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Server Error", http.StatusInternalServerError)
	}
}
