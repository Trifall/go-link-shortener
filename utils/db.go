package utils

import (
	"go-link-shortener/lib"
	"go-link-shortener/models"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func SetDB(database *gorm.DB) {
	db = database
}

func GetDB() *gorm.DB {
	return db
}

func ConnectToDatabase(env *Env) *gorm.DB {
	log.Println("⏳ Connecting to Postgres database...")

	dsn := "host=" + env.DBHost +
		" user=" + env.DBUser +
		" password=" + env.DBPassword +
		" dbname=" + env.DBName +
		" port=" + env.DBPort +
		" sslmode=" + env.DBSSLMode +
		" TimeZone=UTC"

	loggerMode := logger.Silent
	loggerStrVal := "Silent"
	if env.LOG_LEVEL == "debug" || env.LOG_LEVEL == "info" {
		loggerMode = logger.Info
		loggerStrVal = "Info"
	} else if env.LOG_LEVEL == "warn" {
		loggerMode = logger.Warn
		loggerStrVal = "Warn"
	} else if env.LOG_LEVEL == "error" {
		loggerMode = logger.Error
		loggerStrVal = "Error"
	}

	log.Println("🛈  GORM Logging Mode:", loggerStrVal)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(loggerMode),
	})
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func InitializeRootUser(db *gorm.DB, rootUserKey string) {
	if models.SearchKeyByName(db, lib.ROOT_USER_NAME) == nil {
		log.Println("⏳ No Root User detected, loading from .env...")
		// Load root user key from environment variable
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
