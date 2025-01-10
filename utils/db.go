package utils

import (
	"go-link-shortener/models"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectToDatabase(env *Env) *gorm.DB {
	log.Println("⏳ Connecting to postgres database...")

	dsn := "host=" + env.DBHost +
		" user=" + env.DBUser +
		" password=" + env.DBPassword +
		" dbname=" + env.DBName +
		" port=" + env.DBPort +
		" sslmode=" + env.DBSSLMode +
		" TimeZone=UTC"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func InitializeRootUser(db *gorm.DB, rootUserKey string) {
	if models.CheckKeyByName(db, "Root User") == nil {
		log.Println("⏳ No Root User detected, loading from .env...")
		// Load root user key from environment variable
		rootUserKey := os.Getenv("ROOT_USER_KEY")
		if rootUserKey == "" {
			log.Fatal("Error: ROOT_USER_KEY environment variable is not set")
		}

		// create secret key for Root User
		rootUserKey = *models.CreateRootUserKey(db, rootUserKey)
		if rootUserKey == "" {
			log.Fatal("Error creating Root User key")
		}
	}
}
