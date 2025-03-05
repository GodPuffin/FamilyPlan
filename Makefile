.PHONY: run build clean test lint dev docker-build docker-run

# Default target
all: run

# Run the application
run:
	go run main.go

# Run with hot reload using Air
dev:
	air

# Build the application
build:
	go build -o app main.go

# Clean build artifacts
clean:
	rm -f app
	rm -rf pb_data

# Install dependencies
deps:
	go mod tidy

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -html=coverage.txt -o coverage.html

# Lint the code
lint:
	go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping"; \
	fi

# Build Docker image
docker-build:
	docker build -t family-plan-app .

# Run Docker container
docker-run:
	docker run -p 8090:8090 family-plan-app

# Help
help:
	@echo "Available targets:"
	@echo "  run           - Run the application"
	@echo "  dev           - Run with hot reload using Air"
	@echo "  build         - Build the application"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Install dependencies"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  lint          - Lint the code"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  help          - Show this help message" 