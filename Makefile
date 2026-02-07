.PHONY: help build run test test-coverage clean fmt vet lint swagger dev install-tools env migrate-up migrate-down migrate-status migrate-create

# --- OS detection ---
ifeq ($(OS),Windows_NT)
  IS_WINDOWS := 1
else
  IS_WINDOWS := 0
endif

# --- Variables ---
BINARY_NAME := server
ifeq ($(IS_WINDOWS),1)
  EXE := .exe
else
  EXE :=
endif

BINARY_PATH := ./bin/$(BINARY_NAME)$(EXE)
MAIN_PATH := ./cmd/server/main.go
COVERAGE_FILE := coverage.out
MIGRATION_DIR := ./migrations/postgres

# --- Helpers (cross-platform commands) ---
ifeq ($(IS_WINDOWS),1)
  # PowerShell helpers
  MKDIR_BIN = powershell -NoProfile -ExecutionPolicy Bypass -Command "New-Item -ItemType Directory -Force 'bin' | Out-Null"
  RM_BIN    = powershell -NoProfile -ExecutionPolicy Bypass -Command "if (Test-Path 'bin') { Remove-Item -Recurse -Force 'bin' }"
  RM_COV    = powershell -NoProfile -ExecutionPolicy Bypass -Command "if (Test-Path '$(COVERAGE_FILE)') { Remove-Item -Force '$(COVERAGE_FILE)' }"
  ENV_COPY  = powershell -NoProfile -ExecutionPolicy Bypass -Command "if (!(Test-Path '.env')) { Write-Host 'Creating .env from config.example.env...'; Copy-Item 'config.example.env' '.env'; Write-Host '.env created. Please update with your local settings.' }"
  REQUIRE_NAME = powershell -NoProfile -ExecutionPolicy Bypass -Command "if ([string]::IsNullOrWhiteSpace('$(name)')) { Write-Error 'Error: migration name is required. Usage: make migrate-create name=migration_name'; exit 1 }"
else
  # POSIX helpers
  MKDIR_BIN = mkdir -p bin
  RM_BIN    = rm -rf bin/
  RM_COV    = rm -f $(COVERAGE_FILE)
  ENV_COPY  = sh -c 'if [ ! -f .env ]; then echo "Creating .env from config.example.env..."; cp config.example.env .env; echo ".env created. Please update with your local settings."; fi'
  REQUIRE_NAME = sh -c 'if [ -z "$(name)" ]; then echo "Error: migration name is required. Usage: make migrate-create name=migration_name"; exit 1; fi'
endif

## help: Display this help message
help:
	@echo "Available targets:"
	@echo "  make build            - Build the application binary"
	@echo "  make run              - Build and run the application"
	@echo "  make dev              - Run application with hot reload (requires air)"
	@echo "  make test             - Run all tests"
	@echo "  make test-coverage    - Run tests with coverage report"
	@echo "  make fmt              - Format code with go fmt"
	@echo "  make vet              - Run go vet"
	@echo "  make lint             - Run golangci-lint (requires golangci-lint)"
	@echo "  make swagger          - Generate Swagger documentation"
	@echo "  make clean            - Remove binary and coverage files"
	@echo "  make env              - Copy config.example.env to .env if not exists"
	@echo "  make install-tools    - Install development tools (swag, goose, air, golangci-lint)"
	@echo "  make migrate-up       - Run database migrations up"
	@echo "  make migrate-down     - Rollback last database migration"
	@echo "  make migrate-status   - Show migration status"
	@echo "  make migrate-create   - Create new migration (usage: make migrate-create name=migration_name)"

## build: Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@$(MKDIR_BIN)
	@go build -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "Build complete: $(BINARY_PATH)"

## run: Build and run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	@$(BINARY_PATH)

## dev: Run with hot reload using air
dev: env
	@echo "Starting development server with hot reload..."
	@air

## test: Run all tests
test:
	@echo "Running tests..."
	@go test -v ./...

## test-coverage: Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@echo "Coverage report:"
	@go tool cover -func=$(COVERAGE_FILE)
	@echo ""
	@echo "To view HTML coverage report, run: go tool cover -html=$(COVERAGE_FILE)"

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

## lint: Run golangci-lint
lint: fmt vet
	@echo "Running golangci-lint..."
	@golangci-lint run ./...

## swagger: Generate Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	@swag init -g $(MAIN_PATH) -o ./docs
	@echo "Swagger docs generated in ./docs"

## clean: Remove binary and coverage files
clean:
	@echo "Cleaning..."
	@$(RM_BIN)
	@$(RM_COV)
	@echo "Clean complete"

## env: Create .env from example if it doesn't exist
env:
	@$(ENV_COPY)

## install-tools: Install required development tools
install-tools:
	@echo "Installing development tools..."
	@echo "Installing swag..."
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "Installing goose..."
	@go install github.com/pressly/goose/v3/cmd/goose@latest
	@echo "Installing air (hot reload)..."
	@go install github.com/air-verse/air@latest
	@echo "Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "All tools installed!"

## migrate-up: Run migrations up
migrate-up:
	@echo "Running migrations..."
	@goose -dir $(MIGRATION_DIR) postgres "$(DATABASE_URL)" up

## migrate-down: Rollback last migration
migrate-down:
	@echo "Rolling back last migration..."
	@goose -dir $(MIGRATION_DIR) postgres "$(DATABASE_URL)" down

## migrate-status: Show migration status
migrate-status:
	@echo "Migration status:"
	@goose -dir $(MIGRATION_DIR) postgres "$(DATABASE_URL)" status

## migrate-create: Create new migration (usage: make migrate-create name=migration_name)
migrate-create:
	@$(REQUIRE_NAME)
	@echo "Creating migration: $(name)"
	@goose -dir $(MIGRATION_DIR) create $(name) sql
	@echo "Migration created in $(MIGRATION_DIR)"
