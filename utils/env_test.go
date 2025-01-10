package utils

import (
	"bufio"
	"os"
	"strings"
	"testing"
)

func TestLoadEnv(t *testing.T) {
	// Backup the original environment variables
	originalEnv := make(map[string]string)
	for _, envVar := range []string{"DB_HOST", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_PORT", "DB_SSLMODE", "LOG_LEVEL", "ROOT_USER_KEY"} {
		originalEnv[envVar] = os.Getenv(envVar)
	}

	// Clean up environment variables after the test
	defer func() {
		for key, value := range originalEnv {
			os.Setenv(key, value)
		}
	}()

	// Read the .env.example file
	file, err := os.Open("../.env.example")
	if err != nil {
		t.Fatalf("Failed to open .env.example file: %v", err)
	}
	defer file.Close()

	expectedEnv := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue // Skip comments and empty lines
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // Skip malformed lines
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		expectedEnv[key] = value
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Failed to read .env.example file: %v", err)
	}

	// Set the environment variables from .env.example
	for key, value := range expectedEnv {
		os.Setenv(key, value)
	}

	t.Run("Successful loading of environment variables", func(t *testing.T) {
		env := LoadEnv()

		// Verify that the environment variables are correctly loaded
		if env.DBHost != expectedEnv["DB_HOST"] {
			t.Errorf("Expected DBHost to be '%s', got '%s'", expectedEnv["DB_HOST"], env.DBHost)
		}
		if env.DBUser != expectedEnv["DB_USER"] {
			t.Errorf("Expected DBUser to be '%s', got '%s'", expectedEnv["DB_USER"], env.DBUser)
		}
		if env.DBPassword != expectedEnv["DB_PASSWORD"] {
			t.Errorf("Expected DBPassword to be '%s', got '%s'", expectedEnv["DB_PASSWORD"], env.DBPassword)
		}
		if env.DBName != expectedEnv["DB_NAME"] {
			t.Errorf("Expected DBName to be '%s', got '%s'", expectedEnv["DB_NAME"], env.DBName)
		}
		if env.DBPort != expectedEnv["DB_PORT"] {
			t.Errorf("Expected DBPort to be '%s', got '%s'", expectedEnv["DB_PORT"], env.DBPort)
		}
		if env.DBSSLMode != expectedEnv["DB_SSLMODE"] {
			t.Errorf("Expected DBSSLMode to be '%s', got '%s'", expectedEnv["DB_SSLMODE"], env.DBSSLMode)
		}
		if env.LOG_LEVEL != expectedEnv["LOG_LEVEL"] {
			t.Errorf("Expected LOG_LEVEL to be '%s', got '%s'", expectedEnv["LOG_LEVEL"], env.LOG_LEVEL)
		}
		if env.ROOT_USER_KEY != expectedEnv["ROOT_USER_KEY"] {
			t.Errorf("Expected ROOT_USER_KEY to be '%s', got '%s'", expectedEnv["ROOT_USER_KEY"], env.ROOT_USER_KEY)
		}
	})

	t.Run("Missing required environment variables", func(t *testing.T) {
		// Clear all environment variables
		for _, envVar := range []string{"DB_HOST", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_PORT", "DB_SSLMODE"} {
			os.Setenv(envVar, "")
		}

		// Expect the function to panic due to missing required environment variables
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected LoadEnv to panic due to missing required environment variables")
			}
		}()

		LoadEnv()
	})

	t.Run("Optional environment variables", func(t *testing.T) {
		// Set up required environment variables
		os.Setenv("DB_HOST", expectedEnv["DB_HOST"])
		os.Setenv("DB_USER", expectedEnv["DB_USER"])
		os.Setenv("DB_PASSWORD", expectedEnv["DB_PASSWORD"])
		os.Setenv("DB_NAME", expectedEnv["DB_NAME"])
		os.Setenv("DB_PORT", expectedEnv["DB_PORT"])
		os.Setenv("DB_SSLMODE", expectedEnv["DB_SSLMODE"])

		// Optional environment variables are not set
		os.Setenv("LOG_LEVEL", "")
		os.Setenv("ROOT_USER_KEY", "")

		env := LoadEnv()

		// Verify that the optional environment variables are empty
		if env.LOG_LEVEL != "" {
			t.Errorf("Expected LOG_LEVEL to be empty, got '%s'", env.LOG_LEVEL)
		}
		if env.ROOT_USER_KEY != "" {
			t.Errorf("Expected ROOT_USER_KEY to be empty, got '%s'", env.ROOT_USER_KEY)
		}
	})
}
