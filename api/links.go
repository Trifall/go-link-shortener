package api

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"go-link-shortener/database"
	"go-link-shortener/lib"
	"go-link-shortener/models"
	"log"
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
	ShortenedURL string `json:"shortened_url"`
}

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

	// fetch the ID of the secret key from the database
	log.Println("Fetching secret key ID from database...")
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

// CreateLink creates a new shortened link in the database after validating the input.
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
			return nil, errors.New("custom_url is already in use")
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

	return u.String(), nil
}

// isAlphanumeric checks if a string contains only alphanumeric characters.
func isAlphanumeric(s string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(s)
}

// isShortenedURLTaken checks if the given shortened URL already exists in the database.
func isShortenedURLTaken(db *gorm.DB, url string) (bool, error) {
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
			return "", err
		}

		exists, err := isShortenedURLTaken(db, shortURL)
		if err != nil {
			return "", err
		}
		if !exists {
			return shortURL, nil
		}
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

type RetrieveLinkRequest struct {
	Shortened string `json:"shortened"`
}

type RetrieveLinkResponse struct {
	ID            uuid.UUID         `json:"id"`
	RedirectTo    string            `json:"redirect_to"`
	Shortened     string            `json:"shortened"`
	ExpiresAt     *time.Time        `json:"expires_at"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	CreatedBy     uuid.UUID         `json:"created_by"`
	SecretKey     SecretKeyResponse `json:"secret_key"`
	Visits        int               `json:"visits"`
	LastVisitedAt *time.Time        `json:"last_visited_at"`
	IsActive      bool              `json:"is_active"`
}

type SecretKeyResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	IsActive  bool      `json:"is_active"`
}

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
		SecretKey:     ToSecretKeyResponse(l.SecretKey),
		Visits:        l.Visits,
		LastVisitedAt: l.LastVisitedAt,
		IsActive:      l.IsActive,
	}
}

// RetrieveLink, searches for a link by its shortened URL and returns the link object
func RetrieveLink(db *gorm.DB, shortened string) (*models.Link, error) {
	var link models.Link
	// Preload the SecretKey relationship
	result := db.Preload("SecretKey").Where("shortened = ?", shortened).First(&link)

	if result.Error != nil {
		return nil, result.Error
	}

	return &link, nil
}
