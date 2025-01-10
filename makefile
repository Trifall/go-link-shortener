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

build: ## Build the application
	@printf "$(BLUE)Building application...$(NC)\n"
	@go build ${LDFLAGS} -o ${GOBIN}/${BINARY_NAME} .
	@printf "$(GREEN)Build complete! Binary location: ${GOBIN}/${BINARY_NAME}$(NC)\n"

clean: ## Clean build files
	@printf "$(BLUE)Cleaning build cache...$(NC)\n"
	@go clean
	@rm -rf ${GOBIN}
	@printf "$(GREEN)Cleaned build files and cache$(NC)\n"

run: build ## Build and run the application
	@printf "$(BLUE)Starting server...$(NC)\n"
	@${GOBIN}/${BINARY_NAME}

deps: ## Download and verify dependencies
	@printf "$(BLUE)Downloading dependencies...$(NC)\n"
	@go mod download
	@go mod verify
	@printf "$(GREEN)Dependencies ready$(NC)\n"

test: ## Run tests
	@printf "$(BLUE)Running tests...$(NC)\n"
	@go test -v ./...
	@printf "$(GREEN)Tests complete$(NC)\n"

coverage: ## Run tests with coverage
	@printf "$(BLUE)Running tests with coverage...$(NC)\n"
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@printf "$(GREEN)Coverage analysis complete. See coverage.html for detailed report$(NC)\n"

vet: ## Run go vet
	@printf "$(BLUE)Running go vet...$(NC)\n"
	@go vet ./...
	@printf "$(GREEN)Vet check complete$(NC)\n"

lint: ## Run linter
	@printf "$(BLUE)Running linter...$(NC)\n"
	@if command -v golangci-lint >/dev/null; then \
		golangci-lint run ./...; \
	else \
		printf "$(RED)Error: golangci-lint not installed$(NC)\n"; \
		printf "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest\n"; \
		exit 1; \
	fi

setup: ## Setup the project
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

help: ## Display available commands
	@echo "Available commands:"
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(GREEN)%-20s$(NC) %s\n", $$1, $$2}'