.PHONY: build run test clean setup config docker-build docker-push deploy

# Build variables
BINARY_NAME=go-radio-server
SETUP_BINARY=go-radio-setup
DOCKER_IMAGE=feline-dis/go-radio
VERSION=$(shell git describe --tags --always --dirty)

# Go related variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOFILES=$(wildcard cmd/**/*.go internal/**/*.go)

# Setup and build
setup:
	@./scripts/setup.sh

config:
	@./scripts/setup.sh config

build:
	@echo "Building server..."
	@mkdir -p $(GOBIN)
	@go build -o $(GOBIN)/$(BINARY_NAME) cmd/server/main.go
	@echo "Building setup tool..."
	@go build -o $(GOBIN)/$(SETUP_BINARY) cmd/setup/main.go

build-frontend:
	@echo "Building frontend..."
	@cd client && yarn build

run: build
	@echo "Running server..."
	@./bin/$(BINARY_NAME)

test:
	@echo "Running tests..."
	@go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

clean:
	@echo "Cleaning..."
	@rm -f $(GOBIN)/$(BINARY_NAME)
	@rm -f $(GOBIN)/$(SETUP_BINARY)
	@rm -f coverage.out
	@cd client && rm -rf dist

# Development commands
dev-deps:
	@echo "Installing development dependencies..."
	@go mod download
	@cd client && yarn install

dev: build
	@echo "Starting development server..."
	@./bin/$(BINARY_NAME)

# Quick start - full setup and run
start: setup run