package database

import (
	"fmt"
	"go-link-shortener/utils"
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

type Env struct {
	DBHost     string
	DBUser     string
	DBPassword string
	DBName     string
	DBPort     string
	DBSSLMode  string
	LOG_LEVEL  string
}

func ConnectToDatabase(env *utils.Env) *gorm.DB {
	log.Println("⏳ Connecting to Postgres database...")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		env.DBHost,
		env.DBUser,
		env.DBPassword,
		env.DBName,
		env.DBPort,
		env.DBSSLMode,
	)

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
