package models

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"go-link-shortener/lib"
	"log"
	"time"

	"gorm.io/gorm"
)

func CreateSecretKey(db *gorm.DB, name string) *string {
	log.Println("Creating secret key with name:", name, "...")
	var key string

	if name == "" {
		prefix := make([]byte, 6)
		if _, err := rand.Read(prefix); err != nil {
			return nil // Return nil if there's an error
		}
		randomPrefix := base64.URLEncoding.EncodeToString(prefix)
		name = "User " + randomPrefix
	}

	// Generate a random 32 character key
	key = base64.URLEncoding.EncodeToString(make([]byte, 32))

	secretKey := &SecretKey{
		Key:       key,
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Active:    true,
		IsAdmin:   false,
	}

	result := db.Create(secretKey)
	if result.Error != nil {
		return nil
	}

	log.Println("Secret key with name:", name, "created successfully.")
	return &key
}

func CreateRootUserKey(db *gorm.DB, key string) *string {
	log.Println("⏳ Creating Root User key...")

	secretKey := &SecretKey{
		Key:       key,
		Name:      lib.ROOT_USER_NAME,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Active:    true,
		IsAdmin:   true,
	}

	result := db.Create(secretKey)
	if result.Error != nil {
		return nil
	}

	log.Println("✔️  Root User key created successfully.")
	return &key
}

func SearchKeyByName(db *gorm.DB, name string) *SecretKey {
	var secretKey SecretKey
	result := db.Where("name = ?", name).First(&secretKey)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// No record found, return nil without logging the error
		return nil
	} else if result.Error != nil {
		log.Printf("Error querying database: %v", result.Error)
		return nil
	}

	return &secretKey
}

func SearchKeyByKey(db *gorm.DB, key string) *SecretKey {
	var secretKey SecretKey
	result := db.Where("key = ?", key).First(&secretKey)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// No record found, return nil without logging the error
		return nil
	}

	if result.Error != nil {
		log.Printf("Error querying database: %v", result.Error)
		return nil
	}

	return &secretKey
}
