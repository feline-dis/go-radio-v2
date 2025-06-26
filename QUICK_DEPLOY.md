# Quick Deployment Guide

This guide will get your Go Radio v2 application running in minutes!

## Prerequisites

1. **Docker Desktop** - Download from [docker.com](https://www.docker.com/products/docker-desktop/)
2. **Git** - Download from [git-scm.com](https://git-scm.com/)

## Quick Start (5 minutes)

### 1. Clone and Navigate
```bash
git clone <your-repo-url>
cd go-radio-v2
```

### 2. Run the Deployment Script
```bash
./deploy.sh
```

The script will:
- ✅ Check if Docker is installed
- ✅ Create a `.env` template (if needed)
- ✅ Build all Docker images
- ✅ Start all services
- ✅ Show you the URLs to access your app

### 3. Access Your Application

Once deployment is complete, you can access:

- **🎵 Frontend**: http://localhost:3000
- **🔧 Backend API**: http://localhost:8080
- **📊 Metrics**: http://localhost:9090
- **🏥 Health Check**: http://localhost:8080/api/v1/health

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