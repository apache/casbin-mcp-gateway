.PHONY: help build test clean run dev install frontend-install frontend-build frontend-dev

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## Install backend dependencies
	go mod download

test: ## Run tests
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

build: ## Build the backend binary
	go build -o mcp-gateway .

build-prod: ## Build production binary
	go build -ldflags="-s -w" -o mcp-gateway .

run: ## Run the backend server
	go run main.go

frontend-install: ## Install frontend dependencies
	cd web && npm install

frontend-build: ## Build frontend for production
	cd web && npm run build

frontend-dev: ## Run frontend in development mode
	cd web && npm run dev

dev: ## Run backend and build frontend
	$(MAKE) frontend-build
	$(MAKE) run

all: install frontend-install build frontend-build ## Install all dependencies and build both backend and frontend

clean: ## Clean build artifacts
	rm -f mcp-gateway
	rm -rf web/dist
	rm -f coverage.txt

docker-build: ## Build Docker image
	docker build -t mcp-gateway:latest .

.DEFAULT_GOAL := help
