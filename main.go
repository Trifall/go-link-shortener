package main

import (
	"context"
	"log"

	"go-link-shortener/database"
	_ "go-link-shortener/docs"
	"go-link-shortener/models"
	"go-link-shortener/utils"
	"go-link-shortener/workers"
)

// @title Jerren's Link Shortener API
// @version 1.0
// @description ## This is the API documentation for Jerren's Link Shortener.
// @description
// @description # Important Note
// @description **All routes are prefixed with `/api`**. For example:
// @description - `/shorten` is actually accessed at `/api/shorten`
// @description - `/v1/keys/validate` is actually accessed at `/api/v1/keys/validate`
// @description
// @description # Authentication
// @description Most endpoints require API key authentication via the Authorization header.
// @description If you `are` the system administrator, you can use the `root` key to access the API and create more keys (or make your own key).
// @description If you `are not` the system administrator, you will need to acquire a key from the system administrator.
//
// @contact.name Jerren
// @contact.url https://trifall.com
// @contact.email jerren@trifall.com

// @license.name License: Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath /api

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description API key authentication. Add 'Authorization' header with your API key.
func main() {
	log.Println("Starting Link Shortener")

	log.Println("⏳ Loading environment variables...")
	env := utils.LoadEnv()
	log.Println("✔️  Environment variables loaded successfully.")

	database.SetDB(database.ConnectToDatabase(env))

	// Setup database
	if err := models.SetupDatabase(database.GetDB()); err != nil {
		log.Fatal(err)
	}

	models.InitializeRootUser(database.GetDB(), env.ROOT_USER_KEY)

	log.Println("⏳ Setting up background workers...")

	// Initialize the link expiration worker
	worker := workers.NewLinkExpirationWorker(database.GetDB())

	// Start the worker in a goroutine
	ctx := context.Background()
	go func() {
		if err := worker.Start(ctx); err != nil {
			log.Printf("Link expiration worker error: %v", err)
		}
	}()

	log.Println("✔️  Background workers set up successfully.")

	// Spin up the webserver
	err := workers.InitializeWebserver(env)
	if err != nil {
		log.Fatal(err)
	}
}
