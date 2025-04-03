package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Env struct {
	DBHost          string
	DBUser          string
	DBPassword      string
	DBName          string
	DBPort          string
	DBSSLMode       string
	ROOT_USER_KEY   string
	LOG_LEVEL       string
	PUBLIC_SITE_URL string
	ENABLE_DOCS     string
	SERVER_PORT     string
}

func CheckTestEnvironment() bool {
	_, ok := os.LookupEnv("ENVIRONMENT")
	return ok
}

func CheckLocalEnvironment() bool {
	_, ok := os.LookupEnv("LOCAL_BUILD")
	return ok
}

// Make Global ENV available
var ENV *Env

func LoadEnv() *Env {
	isTestMode := CheckTestEnvironment()
	isLocalMode := CheckLocalEnvironment()

	var err error
	if isTestMode {
		err = godotenv.Load("../.env.example")
		if err != nil {
			log.Panicf("Error: Couldn't load env example in test mode. Make sure to create a .env.example file and set the required environment variables.")
		}
	} else if isLocalMode {
		err = godotenv.Load()
		if err != nil {
			log.Panicf("Error: Couldn't load env in local mode. Make sure to create a .env file and set the required environment variables.")
		}
	}

	dbPort := os.Getenv("DB_PORT")

	// if not in local or test mode, set the the port to 5432
	if !isLocalMode && !isTestMode {
		dbPort = "5432"
	}

	serverPort := os.Getenv("SERVER_PORT")

	if serverPort == "" {
		log.Println("🛈  Setting server port to default: 8080")
		serverPort = "8080"
	}

	env := Env{
		DBHost:          os.Getenv("DB_HOST"),
		DBUser:          os.Getenv("DB_USER"),
		DBPassword:      os.Getenv("DB_PASSWORD"),
		DBName:          os.Getenv("DB_NAME"),
		DBPort:          dbPort,
		DBSSLMode:       os.Getenv("DB_SSLMODE"),
		LOG_LEVEL:       os.Getenv("LOG_LEVEL"),
		ROOT_USER_KEY:   os.Getenv("ROOT_USER_KEY"),
		PUBLIC_SITE_URL: os.Getenv("PUBLIC_SITE_URL"),
		ENABLE_DOCS:     os.Getenv("ENABLE_DOCS"),
		SERVER_PORT:     os.Getenv("SERVER_PORT"),
	}

	// verify that all required environment variables are set
	requiredEnvVars := []string{"DB_HOST", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_PORT", "DB_SSLMODE", "ROOT_USER_KEY"}
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			log.Panicf("Error: %s environment variable is not set", envVar)
		}
	}

	ENV = &env
	return &env
}
