.PHONY: build run clean docker-build docker-run deploy test fmt lint help

# Variables
BINARY_NAME=feature-request-app
DOCKER_IMAGE=feature-request-app
PORT=8080

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development
build: ## Build the Go application
	@echo "Building $(BINARY_NAME)..."
	go mod tidy
	go build -o $(BINARY_NAME) .

run: ## Run the application locally
	@echo "Running $(BINARY_NAME) on port $(PORT)..."
	go run main.go

dev: ## Run in development mode with auto-reload (requires air)
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Install air for auto-reload: go install github.com/cosmtrek/air@latest"; \
		make run; \
	fi

# Testing
test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Code quality
fmt: ## Format Go code
	go fmt ./...

lint: ## Run linter (requires golangci-lint)
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "Install golangci-lint: https://golangci-lint.run/usage/install/"; \
	fi

# Docker
docker-build: ## Build Docker image
	@echo "Building Docker image $(DOCKER_IMAGE)..."
	docker build -t $(DOCKER_IMAGE):latest .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p $(PORT):$(PORT) --env-file .env $(DOCKER_IMAGE):latest

docker-compose-up: ## Start with docker-compose
	docker-compose up -d

docker-compose-down: ## Stop docker-compose
	docker-compose down

docker-logs: ## View docker-compose logs
	docker-compose logs -f

# Deployment
deploy: ## Deploy to production
	@echo "Deploying $(BINARY_NAME)..."
	chmod +x deploy.sh
	./deploy.sh

# Cleanup
clean: ## Clean build artifacts
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	docker image prune -f

# Setup
setup: ## Set up development environment
	@echo "Setting up development environment..."
	cp .env.example .env
	@echo "Please edit .env with your configuration"
	go mod tidy

# Production helpers
build-prod: ## Build for production
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o $(BINARY_NAME) .

install-tools: ## Install development tools
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
