.PHONY: build run test clean docker-build docker-push deploy

# Build variables
BINARY_NAME=go-radio
DOCKER_IMAGE=feline-dis/go-radio
VERSION=$(shell git describe --tags --always --dirty)

# Go related variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOFILES=$(wildcard cmd/server/*.go internal/**/*.go)

build:
	@echo "Building..."
	@go build -o $(GOBIN)/$(BINARY_NAME) cmd/server/main.go

run: build
	@echo "Running..."
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
	@rm -f coverage.out

migrate:
	@echo "Running migrations..."
	@atlas migrate diff --dir file://migrations  --to file://schema.hcl --env local

migrate-up:
	@echo "Applying migrations..."
	@atlas migrate apply --dir file://migrations --env local

migrate-down:
	@echo "Rolling back migrations..."
	@atlas migrate down --dir file://migrations --env local