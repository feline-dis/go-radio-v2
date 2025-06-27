# GitHub Actions CI/CD Guide

This guide covers the simplified GitHub Actions workflows for building, testing, and deploying the Go Radio v2 application to Fly.io.

## Workflow Overview

The repository includes two simple GitHub Actions workflows:

1. **`pr-check.yml`** - Pull request validation
2. **`deploy.yml`** - Main deployment workflow (deploys to Fly.io)
3. **`release.yml`** - Release management (deploys to Fly.io)

## Workflow Details

### 1. Pull Request Checks (`pr-check.yml`)

**Triggers:** Pull requests to `main` or `master`

**Jobs:**
- **Test Backend**: Runs Go tests, linting, and coverage
- **Test Frontend**: Runs Node.js tests, linting, and build
- **Build Docker**: Builds Docker images (without pushing)
- **Security Scan**: Runs Trivy vulnerability scanner

**Features:**
- ✅ Parallel job execution
- ✅ Test coverage reporting
- ✅ Security vulnerability scanning
- ✅ Docker image validation

### 2. Main Deployment (`deploy.yml`)

**Triggers:** 
- Push to `main` or `master`

**Jobs:**
- **Test Backend**: Go tests and linting
- **Test Frontend**: Node.js tests and linting
- **Build**: Build and push Docker images to GitHub Container Registry
- **Deploy**: Deploy to Fly.io

**Features:**
- ✅ Automated testing
- ✅ Docker image building and pushing
- ✅ Automatic Fly.io deployment

### 3. Release Management (`release.yml`)

**Triggers:** Push tags matching `v*` (e.g., `v1.0.0`)

**Jobs:**
- **Build and Push**: Build and push versioned Docker images
- **Deploy**: Deploy to Fly.io

**Features:**
- ✅ Versioned Docker images
- ✅ Automatic Fly.io deployment

## Setup Instructions

### 1. Repository Secret

Configure this secret in your GitHub repository:

```bash
# Fly.io
FLY_API_TOKEN=your-fly-api-token
```

### 2. Environment Protection

Set up environment protection rules in GitHub:

1. Go to **Settings** → **Environments**
2. Create environment: `fly.io`
3. Configure protection rules (optional):
   - **Required reviewers** for deployment
   - **Wait timer** for deployment

### 3. GitHub Container Registry

The workflows automatically use GitHub Container Registry:

- **Backend**: `ghcr.io/your-username/go-radio-v2/backend:latest`
- **Frontend**: `ghcr.io/your-username/go-radio-v2/frontend:latest`

## Usage Examples

### Local Development

```bash
# Set up development environment
make dev-setup

# Start services locally
make dev-compose

# View logs
make logs-backend
make logs-frontend
```

### Manual Deployments

```bash
# Deploy to Fly.io via GitHub Actions
make github-deploy

# Create a new release
make github-release
```

### Using GitHub CLI

```bash
# Trigger deployment
gh workflow run deploy.yml

# View workflow runs
gh run list

# View specific run
gh run view <run-id>
```

## Docker Images

### Image Tags

The workflows create several image tags:

- **Latest**: `ghcr.io/your-username/go-radio-v2/backend:latest`
- **Version**: `ghcr.io/your-username/go-radio-v2/backend:v1.0.0`
- **Branch**: `ghcr.io/your-username/go-radio-v2/backend:main-abc123`

### Using Images Locally

```bash
# Pull latest images
docker pull ghcr.io/your-username/go-radio-v2/backend:latest
docker pull ghcr.io/your-username/go-radio-v2/frontend:latest

# Run with specific version
export IMAGE_TAG=v1.0.0
docker-compose up -d
```

## Deployment Flow

### Automatic Deployment

1. **Push to main branch** → Triggers `deploy.yml`
2. **Tests run** → Backend and frontend tests
3. **Docker images built** → Pushed to GitHub Container Registry
4. **Deploy to Fly.io** → Automatic deployment

### Release Deployment

1. **Create and push tag** → `git tag v1.0.0 && git push origin v1.0.0`
2. **Triggers release workflow** → Builds versioned images
3. **Deploy to Fly.io** → Automatic deployment

## Monitoring and Troubleshooting

### Health Checks

Check your Fly.io deployment:

```bash
# Check app status
flyctl status

# View logs
flyctl logs

# Check health endpoint
curl https://your-app.fly.dev/api/v1/health
```

### Debugging

```bash
# SSH into Fly.io app
flyctl ssh console

# View app info
flyctl info

# Restart app
flyctl restart
```

## Troubleshooting

### Common Issues

1. **Workflow Fails on Tests**
   - Check test output in workflow logs
   - Run tests locally: `go test ./...`
   - Fix linting issues: `golangci-lint run`

2. **Docker Build Fails**
   - Check Dockerfile syntax
   - Verify `.dockerignore` files
   - Check for missing dependencies

3. **Fly.io Deployment Fails**
   - Verify `FLY_API_TOKEN` secret is set
   - Check Fly.io app configuration
   - Review Fly.io logs: `flyctl logs`

4. **Health Checks Fail**
   - Check Fly.io app logs
   - Verify environment variables
   - Test endpoints manually

### Getting Help

- 📖 Check workflow logs in GitHub Actions tab
- 🐛 Report issues in GitHub repository
- 💬 Ask questions in GitHub Discussions

## Security Features

### Vulnerability Scanning
- **Trivy**: Scans for known vulnerabilities
- **Results**: Uploaded to GitHub Security tab
- **Blocking**: Can block deployments on high-severity issues

### Secret Management
- **GitHub Secrets**: Encrypted storage for sensitive data
- **Fly.io Secrets**: Managed via `flyctl secrets set`

## Best Practices

### 1. Code Quality
- ✅ Write tests for all new features
- ✅ Maintain test coverage above 80%
- ✅ Use linting and formatting tools
- ✅ Review security scan results

### 2. Deployment
- ✅ Use semantic versioning for releases
- ✅ Monitor deployments and rollback if needed
- ✅ Keep deployment logs for audit

### 3. Security
- ✅ Rotate secrets regularly
- ✅ Monitor for security vulnerabilities
- ✅ Keep dependencies updated

This simplified setup provides a clean CI/CD pipeline focused on Fly.io deployment with proper testing and security scanning. 