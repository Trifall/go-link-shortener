package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Env struct {
	DBHost        string
	DBUser        string
	DBPassword    string
	DBName        string
	DBPort        string
	DBSSLMode     string
	ROOT_USER_KEY string
	LOG_LEVEL     string
}

func CheckTestEnvironment() bool {
	_, ok := os.LookupEnv("ENVIRONMENT")
	return ok
}

func LoadEnv() *Env {
	log.Println("⏳ Loading environment variables...")

	ok := CheckTestEnvironment()

	var err error
	if ok {
		err = godotenv.Load("../.env.example")
	} else {
		err = godotenv.Load()
	}

	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// verify that all required environment variables are set
	requiredEnvVars := []string{"DB_HOST", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_PORT", "DB_SSLMODE"}
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			if ok {
				panic("Missing required environment variables")
			}
			log.Fatalf("Error: %s environment variable is not set", envVar)
		}
	}

	env := Env{
		DBHost:        os.Getenv("DB_HOST"),
		DBUser:        os.Getenv("DB_USER"),
		DBPassword:    os.Getenv("DB_PASSWORD"),
		DBName:        os.Getenv("DB_NAME"),
		DBPort:        os.Getenv("DB_PORT"),
		DBSSLMode:     os.Getenv("DB_SSLMODE"),
		LOG_LEVEL:     os.Getenv("LOG_LEVEL"),
		ROOT_USER_KEY: os.Getenv("ROOT_USER_KEY"),
	}

	log.Println("✔️  Environment variables loaded successfully.")
	return &env
}
