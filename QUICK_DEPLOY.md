# Quick Deployment Guide

This guide will get your Go Radio v2 application running in minutes!

## Prerequisites

1. **Docker Desktop** - Download from [docker.com](https://www.docker.com/products/docker-desktop/)
2. **Git** - Download from [git-scm.com](https://git-scm.com/)

## Quick Start (5 minutes)

### Option 1: Docker Compose (Recommended)
```bash
# Clone the repository
git clone <your-repo-url>
cd go-radio-v2

# Run the deployment script
./deploy.sh
```

### Option 2: Manual Setup
```bash
# Backend
go mod download
make migrate-up
cd cmd/server && go run main.go

# Frontend
cd client
yarn install
yarn dev
```

## 🐳 Deployment Options

### 1. Local Development with Docker Compose
```bash
# Start all services
make dev-compose

# View logs
make compose-logs

# Stop services
make compose-down
```

### 2. Production Deployment
```bash
# Deploy with nginx reverse proxy
make prod-compose

# Or use the deployment script
./deploy.sh
```

### 3. Cloud Deployment (Fly.io)
```bash
# Deploy to Fly.io
make fly-deploy

# Or use GitHub Actions
make github-deploy-fly
```

### 4. GitHub Actions CI/CD
The project includes automated GitHub Actions workflows:

- **Pull Request Checks**: Automated testing and security scanning
- **Automatic Deployment**: Deploy to Fly.io on push to main branch
- **Release Management**: Create releases with versioned Docker images

```bash
# Deploy to Fly.io
make github-deploy

# Create a new release
make github-release
```

## Environment Setup

Before running, you'll need to set up your environment variables. The script will create a `.env` template for you:

```bash
# Edit the .env file with your actual values
JWT_SECRET=your-secret-key-here
AWS_ACCESS_KEY_ID=your-aws-access-key
AWS_SECRET_ACCESS_KEY=your-aws-secret-key
S3_BUCKET_NAME=your-s3-bucket
YOUTUBE_API_KEY=your-youtube-api-key
```

## Useful Commands

```bash
# View logs
./deploy.sh logs

# Check status
./deploy.sh status

# Stop services
./deploy.sh stop

# Deploy to Fly.io (cloud)
make fly-deploy
```

## Troubleshooting

### Docker not found
- Install Docker Desktop from [docker.com](https://www.docker.com/products/docker-desktop/)
- Make sure Docker is running

### Port already in use
- Stop any existing services: `./deploy.sh stop`
- Check what's using the ports: `netstat -an | grep :3000`

### Build fails
- Check your `.env` file has the correct values
- Ensure you have enough disk space
- Try: `docker system prune` to clean up

## Next Steps

- 📖 Read the full [Deployment Guide](docs/DEPLOYMENT.md)
- 🐛 Report issues on GitHub
- ⭐ Star the repository if you like it!

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Backend       │    │   Database      │
│   (React)       │    │   (Go API)      │    │   (SQLite)      │
│   Port 3000     │    │   Port 8080     │    │   (Volume)      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

All services run in Docker containers with automatic health checks and restart capabilities. 