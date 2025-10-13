# Variables
APP_NAME := jam-tracker
BINARY_NAME := jam-tracker
DOCKER_IMAGE := jam-tracker:latest

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod
GORUN := $(GOCMD) run

# Build info
BUILD_DIR := ./bin
MAIN_PATH := ./cmd/api/main.go

# Tools
TOOLS_DIR := $(shell go env GOPATH)/bin
AIR := $(TOOLS_DIR)/air

.PHONY: help build run test clean deps fmt vet dev install-tools

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Development tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/air-verse/air@latest
	@echo "✅ Air installed successfully!"

check-air: ## Check if air is installed
	@which air > /dev/null || (echo "Air not found. Installing..." && $(MAKE) install-tools)

# Development
build: ## Build the application
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

run: ## Run the application directly
	@echo "Running $(APP_NAME)..."
	$(GORUN) $(MAIN_PATH)

dev: check-air ## Run with hot reload using Air
	@echo "Starting development server with hot reload..."
	@mkdir -p air
	@air

dev-clean: check-air ## Clean air cache and restart
	@echo "Cleaning air cache and restarting..."
	@rm -rf air/
	@mkdir -p air
	@air

# Testing
test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	$(GOTEST) -race -v ./...

# Code quality
fmt: ## Format Go code
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOCMD) vet ./...

# Dependencies
deps: ## Download and clean up dependencies
	@echo "Managing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "✅ Dependencies updated!"

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	$(GOCMD) get -u ./...
	$(GOMOD) tidy

# Database
db-up: ## Start database containers
	@echo "Starting database..."
	@docker compose up -d postgres
	@echo "Waiting for database to be ready..."
	@sleep 3

db-down: ## Stop database containers
	@echo "Stopping database..."
	@docker compose down

db-reset: ## Reset database (drop and recreate)
	@echo "Resetting database..."
	@docker compose down -v
	@docker compose up -d postgres
	@echo "Waiting for database to be ready..."
	@sleep 5

db-shell: ## Connect to database shell
	@echo "Connecting to database..."
	@docker exec -it jamtracker_postgres psql -U jamtracker -d jamtracker

db-logs: ## View database logs
	@docker compose logs -f postgres


# Cleanup
clean: ## Clean build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -rf air/
	@rm -f coverage.out coverage.html

# Setup and utility commands
setup: deps install-tools db-up ## Complete development environment setup
	@echo "🎉 Development environment ready!"
	@echo ""
	@echo "Available commands:"
	@echo "  make run     - Run without hot reload"
	@echo "  make dev     - Run with hot reload"
	@echo "  make test    - Run tests"
	@echo "  make help    - Show all commands"

start: db-up dev ## Start database and run app with hot reload

stop: db-down ## Stop all services

check: fmt vet test ## Run all quality checks

ci: deps fmt vet test-race ## Run CI pipeline locally

fresh: clean setup dev ## Complete fresh start (clean + setup + dev)