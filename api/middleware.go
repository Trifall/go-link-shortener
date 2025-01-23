package api

import (
	"context"
	"go-link-shortener/auth"
	"go-link-shortener/database"
	"go-link-shortener/lib"
	"go-link-shortener/models"
	"net/http"
)

type ContextKey string

const secretKeyContextKey ContextKey = "secret_key"

func LogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		models.CreateLog(models.LogTypeInfo, models.LogSourceRequest,
			"Request for "+r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// extract secret key from the request
		secretKey := r.Header.Get("Authorization")
		if secretKey == "" {
			config := ErrorResponseConfig{
				Status:    http.StatusUnauthorized,
				Message:   "Authorization header required",
				LogType:   models.LogTypeError,
				LogSource: models.LogSourceAuth,
				Request:   r,
				CtxValues: nil,
				Addendum:  "Requested by: " + secretKey,
			}
			writeErrorResponse(w, config)
			return
		}

		// Validate the key and get the key object
		keyObj, err := auth.ValidateKey(secretKey)
		if err != nil {
			config := ErrorResponseConfig{
				Status:    http.StatusUnauthorized,
				Message:   "Error: " + err.Error(),
				LogType:   models.LogTypeError,
				LogSource: models.LogSourceAuth,
				Request:   r,
				CtxValues: nil,
				Addendum:  "Requested by: " + secretKey,
			}
			writeErrorResponse(w, config)
			return
		}

		// Save the updated key object back to the database
		db := database.GetDB()
		if db == nil {
			config := ErrorResponseConfig{
				Status:    http.StatusInternalServerError,
				Message:   lib.ERRORS.Database,
				LogType:   models.LogTypeError,
				LogSource: models.LogSourceAuth,
				Request:   r,
				CtxValues: nil,
				Addendum:  "Requested by: " + secretKey,
			}
			writeErrorResponse(w, config)
			return
		}
		db.Save(&keyObj)

		// Attach the key object to the request context
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
		ctxValues, ok := r.Context().Value(secretKeyContextKey).(ContextValues)
		if !ok {
			config := ErrorResponseConfig{
				Status:    http.StatusUnauthorized,
				Message:   "Unauthorized",
				LogType:   models.LogTypeError,
				LogSource: models.LogSourceAuth,
				Request:   r,
				CtxValues: nil,
				Addendum:  "Context values not found",
			}
			writeErrorResponse(w, config)
			return
		}

		if !ctxValues.IsAdmin {
			config := ErrorResponseConfig{
				Status:    http.StatusForbidden,
				Message:   "Forbidden: Admin access required",
				LogType:   models.LogTypeError,
				LogSource: models.LogSourceAuth,
				Request:   r,
				CtxValues: &ctxValues,
				Addendum:  "Requested by: '" + ctxValues.SecretKey + "'",
			}
			writeErrorResponse(w, config)
			return
		}

		next.ServeHTTP(w, r)
	})
}
