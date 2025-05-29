.PHONY: build run test lint clean tidy

APP_NAME=api-template
BIN_DIR=./bin

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BIN_DIR)
	go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME) ./cmd/api

# Run the application
run:
	@echo "Running $(APP_NAME)..."
	go run $(LDFLAGS) ./cmd/api

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run test coverage
cover:
	@echo "Running test coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Clean build files
clean:
	@echo "Cleaning..."
	rm -rf $(BIN_DIR)
	rm -f coverage.out

# Update dependencies
tidy:
	@echo "Updating dependencies..."
	go mod tidy

# Setup pre-commit hooks
setup-pre-commit:
	@echo "Setting up pre-commit hooks..."
	pre-commit install

# Generate OpenAPI documentation
generate-docs:
	@echo "Generating OpenAPI documentation..."
	swag init -g cmd/api/main.go -o docs

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):$(VERSION) .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 $(APP_NAME):$(VERSION)

# Default target
all: lint test generate-docs build
