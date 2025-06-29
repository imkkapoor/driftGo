# Makefile for driftGo project

# Variables
BINARY_NAME=driftGo
MAIN_PATH=./cmd/server
DOCKER_COMPOSE_FILE=docker-compose.yml
MIGRATION_DIR=db/goose_migrations
SQLC_CONFIG=sqlc.yaml
SQLC_GEN_DIR=domain/user/sqlcgen

# Database connection details (matching docker-compose.yml)
DB_HOST=localhost
DB_PORT=5432
DB_USER=drift
DB_PASSWORD=drift
DB_NAME=drift
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# Go build flags
LDFLAGS=-ldflags "-X main.Version=$(shell git describe --tags --always --dirty)"

.PHONY: help build run run-build clean docker-up docker-down docker-restart migrate-up migrate-down migrate-reset sqlc-gen sqlc-clean sqlc-reset test lint fmt vet

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@echo ""
	@echo "Docker Operations:"
	@echo "  docker-up       - Start Docker containers"
	@echo "  docker-down     - Stop and remove Docker containers"
	@echo "  docker-restart  - Restart Docker containers"
	@echo ""
	@echo "Database Operations:"
	@echo "  migrate-up      - Run database migrations up"
	@echo "  migrate-down    - Rollback database migrations"
	@echo "  migrate-reset   - Reset database (down all, then up all)"
	@echo ""
	@echo "Build Operations:"
	@echo "  build           - Build the application"
	@echo "  run             - Run the application with Air (hot reload)"
	@echo "  run-build       - Run the application without hot reload"
	@echo "  clean           - Clean build artifacts"
	@echo ""
	@echo "SQLC Operations:"
	@echo "  sqlc-gen        - Generate SQLC code"
	@echo "  sqlc-clean      - Clean SQLC generated files"
	@echo "  sqlc-reset      - Clean and regenerate SQLC code"
	@echo ""
	@echo "Development:"
	@echo "  test            - Run tests"
	@echo "  lint            - Run linter"
	@echo "  fmt             - Format code"
	@echo "  vet             - Vet code"
	@echo ""
	@echo "Utilities:"
	@echo "  help            - Show this help message"

# Docker Operations
docker-up: ## Start Docker containers
	@echo "Starting Docker containers..."
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d
	@echo "Docker containers started. Waiting for database to be ready..."
	@echo "Database will be available at $(DB_HOST):$(DB_PORT)"

docker-down: ## Stop and remove Docker containers
	@echo "Stopping Docker containers..."
	docker-compose -f $(DOCKER_COMPOSE_FILE) down
	@echo "Docker containers stopped and removed"

docker-restart: ## Restart Docker containers
	@echo "Restarting Docker containers..."
	docker-compose -f $(DOCKER_COMPOSE_FILE) restart
	@echo "Docker containers restarted"

# Database Migration Operations
migrate-up: ## Run database migrations up
	@echo "Running database migrations up..."
	@if ! docker ps | grep -q driftPSQL; then \
		echo "Database container is not running. Starting it first..."; \
		$(MAKE) docker-up; \
		sleep 5; \
	fi
	goose -dir $(MIGRATION_DIR) postgres "$(DB_URL)" up
	@echo "Database migrations completed"

migrate-down: ## Rollback database migrations
	@echo "Rolling back database migrations..."
	goose -dir $(MIGRATION_DIR) postgres "$(DB_URL)" down
	@echo "Database migrations rolled back"

migrate-reset: ## Reset database (down all, then up all)
	@echo "Resetting database migrations..."
	goose -dir $(MIGRATION_DIR) postgres "$(DB_URL)" reset
	@echo "Database migrations reset completed"

# Build Operations
build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "Build completed: $(BINARY_NAME)"

run: ## Run the application with Air (hot reload)
	@echo "Running $(BINARY_NAME) with Air (hot reload)..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "Air not found. Installing Air..."; \
		go install github.com/cosmtrek/air@latest; \
		air; \
	fi

run-build: ## Run the application without hot reload
	@echo "Running $(BINARY_NAME)..."
	go run $(MAIN_PATH)

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	rm -rf $(SQLC_GEN_DIR)
	rm -rf tmp/
	@echo "Build artifacts cleaned"

# SQLC Operations
sqlc-gen: ## Generate SQLC code
	@echo "Generating SQLC code..."
	sqlc generate -f $(SQLC_CONFIG)
	@echo "SQLC code generation completed"

sqlc-clean: ## Clean SQLC generated files
	@echo "Cleaning SQLC generated files..."
	rm -rf $(SQLC_GEN_DIR)
	@echo "SQLC generated files cleaned"

sqlc-reset: sqlc-clean sqlc-gen ## Clean and regenerate SQLC code
	@echo "SQLC reset completed"

# Development Operations
test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

lint: ## Run linter
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

vet: ## Vet code
	@echo "Vetting code..."
	go vet ./...

# Development workflow targets
dev-setup: docker-up migrate-up sqlc-gen ## Complete development setup
	@echo "Development environment setup completed!"

dev-clean: docker-down clean ## Complete cleanup
	@echo "Development environment cleaned up!"

# Database inspection (useful for debugging)
db-status: ## Check database migration status
	@echo "Checking database migration status..."
	goose -dir $(MIGRATION_DIR) postgres "$(DB_URL)" status

db-connect: ## Connect to database (requires psql)
	@echo "Connecting to database..."
	@if command -v psql >/dev/null 2>&1; then \
		psql "$(DB_URL)"; \
	else \
		echo "psql not found. Please install PostgreSQL client tools."; \
	fi 