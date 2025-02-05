# Binary name
BINARY_NAME=go-link-shortener

# Go related variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOFILES=$(wildcard *.go)

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

# Linker flags
LDFLAGS=-ldflags "-w -s"

# Add color support for terminal output
BLUE=\033[0;34m
GREEN=\033[0;32m
RED=\033[0;31m
NC=\033[0m # No Color

# Default target
.DEFAULT_GOAL := help

.PHONY: all build clean run deps test coverage vet lint help

all: clean deps build test ## Build the application and run tests

swagger: 
	@go install github.com/swaggo/swag/cmd/swag@latest
	@printf "$(BLUE)Constructing swagger docs...$(NC)\n"
	@swag init -q
	@printf "$(GREEN)Swagger docs constructed successfully!$(NC)\n"

build: swagger ## Build the application
	@printf "$(BLUE)Building application...$(NC)\n"
	@LOCAL_BUILD=true go build ${LDFLAGS} -o ${GOBIN}/${BINARY_NAME} .
	@printf "$(GREEN)Build complete! Binary location: ${GOBIN}/${BINARY_NAME}$(NC)\n"

clean: ## Clean build files
	@printf "$(BLUE)Cleaning build cache...$(NC)\n"
	@go clean
	@rm -rf ${GOBIN}
	@printf "$(GREEN)Cleaned build files and cache$(NC)\n"

run: build ## Build and run the application
	@printf "$(BLUE)Starting server...$(NC)\n"
	@LOCAL_BUILD=true ${GOBIN}/${BINARY_NAME}

deps: ## Download and verify dependencies
	@printf "$(BLUE)Downloading dependencies...$(NC)\n"
	@go mod download
	@go mod verify
	@printf "$(GREEN)Dependencies ready$(NC)\n"

test: swagger ## Run tests
	@printf "$(BLUE)Running tests...$(NC)\n"
	@ENVIRONMENT=test go test -v ./...
	@printf "$(GREEN)Tests complete$(NC)\n"

coverage: swagger ## Run tests with coverage
	@printf "$(BLUE)Running tests with coverage...$(NC)\n"
	@mkdir -p coverage
	@ENVIRONMENT=test go test -v -coverprofile=coverage/coverage.out ./...
	@go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@printf "$(GREEN)Coverage analysis complete. See coverage/coverage.html for detailed report$(NC)\n"

vet: ## Run go vet
	@printf "$(BLUE)Running go vet...$(NC)\n"
	@go vet ./...
	@printf "$(GREEN)Vet check complete$(NC)\n"

lint: swagger ## Run linter
	@printf "$(BLUE)Running linter...$(NC)\n"
	@if command -v golangci-lint >/dev/null; then \
		golangci-lint run ./...; \
	else \
		printf "$(RED)Error: golangci-lint not installed$(NC)\n"; \
		printf "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest\n"; \
		exit 1; \
	fi

setup: swagger ## Setup the project
	@make deps
	@printf "$(BLUE)Setting up project...$(NC)\n"
	@chmod +x ./scripts/*.sh
	@bash ./scripts/setup.sh
	@printf "$(GREEN)=================================================$(NC)\n"
	@printf "$(GREEN)Project setup complete! $(NC)\n"
	@printf "$(GREEN)Check .env file for credentials.$(NC)\n"
	@printf "$(GREEN)Use 'make run' to start the server.$(NC)\n"
	@printf "$(GREEN)Use 'make help' for more commands.$(NC)\n"
	@printf "$(GREEN)=================================================$(NC)\n"

docker-build: ## Build the Docker image
	@printf "$(BLUE)Building Docker image...$(NC)\n"
	@docker compose build
	@printf "$(GREEN)Docker image built successfully!$(NC)\n"

docker-run: ## Run the application using Docker Compose
	@printf "$(BLUE)Starting Docker containers in interactive mode...$(NC)\n"
	@docker compose up
	@printf "$(GREEN)Docker containers started!$(NC)\n"

docker-rebuild-app: ## Rebuild the application inside the Docker container
	@printf "$(BLUE)Rebuilding application inside Docker container...$(NC)\n"
	@docker compose build --no-cache app && docker compose restart app
	@printf "$(GREEN)Application rebuilt successfully!$(NC)\n"

docker-rebuild-db: ## Rebuild the database inside the Docker container
	@printf "$(BLUE)Rebuilding database inside Docker container...$(NC)\n"
	@docker compose build db && docker compose restart db
	@printf "$(GREEN)Database rebuilt successfully!$(NC)\n"

docker-rebuild-all: docker-rebuild-app docker-rebuild-db ## Rebuild the application and database inside the Docker container
	@printf "$(GREEN)All containers rebuilt successfully!$(NC)\n"
	@printf "$(GREEN)You can now run 'make docker-run' to start the application and database.$(NC)\n"
	@printf "$(GREEN)Use 'make help' for more commands.$(NC)\n"
	@printf "$(GREEN)=================================================$(NC)\n"
	@printf "$(GREEN)Docker commands complete!$(NC)\n"

docker-stop: ## Stop the Docker containers
	@printf "$(BLUE)Stopping Docker containers...$(NC)\n"
	@docker compose stop
	@printf "$(GREEN)Docker containers stopped!$(NC)\n"
	@printf "$(GREEN)Use 'make docker-start' to start the containers.$(NC)\n"

docker-start: ## Start the Docker containers
	@printf "$(BLUE)Starting Docker containers in detached mode...$(NC)\n"
	@docker compose up -d
	@printf "$(GREEN)Docker containers started!$(NC)\n"
	@printf "$(GREEN)Use 'make docker-stop' to stop the containers.$(NC)\n"
	@printf "$(GREEN)=================================================$(NC)\n"
	@printf "$(GREEN)Docker commands complete!$(NC)\n"
	@printf "$(GREEN)You can now access the application at http://localhost:$(SERVER_PORT)$(NC)\n"
	@printf "$(GREEN)The API docs are available at http://localhost:$(SERVER_PORT)/docs/$(NC)\n"
	@printf "$(GREEN)Use 'make help' for more commands.$(NC)\n"
	@printf "$(GREEN)=================================================$(NC)\n"

docker-restart: ## Restart the Docker containers
	@printf "$(BLUE)Restarting Docker containers...$(NC)\n"
	@docker compose restart
	@printf "$(GREEN)Docker containers restarted!$(NC)\n"

help: ## Display available commands
	@echo "Available commands:"
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(GREEN)%-20s$(NC) %s\n", $$1, $$2}'