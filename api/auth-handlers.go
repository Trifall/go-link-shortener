package api

import (
	"encoding/json"
	"go-link-shortener/auth"
	"go-link-shortener/models"
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
		config := ErrorResponseConfig{
			Status:    http.StatusMethodNotAllowed,
			Message:   "Method not allowed",
			LogType:   models.LogTypeError,
			LogSource: models.LogSourceAuth,
			Request:   r,
			CtxValues: nil,
			Addendum:  "",
		}
		writeErrorResponse(w, config)
		return
	}

	if CheckUnauthorized(w, r) {
		return // CheckUnauthorized handles its own error response
	}

	ctxValues, _ := GetContextValues(r)
	log.Println("Validating - Key:'"+ctxValues.SecretKey+"', IsAdmin:", ctxValues.IsAdmin)

	keyObj, err := auth.ValidateKey(ctxValues.SecretKey)
	if err != nil {
		config := ErrorResponseConfig{
			Status:    http.StatusUnauthorized,
			Message:   err.Error(),
			LogType:   models.LogTypeError,
			LogSource: models.LogSourceAuth,
			Request:   r,
			CtxValues: &ctxValues,
			Addendum:  "Requested by: '" + ctxValues.SecretKey + "'",
		}
		writeErrorResponse(w, config)
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
		config := ErrorResponseConfig{
			Status:    http.StatusInternalServerError,
			Message:   "Server Error",
			LogType:   models.LogTypeError,
			LogSource: models.LogSourceAuth,
			Request:   r,
			CtxValues: &ctxValues,
			Addendum:  "Requested by: '" + ctxValues.SecretKey + "'",
		}
		writeErrorResponse(w, config)
		return
	}

	models.CreateLog(models.LogTypeInfo, models.LogSourceAuth,
		"Validated key from IP Address: "+r.RemoteAddr+" with name: "+keyObj.Name+". Requested by: "+ctxValues.SecretKey)
}

type GenerateKeyRequest struct {
	Name    string `json:"name"`
	IsAdmin bool   `json:"is_admin"`
}

type GenerateKeyResponse struct {
	Key       string `json:"key"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	IsAdmin   bool   `json:"is_admin"`
}

func GenerateKeyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		config := ErrorResponseConfig{
			Status:    http.StatusMethodNotAllowed,
			Message:   "Method not allowed",
			LogType:   models.LogTypeError,
			LogSource: models.LogSourceAuth,
			Request:   r,
			CtxValues: nil,
			Addendum:  "",
		}
		writeErrorResponse(w, config)
		return
	}

	if CheckUnauthorized(w, r) {
		return // CheckUnauthorized handles its own error response
	}

	var request GenerateKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		config := ErrorResponseConfig{
			Status:    http.StatusBadRequest,
			Message:   "Invalid request body",
			LogType:   models.LogTypeError,
			LogSource: models.LogSourceAuth,
			Request:   r,
			CtxValues: nil,
			Addendum:  "",
		}
		writeErrorResponse(w, config)
		return
	}

	ctxValues, _ := GetContextValues(r)
	log.Println("Generating key with name:'"+request.Name+"', IsAdmin:", request.IsAdmin, ". Requested by: '"+ctxValues.SecretKey+"'")

	newKeyObj, err := auth.GenerateSecretKey(request.Name, request.IsAdmin)
	if err != nil {
		config := ErrorResponseConfig{
			Status:    http.StatusInternalServerError,
			Message:   err.Error(),
			LogType:   models.LogTypeError,
			LogSource: models.LogSourceAuth,
			Request:   r,
			CtxValues: &ctxValues,
			Addendum:  "Requested by: '" + ctxValues.SecretKey + "'",
		}
		writeErrorResponse(w, config)
		return
	}

	response := GenerateKeyResponse{
		Key:       newKeyObj.Key,
		Name:      newKeyObj.Name,
		CreatedAt: utils.SafeString(&newKeyObj.CreatedAt),
		IsAdmin:   newKeyObj.IsAdmin,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		config := ErrorResponseConfig{
			Status:    http.StatusInternalServerError,
			Message:   "Server Error",
			LogType:   models.LogTypeError,
			LogSource: models.LogSourceAuth,
			Request:   r,
			CtxValues: &ctxValues,
			Addendum:  "Requested by: '" + ctxValues.SecretKey + "'",
		}
		writeErrorResponse(w, config)
		return
	}

	if request.Name == "" {
		request.Name = newKeyObj.Name
	}

	models.CreateLog(models.LogTypeInfo, models.LogSourceAuth,
		"Generated a new key from IP Address: '"+r.RemoteAddr+"' with name: '"+request.Name+"'. Requested by: '"+ctxValues.SecretKey+"'")
}

type DeleteKeyRequest struct {
	Key string `json:"key"`
}

type DeleteKeyResponse struct {
	Message string `json:"message"`
}

func DeleteKeyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		config := ErrorResponseConfig{
			Status:    http.StatusMethodNotAllowed,
			Message:   "Method not allowed",
			LogType:   models.LogTypeError,
			LogSource: models.LogSourceAuth,
			Request:   r,
			CtxValues: nil,
			Addendum:  "",
		}
		writeErrorResponse(w, config)
		return
	}

	if CheckUnauthorized(w, r) {
		return // CheckUnauthorized handles its own error response
	}

	var request DeleteKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		config := ErrorResponseConfig{
			Status:    http.StatusBadRequest,
			Message:   "Invalid request body",
			LogType:   models.LogTypeError,
			LogSource: models.LogSourceAuth,
			Request:   r,
			CtxValues: nil,
			Addendum:  "",
		}
		writeErrorResponse(w, config)
		return
	}

	ctxValues, _ := GetContextValues(r)
	log.Println("Deleting key:'" + request.Key + "'. Requested by:'" + ctxValues.SecretKey + "'")

	message, err := auth.DeleteKeyByKey(request.Key)
	if err != nil {
		config := ErrorResponseConfig{
			Status:    http.StatusInternalServerError,
			Message:   err.Error(),
			LogType:   models.LogTypeError,
			LogSource: models.LogSourceAuth,
			Request:   r,
			CtxValues: &ctxValues,
		}
		writeErrorResponse(w, config)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	errorResponse := DeleteKeyResponse{Message: message}
	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		config := ErrorResponseConfig{
			Status:    http.StatusInternalServerError,
			Message:   "Server Error",
			LogType:   models.LogTypeError,
			LogSource: models.LogSourceAuth,
			Request:   r,
			CtxValues: &ctxValues,
			Addendum:  "Requested by: '" + ctxValues.SecretKey + "'",
		}
		writeErrorResponse(w, config)
		return
	}

	models.CreateLog(models.LogTypeInfo, models.LogSourceAuth,
		"Deleted key: '"+ctxValues.SecretKey+"' from IP Address: '"+r.RemoteAddr+"'. Requested by: '"+ctxValues.SecretKey+"'")
}
