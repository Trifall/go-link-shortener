package main

import (
	"context"
	"log"

	"go-link-shortener/models"
	"go-link-shortener/utils"
	"go-link-shortener/workers"
)

func main() {

	log.Println("Starting Link Shortener")

	env := utils.LoadEnv()

	utils.SetDB(utils.ConnectToDatabase(env))

	// Setup database
	if err := models.SetupDatabase(utils.GetDB()); err != nil {
		log.Fatal(err)
	}

	utils.InitializeRootUser(utils.GetDB(), env.ROOT_USER_KEY)

	log.Println("⏳ Setting up background workers...")

	// Initialize the link expiration worker
	worker := workers.NewLinkExpirationWorker(utils.GetDB())

	// Start the worker in a goroutine
	ctx := context.Background()
	go func() {
		if err := worker.Start(ctx); err != nil {
			log.Printf("Link expiration worker error: %v", err)
		}
	}()

	log.Println("✔️  Background workers set up successfully.")

	// Spin up the webserver
	err := workers.InitializeWebserver()
	if err != nil {
		log.Fatal(err)
	}
}
