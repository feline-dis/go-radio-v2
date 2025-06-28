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

docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE):$(VERSION) .
	@docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest

docker-push:
	@echo "Pushing Docker image..."
	@docker push $(DOCKER_IMAGE):$(VERSION)
	@docker push $(DOCKER_IMAGE):latest

deploy: docker-build docker-push
	@echo "Deploying to fly.io..."
	@flyctl deploy

lint:
	@echo "Running linter..."
	@golangci-lint run

deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Development helpers
dev:
	@echo "Starting development server..."
	@go run cmd/server/main.go

# Database migrations
migrate-up:
	@echo "Running database migrations..."
	@go run cmd/migrate/main.go up

migrate-down:
	@echo "Rolling back database migrations..."
	@go run cmd/migrate/main.go down

atlas-migrate:
	@echo "Applying database migrations with Atlas..."
	@atlas migrate apply --env local

# Fly.io deployment commands
fly-deploy:
	@echo "Deploying to fly.io..."
	@flyctl deploy

fly-launch:
	@echo "Launching new fly.io app..."
	@flyctl launch

fly-status:
	@echo "Checking fly.io app status..."
	@flyctl status

fly-logs:
	@echo "Showing fly.io app logs..."
	@flyctl logs

fly-volumes:
	@echo "Creating fly.io volume for data persistence..."
	@flyctl volumes create go_radio_data --size 1 --region iad

fly-secrets:
	@echo "Setting fly.io secrets..."
	@flyctl secrets set JWT_SECRET="your-secret-key-here"
	@flyctl secrets set AWS_ACCESS_KEY_ID="your-aws-access-key"
	@flyctl secrets set AWS_SECRET_ACCESS_KEY="your-aws-secret-key"
	@flyctl secrets set S3_BUCKET_NAME="your-s3-bucket"
	@flyctl secrets set YOUTUBE_API_KEY="your-youtube-api-key"

# Docker Compose commands
compose-up:
	@echo "Starting all services with Docker Compose..."
	@docker compose up -d || docker-compose up -d

compose-down:
	@echo "Stopping all services..."
	@docker compose down || docker-compose down

compose-build:
	@echo "Building all Docker images..."
	@docker compose build || docker-compose build

compose-logs:
	@echo "Showing Docker Compose logs..."
	@docker compose logs -f || docker-compose logs -f

compose-restart:
	@echo "Restarting all services..."
	@docker compose restart || docker-compose restart

compose-clean:
	@echo "Cleaning up Docker Compose resources..."
	@docker compose down -v --remove-orphans || docker-compose down -v --remove-orphans

# Development with Docker Compose
dev-compose: compose-build compose-up
	@echo "Development environment ready!"
	@echo "Frontend: http://localhost:3000"
	@echo "Backend API: http://localhost:8080"
	@echo "Metrics: http://localhost:9090"

# Production deployment with Docker Compose
prod-compose: compose-build
	@echo "Starting production environment..."
	@docker compose --profile production up -d || docker-compose --profile production up -d

# GitHub Actions commands
github-deploy:
	@echo "Triggering Fly.io deployment via GitHub Actions..."
	@gh workflow run deploy.yml

github-release:
	@echo "Creating a new release..."
	@read -p "Enter version tag (e.g., v1.0.0): " version; \
	git tag $$version; \
	git push origin $$version

# Development helpers
dev-setup:
	@echo "Setting up development environment..."
	@cp .env.example .env 2>/dev/null || echo "No .env.example found"
	@echo "Please edit .env with your configuration values"
	@echo "Then run: make dev-compose"

# Utility commands
logs-backend:
	@echo "Showing backend logs..."
	@docker compose logs -f backend || docker-compose logs -f backend

logs-frontend:
	@echo "Showing frontend logs..."
	@docker compose logs -f frontend || docker-compose logs -f frontend

logs-nginx:
	@echo "Showing nginx logs..."
	@docker compose logs -f nginx || docker-compose logs -f nginx

shell-backend:
	@echo "Opening shell in backend container..."
	@docker compose exec backend sh || docker-compose exec backend sh

shell-frontend:
	@echo "Opening shell in frontend container..."
	@docker compose exec frontend sh || docker-compose exec frontend sh

# Database commands
db-backup:
	@echo "Creating database backup..."
	@docker compose exec backend sqlite3 /app/data/radio.db ".backup /app/data/radio.db.backup.$(shell date +%Y%m%d_%H%M%S)" || docker-compose exec backend sqlite3 /app/data/radio.db ".backup /app/data/radio.db.backup.$(shell date +%Y%m%d_%H%M%S)"

db-restore:
	@echo "Restoring database from backup..."
	@read -p "Enter backup filename: " backup; \
	docker compose exec backend sqlite3 /app/data/radio.db ".restore /app/data/$$backup" || docker-compose exec backend sqlite3 /app/data/radio.db ".restore /app/data/$$backup" 