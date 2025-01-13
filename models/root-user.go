package models

import (
	"go-link-shortener/lib"
	"log"

	"gorm.io/gorm"
)

func InitializeRootUser(db *gorm.DB, rootUserKey string) {
	if SearchKeyByName(db, lib.ROOT_USER_NAME) == nil {
		log.Println("⏳ No Root User detected, loading from .env...")
		// Load root user key from environment variable
		if rootUserKey == "" {
			log.Fatal("Error: ROOT_USER_KEY environment variable is not set")
		}

		// create secret key for Root User
		rootUserKey = *CreateRootUserKey(db, rootUserKey)
		if rootUserKey == "" {
			log.Fatal("Error creating Root User key")
		}
	}
}
