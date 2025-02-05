package api

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"go-link-shortener/database"
	"go-link-shortener/lib"
	"go-link-shortener/models"
	"go-link-shortener/utils"
	"math/big"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// custom_url: The URL to be shortened, if empty, the URL will be generated
// redirect_to: The URL to redirect to after the link is shortened
// expires_at: The expiration date of the link
type ShortenRequest struct {
	CustomURL  string     `json:"custom_url"`
	RedirectTo string     `json:"redirect_to"`
	ExpiresAt  *time.Time `json:"expires_at"`
}

type ShortenResponse struct {
	ShortenedURL string `json:"shortened"`
}

// ShortenHandler shortens a URL.
// @Summary Shorten a URL
// @Description Shortens a given URL and returns the shortened version.
// @Tags links
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body ShortenRequest true "URL shortening request"
// @Success 200 {object} ShortenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 405 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/links/shorten [post]
func ShortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var request ShortenRequest
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

	if ctxValues.SecretKey == "" {
		config := ErrorResponseConfig{
			Status:    http.StatusUnauthorized,
			Message:   "Unauthorized",
			LogType:   models.LogTypeError,
			LogSource: models.LogSourceLinks,
			Request:   r,
			CtxValues: &ctxValues,
			Addendum:  "Context values not found",
		}
		writeErrorResponse(w, config)
		return
	}

	db := database.GetDB()
	if db == nil {
		config := ErrorResponseConfig{
			Status:    http.StatusInternalServerError,
			Message:   lib.ERRORS.Database,
			LogType:   models.LogTypeError,
			LogSource: models.LogSourceLinks,
			Request:   r,
			CtxValues: &ctxValues,
			Addendum:  "Requested by: '" + ctxValues.SecretKey + "'",
		}
		writeErrorResponse(w, config)
		return
	}

	secretKey := models.SearchKeyByKey(db, ctxValues.SecretKey)
	if secretKey == nil {
		config := ErrorResponseConfig{
			Status:    http.StatusInternalServerError,
			Message:   lib.ERRORS.KeyNotFound,
			LogType:   models.LogTypeError,
			LogSource: models.LogSourceLinks,
			Request:   r,
			CtxValues: &ctxValues,
			Addendum:  "Requested by: '" + ctxValues.SecretKey + "'",
		}
		writeErrorResponse(w, config)
		return
	}

	res, err := CreateLink(database.GetDB(), request, secretKey.ID)
	if err != nil {
		config := ErrorResponseConfig{
			Status:    http.StatusInternalServerError,
			Message:   err.Error(),
			LogType:   models.LogTypeError,
			LogSource: models.LogSourceLinks,
			Request:   r,
			CtxValues: &ctxValues,
			Addendum:  "Requested by: '" + ctxValues.SecretKey + "'",
		}
		writeErrorResponse(w, config)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}
}

func CreateLink(db *gorm.DB, req ShortenRequest, createdBy uuid.UUID) (*ShortenResponse, error) {
	// Validate RedirectTo
	if req.RedirectTo == "" {
		return nil, errors.New("redirect_to is required")
	}

	redirectURL, err := validateAndNormalizeURL(req.RedirectTo)
	if err != nil {
		return nil, fmt.Errorf("invalid redirect_to: %v", err)
	}

	var shortened string
	if req.CustomURL != "" {
		// Validate custom URL
		if !isAlphanumeric(req.CustomURL) {
			return nil, errors.New("custom_url must be alphanumeric")
		}

		exists, err := isShortenedURLTaken(db, req.CustomURL)
		if err != nil {
			return nil, fmt.Errorf("failed to check custom URL: %v", err)
		}
		if exists {
			return nil, errors.New("custom_url is already in use, or is reserved")
		}
		shortened = req.CustomURL
	} else {
		// Generate a unique random URL
		shortened, err = generateUniqueShortURL(db)
		if err != nil {
			return nil, fmt.Errorf("failed to generate URL: %v", err)
		}
	}

	// Create the link record
	link := models.Link{
		RedirectTo: redirectURL,
		Shortened:  shortened,
		ExpiresAt:  req.ExpiresAt,
		CreatedBy:  createdBy,
		IsActive:   true,
	}

	if err := db.Create(&link).Error; err != nil {
		return nil, fmt.Errorf("database error: %v", err)
	}

	return &ShortenResponse{ShortenedURL: shortened}, nil
}

// validateAndNormalizeURL checks and normalizes the RedirectTo URL.
func validateAndNormalizeURL(rawURL string) (string, error) {
	allowedSchemes := map[string]bool{
		"http": true, "https": true, "magnet": true, "steam": true, "spotify": true,
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		// Try prepending https:// if parsing fails
		u, err = url.Parse("https://" + rawURL)
		if err != nil {
			return "", errors.New("invalid URL format")
		}
	}

	if !allowedSchemes[u.Scheme] {
		return "", fmt.Errorf("protocol %s is not allowed", u.Scheme)
	}

	// validate that the redirect URL is not the same as the public site URL
	env := utils.LoadEnv()

	if env.PUBLIC_SITE_URL != "" {
		if u.Host == env.PUBLIC_SITE_URL {
			return "", errors.New("cannot redirect to link shortener")
		}
	}

	return u.String(), nil
}

// isAlphanumeric checks if a string contains only alphanumeric characters.
func isAlphanumeric(s string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(s)
}

// isShortenedURLTaken checks if the given shortened URL already exists in the database or if it matches any reserved routes.
func isShortenedURLTaken(db *gorm.DB, url string) (bool, error) {
	// Check if the URL matches any reserved routes
	if url == lib.RESERVED_ROUTES.API || url == lib.RESERVED_ROUTES.Docs || url == lib.RESERVED_ROUTES.NotFound {
		return true, nil // URL is a reserved route, so it's "taken"
	}

	// Proceed with the database check
	var link models.Link
	result := db.Where("shortened = ?", url).First(&link)
	if result.Error == nil {
		return true, nil // Found existing entry
	}
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return false, nil // No existing entry
	}
	return false, result.Error // Database error
}

// generateUniqueShortURL generates a random alphanumeric URL and ensures it's unique.
func generateUniqueShortURL(db *gorm.DB) (string, error) {
	const maxAttempts = 10
	for i := 0; i < maxAttempts; i++ {
		length, err := randomInt(3, 6)
		if err != nil {
			return "", err
		}

		shortURL, err := generateRandomAlphanumeric(length)

		if err != nil {
			continue
		}

		exists, err := isShortenedURLTaken(db, shortURL)
		if err != nil || exists {
			continue
		}

		return shortURL, nil
	}
	return "", errors.New("failed to generate a unique short URL after multiple attempts")
}

// generateRandomAlphanumeric generates a random alphanumeric string of given length.
func generateRandomAlphanumeric(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b), nil
}

// randomInt generates a random integer between min and max (inclusive).
func randomInt(min, max int) (int, error) {
	if min >= max {
		return 0, errors.New("invalid range for randomInt")
	}
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		return 0, err
	}
	return min + int(nBig.Int64()), nil
}

type SecretKeyResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	IsActive  bool      `json:"is_active"`
}

type RetrieveLinkRequest struct {
	Shortened string `json:"shortened"`
}

type PartialSecretKey struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type RetrieveLinkResponse struct {
	ID            uuid.UUID        `json:"id"`
	RedirectTo    string           `json:"redirect_to"`
	Shortened     string           `json:"shortened"`
	ExpiresAt     *time.Time       `json:"expires_at"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
	CreatedBy     uuid.UUID        `json:"created_by"`
	SecretKey     PartialSecretKey `json:"secret_key"`
	Visits        int              `json:"visits"`
	LastVisitedAt *time.Time       `json:"last_visited_at"`
	IsActive      bool             `json:"is_active"`
}

// RetrieveLinkHandler retrieves details of a shortened link.
// @Summary Retrieve a shortened link
// @Description Retrieves details of a shortened link by its shortened URL.
// @Tags links
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body RetrieveLinkRequest true "Link retrieval request"
// @Success 200 {object} RetrieveLinkResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 405 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/links/retrieve [post]
func RetrieveLinkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var request RetrieveLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		config := ErrorResponseConfig{
			Status:    http.StatusBadRequest,
			Message:   "Invalid request body",
			LogType:   models.LogTypeError,
			LogSource: models.LogSourceAuth,
			Request:   r,
		}
		writeErrorResponse(w, config)
		return
	}

	linkObject, err := RetrieveLink(database.GetDB(), request.Shortened)
	if err != nil {
		config := ErrorResponseConfig{
			Status:    http.StatusInternalServerError,
			Message:   err.Error(),
			LogType:   models.LogTypeError,
			LogSource: models.LogSourceLinks,
			Request:   r,
		}
		writeErrorResponse(w, config)
		return
	}

	// Convert to response struct using helper
	response := ToRetrieveLinkResponse(*linkObject)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}
}

// convert models.SecretKey to SecretKeyResponse
func ToSecretKeyResponse(s models.SecretKey) SecretKeyResponse {
	return SecretKeyResponse{
		ID:        s.ID,
		Name:      s.Name,
		CreatedAt: s.CreatedAt,
		IsActive:  s.IsActive,
	}
}

// convert models.Link to RetrieveLinkResponse
func ToRetrieveLinkResponse(l models.Link) RetrieveLinkResponse {
	return RetrieveLinkResponse{
		ID:            l.ID,
		RedirectTo:    l.RedirectTo,
		Shortened:     l.Shortened,
		ExpiresAt:     l.ExpiresAt,
		CreatedAt:     l.CreatedAt,
		UpdatedAt:     l.UpdatedAt,
		CreatedBy:     l.CreatedBy,
		SecretKey:     PartialSecretKey{Key: l.SecretKey.Key, Name: l.SecretKey.Name},
		Visits:        l.Visits,
		LastVisitedAt: l.LastVisitedAt,
		IsActive:      l.IsActive,
	}
}

// RetrieveLink, searches for a link by its shortened URL and returns the link object
func RetrieveLink(db *gorm.DB, shortened string) (*models.Link, error) {
	var link models.Link
	// preload the SecretKey relationship
	result := db.Preload("SecretKey").Where("shortened = ?", shortened).First(&link)

	if result.Error != nil {
		return nil, result.Error
	}

	return &link, nil
}

func RetrieveRedirectURL(db *gorm.DB, shortened string) (*models.Link, error) {
	var link models.Link
	result := db.Where("shortened = ? AND is_active = ?", shortened, true).First(&link)

	if result.Error != nil {
		return nil, result.Error
	}

	if link.RedirectTo == "" {
		return nil, errors.New("invalid redirect")
	}

	return &link, nil
}

type DeleteLinkRequest struct {
	Shortened string `json:"shortened"`
}

type DeleteLinkResponse struct {
	Message string `json:"message"`
}

// DeleteLinkHandler deletes a shortened link.
// @Summary Delete a shortened link
// @Description Deletes a shortened link by its shortened URL.
// @Tags links
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body DeleteLinkRequest true "Link deletion request"
// @Success 200 {object} DeleteLinkResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 405 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/links/delete [post]
func DeleteLinkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var request DeleteLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		config := ErrorResponseConfig{
			Status:    http.StatusBadRequest,
			Message:   "Invalid request body",
			LogType:   models.LogTypeError,
			LogSource: models.LogSourceLinks,
			Request:   r,
		}
		writeErrorResponse(w, config)
		return
	}

	ctxValues, _ := GetContextValues(r)
	db := database.GetDB()

	// retrieve the link to be deleted
	link, err := RetrieveLink(db, request.Shortened)
	if err != nil {
		config := ErrorResponseConfig{
			Status:    http.StatusNotFound,
			Message:   "Link not found",
			LogType:   models.LogTypeError,
			LogSource: models.LogSourceLinks,
			Request:   r,
			CtxValues: &ctxValues,
			Addendum:  fmt.Sprintf("Shortened URL: %s", request.Shortened),
		}
		writeErrorResponse(w, config)
		return
	}

	// autho
	if !ctxValues.IsAdmin {
		// check ownership
		secretKey := models.SearchKeyByKey(db, ctxValues.SecretKey)
		if secretKey == nil {
			config := ErrorResponseConfig{
				Status:    http.StatusInternalServerError,
				Message:   lib.ERRORS.KeyNotFound,
				LogType:   models.LogTypeError,
				LogSource: models.LogSourceLinks,
				Request:   r,
				CtxValues: &ctxValues,
				Addendum:  "Secret key not found in database",
			}
			writeErrorResponse(w, config)
			return
		}

		if secretKey.ID != link.CreatedBy {
			config := ErrorResponseConfig{
				Status:    http.StatusUnauthorized,
				Message:   "Unauthorized to delete this link",
				LogType:   models.LogTypeWarning,
				LogSource: models.LogSourceLinks,
				Request:   r,
				CtxValues: &ctxValues,
				Addendum:  fmt.Sprintf("User ID: %s, Link Creator: %s", secretKey.ID, link.CreatedBy),
			}
			writeErrorResponse(w, config)
			return
		}
	}

	// delete link
	if err := db.Delete(&link).Error; err != nil {
		config := ErrorResponseConfig{
			Status:    http.StatusInternalServerError,
			Message:   "Failed to delete link",
			LogType:   models.LogTypeError,
			LogSource: models.LogSourceLinks,
			Request:   r,
			CtxValues: &ctxValues,
			Addendum:  fmt.Sprintf("Error: %v", err),
		}
		writeErrorResponse(w, config)
		return
	}

	// Success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Link deleted successfully",
	})
}

type UpdateLinkRequest struct {
	Shortened    string     `json:"shortened"`
	RedirectTo   *string    `json:"redirect_to,omitempty"`
	NewShortened *string    `json:"new_shortened,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	IsActive     *bool      `json:"is_active,omitempty"`
}

type UpdateLinkResponse struct {
	ID            uuid.UUID  `json:"id"`
	RedirectTo    string     `json:"redirect_to"`
	Shortened     string     `json:"shortened"`
	ExpiresAt     *time.Time `json:"expires_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	CreatedBy     uuid.UUID  `json:"created_by"`
	Visits        int        `json:"visits"`
	LastVisitedAt *time.Time `json:"last_visited_at"`
	IsActive      bool       `json:"is_active"`
}

// UpdateLinkHandler updates a shortened link.
// @Summary Update a shortened link
// @Description Updates a shortened link with new values for redirect URL, shortened URL, expiration date, or active status.
// @Tags links
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body UpdateLinkRequest true "Link update request"
// @Success 200 {object} UpdateLinkResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 405 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/links/update [post]
func UpdateLinkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var request UpdateLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		config := ErrorResponseConfig{
			Status:    http.StatusBadRequest,
			Message:   "Invalid request body",
			LogType:   models.LogTypeError,
			LogSource: models.LogSourceLinks,
			Request:   r,
		}
		writeErrorResponse(w, config)
		return
	}

	ctxValues, _ := GetContextValues(r)
	db := database.GetDB()

	// retrieve existing link
	link, err := RetrieveLink(db, request.Shortened)
	if err != nil {
		config := ErrorResponseConfig{
			Status:    http.StatusNotFound,
			Message:   "Link not found",
			LogType:   models.LogTypeError,
			LogSource: models.LogSourceLinks,
			Request:   r,
			CtxValues: &ctxValues,
			Addendum:  fmt.Sprintf("Shortened: %s", request.Shortened),
		}
		writeErrorResponse(w, config)
		return
	}

	// auth check
	if !ctxValues.IsAdmin {
		secretKey := models.SearchKeyByKey(db, ctxValues.SecretKey)
		if secretKey == nil || secretKey.ID != link.CreatedBy {
			config := ErrorResponseConfig{
				Status:    http.StatusUnauthorized,
				Message:   "Unauthorized to update this link",
				LogType:   models.LogTypeWarning,
				LogSource: models.LogSourceLinks,
				Request:   r,
				CtxValues: &ctxValues,
			}
			writeErrorResponse(w, config)
			return
		}
	}

	// validate and apply updates
	if request.RedirectTo != nil {
		normalized, err := validateAndNormalizeURL(*request.RedirectTo)
		if err != nil {
			config := ErrorResponseConfig{
				Status:    http.StatusBadRequest,
				Message:   fmt.Sprintf("Invalid redirect URL: %v", err),
				LogType:   models.LogTypeError,
				LogSource: models.LogSourceLinks,
				Request:   r,
				CtxValues: &ctxValues,
			}
			writeErrorResponse(w, config)
			return
		}
		link.RedirectTo = normalized
	}

	if request.NewShortened != nil {
		newShort := *request.NewShortened
		if newShort != link.Shortened {
			if !isAlphanumeric(newShort) {
				config := ErrorResponseConfig{
					Status:    http.StatusBadRequest,
					Message:   "New shortened URL must be alphanumeric",
					LogType:   models.LogTypeError,
					LogSource: models.LogSourceLinks,
					Request:   r,
					CtxValues: &ctxValues,
				}
				writeErrorResponse(w, config)
				return
			}

			exists, err := isShortenedURLTaken(db, newShort)
			if err != nil {
				config := ErrorResponseConfig{
					Status:    http.StatusInternalServerError,
					Message:   "Failed to validate new shortened URL",
					LogType:   models.LogTypeError,
					LogSource: models.LogSourceLinks,
					Request:   r,
					CtxValues: &ctxValues,
				}
				writeErrorResponse(w, config)
				return
			}
			if exists {
				config := ErrorResponseConfig{
					Status:    http.StatusConflict,
					Message:   "New shortened URL already exists",
					LogType:   models.LogTypeError,
					LogSource: models.LogSourceLinks,
					Request:   r,
					CtxValues: &ctxValues,
				}
				writeErrorResponse(w, config)
				return
			}
			link.Shortened = newShort
		}
	}

	if request.ExpiresAt != nil {
		// define the earliest possible timestamp
		earliest, err := time.Parse(time.RFC3339, "1970-01-01T00:00:00.000Z")
		if err != nil {
			// if parsing fails, fall back to using the provided value.
			link.ExpiresAt = request.ExpiresAt
		} else {
			// if the provided expires_at equals the earliest possible timestamp, unset it.
			if request.ExpiresAt.Equal(earliest) {
				link.ExpiresAt = nil
			} else {
				link.ExpiresAt = request.ExpiresAt
			}
		}
	}

	if request.IsActive != nil {
		link.IsActive = *request.IsActive
	}

	// Save updates
	if err := db.Save(&link).Error; err != nil {
		config := ErrorResponseConfig{
			Status:    http.StatusInternalServerError,
			Message:   "Failed to update link",
			LogType:   models.LogTypeError,
			LogSource: models.LogSourceLinks,
			Request:   r,
			CtxValues: &ctxValues,
			Addendum:  fmt.Sprintf("Error: %v", err),
		}
		writeErrorResponse(w, config)
		return
	}

	// Return updated link
	response := ToRetrieveLinkResponse(*link)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

type RetrieveAllLinksResponse struct {
	Message string                 `json:"message"`
	Links   []RetrieveLinkResponse `json:"links"`
}

// RetrieveAllLinksHandler retrieves all shortened links.
// @Summary Retrieve all shortened links
// @Description Retrieves all shortened links from the database.
// @Tags links,admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} RetrieveAllLinksResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 405 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/links/retrieve-all [get]
func RetrieveAllLinksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		config := ErrorResponseConfig{
			Status:    http.StatusMethodNotAllowed,
			Message:   "Method not allowed",
			LogType:   models.LogTypeError,
			LogSource: models.LogSourceLinks,
			Request:   r,
			CtxValues: nil,
			Addendum:  "",
		}
		writeErrorResponse(w, config)
		return
	}

	if CheckUnauthorized(w, r) {
		return
	}

	ctxValues, _ := GetContextValues(r)

	db := database.GetDB()

	if db == nil {
		config := ErrorResponseConfig{
			Status:    http.StatusInternalServerError,
			Message:   "Database Error",
			LogType:   models.LogTypeError,
			LogSource: models.LogSourceLinks,
			Request:   r,
			CtxValues: &ctxValues,
			Addendum:  "Requested by: '" + ctxValues.SecretKey + "'",
		}
		writeErrorResponse(w, config)
		return
	}

	// Retrieve all links from the database
	links := models.RetrieveAllLinks(db)

	// Convert each link to RetrieveLinkResponse
	var responseLinks []RetrieveLinkResponse
	for _, link := range links {
		responseLinks = append(responseLinks, ToRetrieveLinkResponse(link))
	}

	response := RetrieveAllLinksResponse{
		Message: "Links retrieved successfully",
		Links:   responseLinks,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		config := ErrorResponseConfig{
			Status:    http.StatusInternalServerError,
			Message:   "Server Error",
			LogType:   models.LogTypeError,
			LogSource: models.LogSourceLinks,
			Request:   r,
			CtxValues: &ctxValues,
			Addendum:  "Requested by: '" + ctxValues.SecretKey + "'",
		}
		writeErrorResponse(w, config)
		return
	}

	models.CreateLog(models.LogTypeInfo, models.LogSourceLinks,
		"Retrieved all links. Requested by: '"+ctxValues.SecretKey+"'", r.RemoteAddr)
}

type RetrieveAllLinksByKeyRequest struct {
	Key string `json:"key"`
}

type RetrieveAllLinksByKeyResponse struct {
	Message string                 `json:"message"`
	Links   []RetrieveLinkResponse `json:"links"`
}

// RetrieveAllLinksByKeyHandler retrieves all shortened links by a secret key.
// @Summary Retrieve all shortened links by a secret key
// @Description Retrieves all shortened links by a secret key from the database.
// @Tags links
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body RetrieveAllLinksByKeyRequest true "Secret key retrieval request"
// @Success 200 {object} RetrieveAllLinksByKeyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 405 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/links/retrieve-all-by-key [post]
func RetrieveAllLinksByKeyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		config := ErrorResponseConfig{
			Status:    http.StatusMethodNotAllowed,
			Message:   "Method not allowed",
			LogType:   models.LogTypeError,
			LogSource: models.LogSourceLinks,
			Request:   r,
			CtxValues: nil,
			Addendum:  "",
		}
		writeErrorResponse(w, config)
		return
	}

	if CheckUnauthorized(w, r) {
		return
	}

	var request RetrieveAllLinksByKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		config := ErrorResponseConfig{
			Status:    http.StatusBadRequest,
			Message:   "Invalid request body",
			LogType:   models.LogTypeError,
			LogSource: models.LogSourceLinks,
			Request:   r,
			CtxValues: nil,
			Addendum:  "",
		}
		writeErrorResponse(w, config)
		return
	}

	ctxValues, _ := GetContextValues(r)

	db := database.GetDB()

	if db == nil {
		config := ErrorResponseConfig{
			Status:    http.StatusInternalServerError,
			Message:   "Database Error",
			LogType:   models.LogTypeError,
			LogSource: models.LogSourceLinks,
			Request:   r,
			CtxValues: &ctxValues,
			Addendum:  "Requested by: '" + ctxValues.SecretKey + "'",
		}
		writeErrorResponse(w, config)
		return
	}

	// auth check
	if !ctxValues.IsAdmin {
		// verify that the secret key matches the key in the request
		if request.Key != ctxValues.SecretKey {
			config := ErrorResponseConfig{
				Status:    http.StatusUnauthorized,
				Message:   "Unauthorized to retrieve links by key",
				LogType:   models.LogTypeWarning,
				LogSource: models.LogSourceLinks,
				Request:   r,
				CtxValues: &ctxValues,
				Addendum:  "Requested by: '" + ctxValues.SecretKey + "'",
			}
			writeErrorResponse(w, config)
			return
		}
	}

	// retrieve all links by the secret key
	links, err := models.RetrieveAllLinksByKey(db, request.Key)
	if err != nil {
		config := ErrorResponseConfig{
			Status:    http.StatusInternalServerError,
			Message:   "Database Error",
			LogType:   models.LogTypeError,
			LogSource: models.LogSourceLinks,
			Request:   r,
			CtxValues: &ctxValues,
			Addendum:  "Requested by: '" + ctxValues.SecretKey + "'",
		}
		writeErrorResponse(w, config)
		return
	}

	// Convert to response struct
	response := RetrieveAllLinksByKeyResponse{
		Message: "Links retrieved successfully",
		Links:   ToRetrieveLinkResponses(links),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		config := ErrorResponseConfig{
			Status:    http.StatusInternalServerError,
			Message:   "Server Error",
			LogType:   models.LogTypeError,
			LogSource: models.LogSourceLinks,
			Request:   r,
			CtxValues: &ctxValues,
			Addendum:  "Requested by: '" + ctxValues.SecretKey + "'",
		}
		writeErrorResponse(w, config)
		return
	}

	models.CreateLog(models.LogTypeInfo, models.LogSourceLinks,
		"Retrieved all links for key: '"+request.Key+"'. Requested by: '"+ctxValues.SecretKey+"'", r.RemoteAddr)
}

func ToRetrieveLinkResponses(links []models.Link) []RetrieveLinkResponse {
	var responseLinks []RetrieveLinkResponse
	for _, link := range links {
		responseLinks = append(responseLinks, ToRetrieveLinkResponse(link))
	}
	return responseLinks
}
