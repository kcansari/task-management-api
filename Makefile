# Makefile for Task Management API
# Provides convenient commands for development, testing, and deployment

# Variables
APP_NAME=task-api
DOCKER_IMAGE=task-management-api
VERSION?=latest
GOOS?=linux
GOARCH?=amd64

# Default target
.DEFAULT_GOAL := help

# Help target - displays available commands
.PHONY: help
help: ## Display this help message
	@echo "Task Management API - Available Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

# Development commands
.PHONY: dev
dev: ## Start development server with hot reload (requires air)
	@echo "Starting development server with hot reload..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not installed. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Running without hot reload..."; \
		go run main.go; \
	fi

.PHONY: run
run: ## Run the application
	@echo "Starting Task Management API..."
	go run main.go

.PHONY: build
build: ## Build the application binary
	@echo "Building $(APP_NAME)..."
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "-w -s" -o $(APP_NAME) .

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -f $(APP_NAME)
	go clean

# Testing commands
.PHONY: test
test: ## Run all tests
	@echo "Running tests..."
	go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-unit
test-unit: ## Run only unit tests
	@echo "Running unit tests..."
	go test -v ./utils/...

.PHONY: test-integration
test-integration: ## Run only integration tests
	@echo "Running integration tests..."
	go test -v ./handlers/...

# Code quality commands
.PHONY: lint
lint: ## Run linter (requires golangci-lint)
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

.PHONY: format
format: ## Format code with gofmt
	@echo "Formatting code..."
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

.PHONY: deps
deps: ## Download and tidy dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Database commands
.PHONY: db-setup
db-setup: ## Set up database (requires running PostgreSQL)
	@echo "Setting up database..."
	@echo "Make sure PostgreSQL is running and update .env with correct credentials"
	go run main.go &
	sleep 5
	pkill -f "go run main.go" || true

# Docker commands
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(VERSION) .

.PHONY: docker-run
docker-run: ## Run application in Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE):$(VERSION)

.PHONY: docker-compose-up
docker-compose-up: ## Start full stack with Docker Compose
	@echo "Starting full stack with Docker Compose..."
	docker-compose up --build

.PHONY: docker-compose-down
docker-compose-down: ## Stop Docker Compose stack
	@echo "Stopping Docker Compose stack..."
	docker-compose down

.PHONY: docker-compose-logs
docker-compose-logs: ## View Docker Compose logs
	docker-compose logs -f

# Production deployment commands
.PHONY: build-prod
build-prod: ## Build production binary with optimizations
	@echo "Building production binary..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags "-w -s -X main.version=$(VERSION)" \
		-a -installsuffix cgo \
		-o $(APP_NAME) .

.PHONY: docker-prod
docker-prod: ## Build production Docker image
	@echo "Building production Docker image..."
	docker build -t $(DOCKER_IMAGE):$(VERSION) -f Dockerfile .
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest

# Security commands
.PHONY: security-scan
security-scan: ## Run security scan (requires gosec)
	@echo "Running security scan..."
	@if command -v gosec > /dev/null; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Performance commands
.PHONY: benchmark
benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# API testing commands
.PHONY: api-test
api-test: ## Test API endpoints (requires server running)
	@echo "Testing API endpoints..."
	@echo "Make sure the server is running on localhost:8080"
	curl -s http://localhost:8080/health || echo "Server not responding"
	@echo "\nAPI Health Check completed"

# Environment setup
.PHONY: setup
setup: ## Initial project setup
	@echo "Setting up development environment..."
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "Created .env file from .env.example"; \
		echo "Please update .env with your configuration"; \
	fi
	go mod download
	@echo "Setup completed!"

# Migration commands (when implemented)
.PHONY: migrate
migrate: ## Run database migrations
	@echo "Running database migrations..."
	@echo "Migrations are handled automatically on startup"

# Utility commands
.PHONY: check
check: format vet lint test ## Run all checks (format, vet, lint, test)

.PHONY: pre-commit
pre-commit: format vet test ## Pre-commit hook (format, vet, test)

.PHONY: install-tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Documentation
.PHONY: docs
docs: ## Open API documentation
	@echo "Opening API documentation..."
	@if command -v open > /dev/null; then \
		open API_DOCUMENTATION.md; \
	elif command -v xdg-open > /dev/null; then \
		xdg-open API_DOCUMENTATION.md; \
	else \
		echo "Please open API_DOCUMENTATION.md in your preferred editor"; \
	fi