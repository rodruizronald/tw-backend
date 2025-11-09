.PHONY: test lint lint-fix docs help migrate-up migrate-down migrate-status migrate-create migrate-force migrate-drop migrate-goto migrate-up-by migrate-down-by docker-build docker-up docker-down docker-logs docker-restart docker-clean docker-ps

# Load environment variables from .env file if it exists
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Database configuration - these will use values from .env if present, otherwise defaults
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= postgres
DB_PASSWORD ?= postgres
DB_NAME ?= postgres
DB_SSL_MODE ?= disable

# Construct DATABASE_URL from individual components
DATABASE_URL = postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)

# Default target
.DEFAULT_GOAL := help

# Run all unit tests
test:
	@echo "Running all unit tests..."
	@go test ./...
	@echo "✅ Tests completed successfully"

# Run linter
lint:
	@echo "Running golangci-lint..."
	@golangci-lint run
	@echo "✅ Linting completed successfully"

# Run linter with fix
lint-fix:
	@echo "Running golangci-lint with fix..."
	@golangci-lint run --fix
	@echo "✅ Linting with fixes completed successfully"

# Generate swagger documentation
docs:
	@echo "Generating swagger documentation..."
	@swag init \
		-g main.go \
		-d .,\
./internal/jobs,\
./internal/company,\
./internal/technology,\
./internal/jobtech,\
./internal/techalias \
		-o ./docs
	@echo "✅ Swagger docs generated successfully"

# Migration commands
migrate-up:
	@echo "Applying all migrations..."
	@migrate -path migrations -database "$(DATABASE_URL)" up
	@echo "✅ All migrations applied successfully"

migrate-down:
	@echo "Rolling back all migrations..."
	@migrate -path migrations -database "$(DATABASE_URL)" down
	@echo "✅ All migrations rolled back successfully"

migrate-status:
	@echo "Checking migration status..."
	@migrate -path migrations -database "$(DATABASE_URL)" version

migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "❌ Error: NAME is required. Usage: make migrate-create NAME=your_migration_name"; \
		exit 1; \
	fi
	@echo "Creating new migration: $(NAME)"
	@migrate create -ext sql -dir migrations -seq $(NAME)
	@echo "✅ Migration files created successfully"

migrate-force:
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ Error: VERSION is required. Usage: make migrate-force VERSION=001"; \
		exit 1; \
	fi
	@echo "⚠️  Forcing migration to version $(VERSION)..."
	@migrate -path migrations -database "$(DATABASE_URL)" force $(VERSION)
	@echo "✅ Migration forced to version $(VERSION)"

migrate-drop:
	@echo "⚠️  WARNING: This will drop all tables and data!"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
    echo; \
    if [ "$$REPLY" = "y" ] || [ "$$REPLY" = "Y" ]; then \
        migrate -path migrations -database "$(DATABASE_URL)" drop; \
        echo "✅ Database dropped successfully"; \
    else \
        echo "❌ Operation cancelled"; \
    fi

migrate-goto:
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ Error: VERSION is required. Usage: make migrate-goto VERSION=3"; \
		exit 1; \
	fi
	@echo "Migrating to version $(VERSION)..."
	@migrate -path migrations -database "$(DATABASE_URL)" goto $(VERSION)
	@echo "✅ Migrated to version $(VERSION)"

migrate-up-by:
	@if [ -z "$(STEPS)" ]; then \
		echo "❌ Error: STEPS is required. Usage: make migrate-up-by STEPS=2"; \
		exit 1; \
	fi
	@echo "Applying $(STEPS) migrations..."
	@migrate -path migrations -database "$(DATABASE_URL)" up $(STEPS)
	@echo "✅ Applied $(STEPS) migrations successfully"

migrate-down-by:
	@if [ -z "$(STEPS)" ]; then \
		echo "❌ Error: STEPS is required. Usage: make migrate-down-by STEPS=1"; \
		exit 1; \
	fi
	@echo "Rolling back $(STEPS) migrations..."
	@migrate -path migrations -database "$(DATABASE_URL)" down $(STEPS)
	@echo "✅ Rolled back $(STEPS) migrations successfully"

# Docker commands
docker-build:
	@echo "Building Docker images..."
	@docker-compose -f docker/docker-compose.yml build
	@echo "✅ Docker images built successfully"

docker-up:
	@echo "Starting Docker containers..."
	@docker-compose -f docker/docker-compose.yml up -d
	@echo "✅ Docker containers started successfully"
	@echo "Application available at http://localhost:8080"

docker-down:
	@echo "Stopping Docker containers..."
	@docker-compose -f docker/docker-compose.yml down
	@echo "✅ Docker containers stopped successfully"

docker-logs:
	@echo "Showing Docker logs (Ctrl+C to exit)..."
	@docker-compose -f docker/docker-compose.yml logs -f

docker-restart:
	@echo "Restarting Docker containers..."
	@docker-compose -f docker/docker-compose.yml restart
	@echo "✅ Docker containers restarted successfully"

docker-clean:
	@echo "⚠️  WARNING: This will remove all containers, volumes, and images!"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
    echo; \
    if [ "$$REPLY" = "y" ] || [ "$$REPLY" = "Y" ]; then \
        docker-compose -f docker/docker-compose.yml down -v --rmi all; \
        echo "✅ Docker resources cleaned successfully"; \
    else \
        echo "❌ Operation cancelled"; \
    fi

docker-ps:
	@echo "Docker container status:"
	@docker-compose -f docker/docker-compose.yml ps

# Show help
help:
	@echo "Available commands:"
	@echo ""
	@echo "Development:"
	@echo "  test           - Run all unit tests"
	@echo "  lint           - Run golangci-lint"
	@echo "  lint-fix       - Run golangci-lint with fix"
	@echo "  docs           - Generate swagger documentation"
	@echo ""
	@echo "Database Migrations:"
	@echo "  migrate-up     - Apply all pending migrations"
	@echo "  migrate-down   - Rollback all migrations"
	@echo "  migrate-status - Check current migration version"
	@echo "  migrate-create NAME=<name> - Create new migration files"
	@echo "  migrate-force VERSION=<ver> - Force migration to specific version"
	@echo "  migrate-drop   - Drop all tables (with confirmation)"
	@echo "  migrate-goto VERSION=<ver>  - Migrate to specific version"
	@echo "  migrate-up-by STEPS=<n>     - Apply N migrations"
	@echo "  migrate-down-by STEPS=<n>   - Rollback N migrations"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build   - Build Docker images"
	@echo "  docker-up      - Start Docker containers in detached mode"
	@echo "  docker-down    - Stop Docker containers"
	@echo "  docker-logs    - Show Docker container logs"
	@echo "  docker-restart - Restart Docker containers"
	@echo "  docker-clean   - Remove all containers, volumes, and images (with confirmation)"
	@echo "  docker-ps      - Show Docker container status"
	@echo ""
	@echo "Examples:"
	@echo "  make migrate-create NAME=add_user_table"
	@echo "  make migrate-up-by STEPS=2"
	@echo "  make migrate-force VERSION=001"
	@echo "  make docker-up"
	@echo "  make docker-logs"
	@echo ""
	@echo "Note: Database configuration is read from .env file or environment variables"