# Go Radio v2 - Deployment Guide

This guide covers deploying the Go Radio application using both Docker Compose (for local/production servers) and Fly.io (for cloud deployment).

## Prerequisites

### For Docker Compose Deployment
1. **Docker**: Install Docker and Docker Compose
2. **Make**: Install Make (usually pre-installed on Linux/macOS)

### For Fly.io Deployment
1. **Fly.io CLI**: Install the Fly CLI
   ```bash
   # macOS
   brew install flyctl
   
   # Windows
   powershell -Command "iwr https://fly.io/install.ps1 -useb | iex"
   
   # Linux
   curl -L https://fly.io/install.sh | sh
   ```

2. **Fly.io Account**: Sign up at [fly.io](https://fly.io) and authenticate
   ```bash
   flyctl auth login
   ```

## Docker Compose Deployment

### Local Development

1. **Start the development environment**:
   ```bash
   make dev-compose
   ```

   This will:
   - Build all Docker images
   - Start backend, frontend, and nginx services
   - Make the application available at:
     - Frontend: http://localhost:3000
     - Backend API: http://localhost:8080
     - Metrics: http://localhost:9090

2. **View logs**:
   ```bash
   make compose-logs
   ```

3. **Stop services**:
   ```bash
   make compose-down
   ```

### Production Deployment

1. **Build and start production environment**:
   ```bash
   make prod-compose
   ```

   This includes the nginx reverse proxy for production use.

2. **Set environment variables** (create a `.env` file):
   ```bash
   JWT_SECRET=your-secret-key-here
   AWS_ACCESS_KEY_ID=your-aws-access-key
   AWS_SECRET_ACCESS_KEY=your-aws-secret-key
   S3_BUCKET_NAME=your-s3-bucket
   YOUTUBE_API_KEY=your-youtube-api-key
   ```

### Docker Compose Commands

```bash
# Build all images
make compose-build

# Start services
make compose-up

# Stop services
make compose-down

# View logs
make compose-logs

# Restart services
make compose-restart

# Clean up (removes volumes)
make compose-clean
```

## Fly.io Cloud Deployment

### Initial Setup

1. **Create Fly.io App**:
   ```bash
   make fly-launch
   ```

   Follow the prompts and select:
   - App name: `go-radio-v2` (or your preferred name)
   - Region: `iad` (US East)
   - Build strategy: `Dockerfile`

2. **Create Persistent Volume**:
   ```bash
   make fly-volumes
   ```

   This creates a 1GB volume in the `iad` region for storing the database.

3. **Set Environment Secrets**:
   ```bash
   make fly-secrets
   ```

   **Important**: Replace the placeholder values with your actual secrets:
   ```bash
   flyctl secrets set JWT_SECRET="your-actual-jwt-secret"
   flyctl secrets set AWS_ACCESS_KEY_ID="your-actual-aws-access-key"
   flyctl secrets set AWS_SECRET_ACCESS_KEY="your-actual-aws-secret-key"
   flyctl secrets set S3_BUCKET_NAME="your-actual-s3-bucket"
   flyctl secrets set YOUTUBE_API_KEY="your-actual-youtube-api-key"
   ```

### Deploy to Fly.io

```bash
make fly-deploy
```

This will:
1. Build the Docker image
2. Push it to Fly.io
3. Deploy the application

### Fly.io Management

```bash
# Check deployment status
make fly-status

# View application logs
make fly-logs

# Scale the application
flyctl scale count 2

# Restart the application
flyctl restart
```

## Architecture Overview

### Docker Compose Setup

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Nginx Proxy   │    │   Frontend      │    │   Backend       │
│   (Port 80/443) │    │   (Port 3000)   │    │   (Port 8080)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   SQLite DB     │
                    │   (Volume)      │
                    └─────────────────┘
```

### Service Communication

- **Frontend** (React): Serves the web interface
- **Backend** (Go): Provides API endpoints and WebSocket connections
- **Nginx**: Reverse proxy that routes requests appropriately
- **Database**: SQLite database stored in a Docker volume

### Port Mapping

- **80/443**: Nginx reverse proxy (HTTP/HTTPS)
- **3000**: Frontend service (direct access)
- **8080**: Backend API (direct access)
- **9090**: Metrics endpoint (direct access)

## Configuration

### Environment Variables

#### Backend Variables
- `PORT`: Application port (8080)
- `SQLITE_DB_PATH`: Database file path (/app/data/radio.db)
- `LOG_LEVEL`: Logging level (info)
- `ENABLE_METRICS`: Enable metrics endpoint (true)
- `METRICS_PORT`: Metrics port (9090)

#### Frontend Variables
- `VITE_API_BASE_URL`: API base URL (http://localhost:8080 for local, /api/v1 for production)

#### Secrets (set via flyctl secrets or .env file)
- `JWT_SECRET`: Secret for JWT token signing
- `AWS_ACCESS_KEY_ID`: AWS access key for S3
- `AWS_SECRET_ACCESS_KEY`: AWS secret key for S3
- `S3_BUCKET_NAME`: S3 bucket name for audio files
- `YOUTUBE_API_KEY`: YouTube API key for video processing

## Health Checks

### Backend Health Check
- Endpoint: `/api/v1/health`
- Response: `{"status": "healthy", "timestamp": 1703000000000}`

### Frontend Health Check
- Endpoint: `/health`
- Response: `healthy`

### Docker Compose Health Checks
- Backend: Checks `/api/v1/health` every 30s
- Frontend: Checks `/health` every 30s
- Services wait for health checks before starting dependent services

## Monitoring

### Application Metrics
- Backend metrics: `http://localhost:9090/metrics` (local) or `https://your-app.fly.dev:9090/metrics` (fly.io)

### Logs
```bash
# Docker Compose logs
make compose-logs

# Fly.io logs
make fly-logs
```

### Status
```bash
# Docker Compose status
docker-compose ps

# Fly.io status
make fly-status
```

## Troubleshooting

### Common Issues

1. **Database Connection Errors**
   - Ensure the volume is created: `make fly-volumes` (fly.io) or check Docker volumes
   - Check volume mount in configuration files

2. **Build Failures**
   - Verify Dockerfiles are correct
   - Check `.dockerignore` files exclude unnecessary files
   - Ensure all dependencies are properly specified

3. **Environment Variable Issues**
   - Verify secrets are set: `flyctl secrets list` (fly.io)
   - Check `.env` file exists (Docker Compose)
   - Verify configuration files

4. **Port Binding Issues**
   - Ensure ports are exposed in Dockerfiles
   - Verify port mappings in `docker-compose.yml`
   - Check `fly.toml` port configuration

5. **Frontend-Backend Communication Issues**
   - Verify API base URL configuration
   - Check nginx proxy configuration
   - Ensure WebSocket connections work through proxy

### Debug Commands

```bash
# Docker Compose debugging
docker-compose exec backend sh
docker-compose exec frontend sh
docker-compose logs backend
docker-compose logs frontend

# Fly.io debugging
flyctl ssh console
flyctl info
flyctl scale count 1
flyctl restart
```

## Scaling

### Docker Compose Scaling
```bash
# Scale backend service
docker-compose up -d --scale backend=2

# Scale frontend service
docker-compose up -d --scale frontend=2
```

### Fly.io Scaling
```bash
# Scale to 2 instances
flyctl scale count 2

# Scale with specific resources
flyctl scale vm shared-cpu-1x --memory 1024
```

## Security

### HTTPS
- Fly.io automatically provides HTTPS
- Docker Compose can be configured with SSL certificates in nginx

### Secrets Management
- Never commit secrets to version control
- Use `flyctl secrets` for Fly.io deployment
- Use `.env` files for Docker Compose (add to `.gitignore`)

### Rate Limiting
- Nginx configuration includes rate limiting for API endpoints
- WebSocket connections have appropriate timeouts

## Cost Optimization

### Fly.io
- Uses `shared-cpu-1x` with 1GB RAM
- Auto-scaling reduces costs when not in use
- Volume storage charged per GB used

### Docker Compose
- Runs on your own infrastructure
- No additional cloud costs beyond hosting

## Support

For deployment issues:
- Check application logs: `make compose-logs` or `make fly-logs`
- Review health endpoints: `/api/v1/health` and `/health`
- Verify configuration files are correct

For Fly.io specific issues:
- [Fly.io Documentation](https://fly.io/docs/)
- [Fly.io Community](https://community.fly.io/)

For Docker Compose issues:
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Docker Community](https://community.docker.com/) 