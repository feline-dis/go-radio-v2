#!/bin/bash

# Go Radio v2 Deployment Script
# This script helps deploy the application using Docker Compose

set -e

echo "🎵 Go Radio v2 Deployment Script"
echo "=================================="

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed or not in PATH"
    echo "Please install Docker from https://docs.docker.com/get-docker/"
    exit 1
fi

# Check if Docker Compose is available
if docker compose version &> /dev/null; then
    DOCKER_COMPOSE="docker compose"
elif docker-compose --version &> /dev/null; then
    DOCKER_COMPOSE="docker-compose"
else
    echo "❌ Docker Compose is not available"
    echo "Please install Docker Compose or ensure Docker Desktop is running"
    exit 1
fi

echo "✅ Docker and Docker Compose are available"

# Function to check if .env file exists
check_env_file() {
    if [ ! -f .env ]; then
        echo "⚠️  No .env file found. Creating template..."
        cat > .env << EOF
# Go Radio v2 Environment Variables
# Replace these values with your actual secrets

JWT_SECRET=your-secret-key-here
AWS_ACCESS_KEY_ID=your-aws-access-key
AWS_SECRET_ACCESS_KEY=your-aws-secret-key
S3_BUCKET_NAME=your-s3-bucket
YOUTUBE_API_KEY=your-youtube-api-key
EOF
        echo "📝 Created .env template. Please edit it with your actual values."
        echo "   Then run this script again."
        exit 1
    fi
}

# Function to deploy with Docker Compose
deploy_docker_compose() {
    echo "🐳 Building Docker images..."
    $DOCKER_COMPOSE build

    echo "🚀 Starting services..."
    $DOCKER_COMPOSE up -d

    echo "⏳ Waiting for services to be ready..."
    sleep 10

    echo "🔍 Checking service status..."
    $DOCKER_COMPOSE ps

    echo ""
    echo "🎉 Deployment complete!"
    echo "📱 Frontend: http://localhost:3000"
    echo "🔧 Backend API: http://localhost:8080"
    echo "📊 Metrics: http://localhost:9090"
    echo "🏥 Health Check: http://localhost:8080/api/v1/health"
    echo ""
    echo "📋 Useful commands:"
    echo "   View logs: $DOCKER_COMPOSE logs -f"
    echo "   Stop services: $DOCKER_COMPOSE down"
    echo "   Restart services: $DOCKER_COMPOSE restart"
}

# Function to show deployment options
show_options() {
    echo ""
    echo "Choose deployment option:"
    echo "1) Deploy with Docker Compose (recommended)"
    echo "2) Deploy to Fly.io (cloud)"
    echo "3) Show deployment status"
    echo "4) Stop all services"
    echo "5) View logs"
    echo "6) Exit"
    echo ""
    read -p "Enter your choice (1-6): " choice

    case $choice in
        1)
            check_env_file
            deploy_docker_compose
            ;;
        2)
            echo "☁️  Deploying to Fly.io..."
            echo "Please ensure you have flyctl installed and are authenticated."
            echo "Run: make fly-deploy"
            ;;
        3)
            echo "📊 Service Status:"
            $DOCKER_COMPOSE ps
            ;;
        4)
            echo "🛑 Stopping all services..."
            $DOCKER_COMPOSE down
            echo "✅ Services stopped"
            ;;
        5)
            echo "📋 Showing logs (Ctrl+C to exit):"
            $DOCKER_COMPOSE logs -f
            ;;
        6)
            echo "👋 Goodbye!"
            exit 0
            ;;
        *)
            echo "❌ Invalid choice. Please try again."
            show_options
            ;;
    esac
}

# Main execution
if [ "$1" = "deploy" ]; then
    check_env_file
    deploy_docker_compose
elif [ "$1" = "stop" ]; then
    echo "🛑 Stopping all services..."
    $DOCKER_COMPOSE down
    echo "✅ Services stopped"
elif [ "$1" = "logs" ]; then
    echo "📋 Showing logs (Ctrl+C to exit):"
    $DOCKER_COMPOSE logs -f
elif [ "$1" = "status" ]; then
    echo "📊 Service Status:"
    $DOCKER_COMPOSE ps
else
    show_options
fi 