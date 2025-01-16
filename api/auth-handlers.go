package api

import (
	"encoding/json"
	"go-link-shortener/auth"
	"go-link-shortener/lib"
	"go-link-shortener/models"
	"go-link-shortener/utils"
	"log"
	"net/http"
)

type ValidateKeyResponse struct {
	Message string      `json:"message"`
	Key     StrippedKey `json:"key"`
}

type ContextValues struct {
	SecretKey string
	IsAdmin   bool
}

// ValidateKeyHandler validates a secret key.
// @Summary Validate a secret key
// @Description Validates a secret key and returns its details if valid.
// @Tags auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} ValidateKeyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 405 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/keys/validate [post]
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
	log.Println("Validate Key Request:'"+ctxValues.SecretKey+"', IsAdmin:", ctxValues.IsAdmin)

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
		Message: "Key validated successfully",
		Key: StrippedKey{
			Key:       keyObj.Key,
			Name:      keyObj.Name,
			CreatedAt: utils.SafeString(&keyObj.CreatedAt),
			UpdatedAt: utils.SafeString(&keyObj.UpdatedAt),
			IsActive:  keyObj.IsActive,
			IsAdmin:   keyObj.IsAdmin,
		},
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
	Message string      `json:"message"`
	Key     StrippedKey `json:"key"`
}

// GenerateKeyHandler generates a new secret key.
// @Summary Generate a new secret key
// @Description Generates a new secret key with the specified name and admin status.
// @Tags auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body GenerateKeyRequest true "Key generation request"
// @Success 200 {object} GenerateKeyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 405 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/keys/generate [post]
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
	log.Println("Generate Key Request with name:'"+request.Name+"', IsAdmin:", request.IsAdmin, ". Requested by: '"+ctxValues.SecretKey+"'")

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
		Message: "Key generated successfully",
		Key: StrippedKey{
			Key:       newKeyObj.Key,
			Name:      newKeyObj.Name,
			CreatedAt: utils.SafeString(&newKeyObj.CreatedAt),
			UpdatedAt: utils.SafeString(&newKeyObj.UpdatedAt),
			IsActive:  newKeyObj.IsActive,
			IsAdmin:   newKeyObj.IsAdmin,
		},
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

// DeleteKeyHandler deletes a secret key.
// @Summary Delete a secret key
// @Description Deletes a secret key by its value.
// @Tags auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body DeleteKeyRequest true "Key deletion request"
// @Success 200 {object} DeleteKeyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 405 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/keys/delete [post]
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
	log.Println("Delete Key Request:'" + request.Key + "'. Requested by:'" + ctxValues.SecretKey + "'")

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

type UpdateKeyRequest struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	IsAdmin  *bool  `json:"is_admin"`
	IsActive *bool  `json:"is_active"`
}

type UpdateKeyResponse struct {
	Message string       `json:"message"`
	Key     *StrippedKey `json:"key"`
}

type StrippedKey struct {
	Key       string `json:"key"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	IsActive  bool   `json:"is_active"`
	IsAdmin   bool   `json:"is_admin"`
}

func buildUpdateRequest(req UpdateKeyRequest) auth.UpdateKeyS {
	var updateReq auth.UpdateKeyS

	if req.Key != "" {
		updateReq.Key = &req.Key
	}
	if req.Name != "" {
		updateReq.Name = &req.Name
	}
	if req.IsActive != nil {
		updateReq.IsActive = req.IsActive
	}
	if req.IsActive != nil {
		updateReq.IsAdmin = req.IsAdmin
	}

	return updateReq
}

// UpdateKeyHandler updates an existing secret key.
// @Summary Update a secret key
// @Description Updates an existing secret key with new values for name, admin status, or active status.
// @Tags auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body UpdateKeyRequest true "Key update request"
// @Success 200 {object} UpdateKeyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 405 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/keys/update [post]
func UpdateKeyHandler(w http.ResponseWriter, r *http.Request) {
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

	var request UpdateKeyRequest
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
	log.Println("Update Key Request:'"+request.Key+"', Name:'"+request.Name+"', IsAdmin:", request.IsAdmin, ", IsActive:", request.IsActive, ". Requested by:'"+ctxValues.SecretKey+"'")

	updateRequest := buildUpdateRequest(request)

	message, updatedKeyObj, err := auth.UpdateKey(updateRequest)
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

		if err.Error() == lib.ERRORS.NoNewFields {
			config.Status = http.StatusBadRequest
		}

		writeErrorResponse(w, config)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Prepare and send response
	response := UpdateKeyResponse{
		Message: message,
		Key: &StrippedKey{
			Key:       updatedKeyObj.Key,
			Name:      updatedKeyObj.Name,
			CreatedAt: utils.SafeString(&updatedKeyObj.CreatedAt),
			UpdatedAt: utils.SafeString(&updatedKeyObj.UpdatedAt),
			IsActive:  updatedKeyObj.IsActive,
			IsAdmin:   updatedKeyObj.IsAdmin,
		},
	}

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

	models.CreateLog(models.LogTypeInfo, models.LogSourceAuth, "Updated key: '"+request.Key+"' from IP Address: '"+r.RemoteAddr+"'. Requested by: '"+ctxValues.SecretKey+"'")
}
