package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

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

func RetrieveAllLinks(db *gorm.DB) []Link {
	var links []Link
	db.Find(&links)
	return links
}

func RetrieveAllLinksByKey(db *gorm.DB, key string) ([]Link, error) {
	// retrieve the UUID associated with the key
	var secretKey SecretKey
	if err := db.Where("key = ?", key).First(&secretKey).Error; err != nil {
		return nil, errors.New("key not found")
	}

	var links []Link
	if err := db.Where("created_by = ?", secretKey.ID).Find(&links).Error; err != nil {
		return nil, errors.New("failed to retrieve links by key")
	}

	return links, nil
}
