# Makefile for driftGo project

# Variables
BINARY_NAME=driftGo
MAIN_PATH=./cmd/server
DOCKER_COMPOSE_FILE=docker-compose.yml
MIGRATION_DIR=db/goose_migrations
SQLC_CONFIG=sqlc.yaml
SQLC_GEN_DIRS=domain/user domain/link

# Database connection details (matching docker-compose.yml)
DB_HOST=localhost
DB_PORT=5434
DB_USER=drift
DB_PASSWORD=drift
DB_NAME=drift
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# Go build flags
LDFLAGS=-ldflags "-X main.Version=$(shell git describe --tags --always --dirty)"

.PHONY: help build run run-build clean docker-up docker-down docker-down-volumes docker-restart wait-for-db migrate-up migrate-down migrate-reset sqlc-gen sqlc-clean sqlc-reset test lint fmt vet dev-setup dev-clean dev-reset db-status db-connect

# Default target
help: ## Show this help message
	@echo ""
	@echo "🚀 driftGo Development Commands"
	@echo "================================="
	@echo ""
	@echo "📦 Docker Operations:"
	@echo "  docker-up           - Start Docker containers"
	@echo "  docker-down         - Stop and remove Docker containers"
	@echo "  docker-down-volumes - Stop containers AND delete all data (WARNING!)"
	@echo "  docker-restart      - Restart Docker containers"
	@echo ""
	@echo "🗄️  Database Operations:"
	@echo "  migrate-up          - Run database migrations up"
	@echo "  migrate-down        - Rollback database migrations"
	@echo "  migrate-reset       - Reset database (down all, then up all)"
	@echo "  db-status           - Check database migration status"
	@echo "  db-connect          - Connect to database (requires psql)"
	@echo ""
	@echo "🔨 Build Operations:"
	@echo "  build               - Build the application"
	@echo "  run                 - Run the application with Air (hot reload)"
	@echo "  run-build           - Run the application without hot reload"
	@echo "  clean               - Clean build artifacts"
	@echo ""
	@echo "⚙️  SQLC Operations:"
	@echo "  sqlc-gen            - Generate SQLC code"
	@echo "  sqlc-clean          - Clean SQLC generated files"
	@echo "  sqlc-reset          - Clean and regenerate SQLC code"
	@echo ""
	@echo "🛠️  Development Workflow:"
	@echo "  dev-setup           - Complete development setup"
	@echo "  dev-clean           - Clean development environment (keeps data)"
	@echo "  dev-reset           - Complete reset (WARNING: deletes all data!)"
	@echo ""
	@echo "🧪 Development Tools:"
	@echo "  test                - Run tests"
	@echo "  lint                - Run linter"
	@echo "  fmt                 - Format code"
	@echo "  vet                 - Vet code"
	@echo ""
	@echo "❓ Utilities:"
	@echo "  help                - Show this help message"
	@echo ""

# Database readiness check function
wait-for-db: ## Wait for database to be ready
	@echo "⏳ Waiting for database to be ready..."
	@timeout=30; \
	while [ $$timeout -gt 0 ]; do \
		if docker exec driftPSQL pg_isready -U $(DB_USER) -d $(DB_NAME) >/dev/null 2>&1; then \
			echo "🟢 Database is ready!"; \
			break; \
		fi; \
		echo "⏳ Database not ready yet, waiting... ($$timeout seconds left)"; \
		sleep 1; \
		timeout=$$((timeout - 1)); \
	done; \
	if [ $$timeout -eq 0 ]; then \
		echo "🔴 Database failed to start within 30 seconds"; \
		exit 1; \
	fi

# Docker Operations
docker-up: ## Start Docker containers
	@echo "🐳 Starting Docker containers..."
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d
	@echo "🟢 Docker containers started successfully!"
	@echo "📊 Database will be available at $(DB_HOST):$(DB_PORT)"
	$(MAKE) wait-for-db

docker-down: ## Stop and remove Docker containers
	@echo "🛑 Stopping Docker containers..."
	docker-compose -f $(DOCKER_COMPOSE_FILE) down
	@echo "🟢 Docker containers stopped and removed"

docker-down-volumes: ## Stop and remove Docker containers AND volumes (WARNING: Deletes all data!)
	@echo "🟡 WARNING: Stopping Docker containers and removing volumes..."
	@echo "🟡 This will delete ALL database data!"
	docker-compose -f $(DOCKER_COMPOSE_FILE) down -v
	@echo "🔴 Docker containers and volumes removed"

docker-restart: ## Restart Docker containers
	@echo "🔄 Restarting Docker containers..."
	docker-compose -f $(DOCKER_COMPOSE_FILE) restart
	@echo "🟢 Docker containers restarted"

# Database Migration Operations
migrate-up: ## Run database migrations up
	@echo "📈 Running database migrations up..."
	@if ! docker ps | grep -q driftPSQL; then \
		echo "🟡 Database container is not running. Starting it first..."; \
		$(MAKE) docker-up; \
	else \
		$(MAKE) wait-for-db; \
	fi
	goose -dir $(MIGRATION_DIR) postgres "$(DB_URL)" up
	@echo "🟢 Database migrations completed successfully!"

migrate-down: ## Rollback database migrations
	@echo "📉 Rolling back database migrations..."
	goose -dir $(MIGRATION_DIR) postgres "$(DB_URL)" down
	@echo "🟢 Database migrations rolled back"

migrate-reset: ## Reset database (down all, then up all)
	@echo "🔄 Resetting database migrations..."
	goose -dir $(MIGRATION_DIR) postgres "$(DB_URL)" reset
	@echo "🟢 Database migrations reset completed"

# Build Operations
build: ## Build the application
	@echo "🔨 Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "🟢 Build completed: $(BINARY_NAME)"

run: ## Run the application with Air (hot reload)
	@echo "🚀 Running $(BINARY_NAME) with Air (hot reload)..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "🟡 Air not found. Installing Air..."; \
		go install github.com/cosmtrek/air@latest; \
		air; \
	fi

run-build: ## Run the application without hot reload
	@echo "🚀 Running $(BINARY_NAME)..."
	go run $(MAIN_PATH)

clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	@for dir in $(SQLC_GEN_DIRS); do \
		rm -f $$dir/*.gen.go; \
	done
	rm -rf tmp/
	@echo "🟢 Build artifacts cleaned"

# SQLC Operations
sqlc-gen: ## Generate SQLC code
	@echo "⚙️  Generating SQLC code..."
	sqlc generate -f $(SQLC_CONFIG)
	@echo "🟢 SQLC code generation completed"

sqlc-clean: ## Clean SQLC generated files
	@echo "🧹 Cleaning SQLC generated files..."
	@for dir in $(SQLC_GEN_DIRS); do \
		rm -f $$dir/*.gen.go; \
	done
	@echo "🟢 SQLC generated files cleaned"

sqlc-reset: sqlc-clean sqlc-gen ## Clean and regenerate SQLC code
	@echo "🟢 SQLC reset completed"

# Development Operations
test: ## Run tests
	@echo "🧪 Running tests..."
	go test -v ./...

lint: ## Run linter
	@echo "🔍 Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "🟡 golangci-lint not found. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

fmt: ## Format code
	@echo "✨ Formatting code..."
	go fmt ./...

vet: ## Vet code
	@echo "🔍 Vetting code..."
	go vet ./...

# Development workflow targets
dev-setup: docker-up migrate-up sqlc-gen ## Complete development setup
	@echo "🟢 Development environment setup completed!"

dev-clean: docker-down clean ## Complete cleanup
	@echo "🟢 Development environment cleaned up!"

dev-reset: docker-down-volumes clean ## Complete reset (WARNING: Deletes all data!)
	@echo "🔴 WARNING: Development environment completely reset!"
	@echo "🔴 All database data has been deleted!"
	@echo "🟡 Run 'make dev-setup' to start fresh"

# Database inspection (useful for debugging)
db-status: ## Check database migration status
	@echo "📊 Checking database migration status..."
	goose -dir $(MIGRATION_DIR) postgres "$(DB_URL)" status

db-connect: ## Connect to database (requires psql)
	@echo "🔌 Connecting to database..."
	@if command -v psql >/dev/null 2>&1; then \
		psql "$(DB_URL)"; \
	else \
		echo "🟡 psql not found. Please install PostgreSQL client tools."; \
	fi 