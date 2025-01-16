package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Create a new link
func CreateLink(db *gorm.DB, redirectTo string, shortened string, expiresAt *time.Time, createdBy uuid.UUID) (*Link, error) {
	link := &Link{
		RedirectTo: redirectTo,
		Shortened:  shortened,
		ExpiresAt:  expiresAt,
		CreatedBy:  createdBy,
	}

	result := db.Create(link)
	if result.Error != nil {
		return nil, result.Error
	}

	return link, nil
}

// Find an active link by shortened URL
func FindActiveLink(db *gorm.DB, shortened string) (*Link, error) {
	var link Link
	result := db.Where("shortened = ? AND is_active = ? AND (expires_at IS NULL OR expires_at > ?)",
		shortened, true, time.Now()).First(&link)

	if result.Error != nil {
		return nil, result.Error
	}

	return &link, nil
}
