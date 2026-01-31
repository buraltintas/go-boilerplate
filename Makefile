.PHONY: help run build test clean deps

# Variables
BINARY_NAME=api
MAIN_PATH=cmd/api/main.go

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

run: ## Run the application
	@echo "Starting server..."
	go run $(MAIN_PATH)

build: ## Build the application
	@echo "Building..."
	go build -o bin/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Binary built: bin/$(BINARY_NAME)"

test: ## Run tests
	@echo "Running tests..."
	go test -v -race ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	@echo "Clean complete"

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

lint: fmt vet ## Run linters

dev: ## Run in development mode with auto-reload (requires air)
	@which air > /dev/null || (echo "air not installed. Install with: go install github.com/cosmtrek/air@latest" && exit 1)
	air

install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/cosmtrek/air@latest
	@echo "Tools installed"
