MAKEFILE_VERSION := 1.0.0

.PHONY: help
help: ## Display this help message
	@echo "WardFlow Backend - Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: run
run: ## Run the application locally
	go run cmd/api/main.go

.PHONY: build
build: ## Build the application binary
	go build -o bin/wardflow-api cmd/api/main.go

.PHONY: test
test: ## Run all tests
	go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: lint
lint: ## Run linter
	golangci-lint run

.PHONY: fmt
fmt: ## Format code
	go fmt ./...

.PHONY: tidy
tidy: ## Tidy go modules
	go mod tidy

.PHONY: podman-build
podman-build: ## Build Podman images
	podman compose build

.PHONY: podman-up
podman-up: ## Start all containers
	podman compose up -d

.PHONY: podman-down
podman-down: ## Stop all containers
	podman compose down

.PHONY: podman-logs
podman-logs: ## View backend logs
	podman logs -f wardflow-backend

.PHONY: podman-restart
podman-restart: ## Restart backend after code changes
	podman compose build backend
	podman compose up -d --force-recreate backend

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf bin/ coverage.out coverage.html

.PHONY: dev
dev: fmt tidy ## Prepare for development (format + tidy)
	@echo "Ready for development"
