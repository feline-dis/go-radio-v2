# Go Radio V2

A modern web-based radio player built with Go backend and React frontend with real-time user interactions.

## Features

### Backend (Go)
- RESTful API for playlist and song management
- WebSocket support for real-time updates
- S3 integration for audio file storage
- YouTube integration for song metadata
- SQLite database with migrations
- Clean architecture with separation of concerns
- Real-time reaction system for user interactions

### Frontend (React)
- Modern React with TypeScript
- Real-time audio playback using HTML5 Audio API
- Volume control with mute/unmute functionality
- Progress tracking and time display
- Queue management and display
- WebSocket integration for live updates
- Responsive design with Tailwind CSS
- Interactive reaction bar with animated emotes
- Real-time user reaction display

## 🚀 Quick Start

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
npm install
npm run dev
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

# Create release
make github-release
```

📖 **See [GitHub Actions Guide](docs/GITHUB_ACTIONS.md) for detailed CI/CD documentation**

## Audio Playback Features

The radio player now supports full audio playback through the browser:

- **Direct MP3 Streaming**: Audio files are streamed directly from S3 through the backend API
- **HTML5 Audio API**: Uses native browser audio capabilities for optimal performance
- **Volume Control**: Adjustable volume slider with mute/unmute toggle
- **Progress Tracking**: Real-time progress bar with elapsed/remaining time
- **Auto-play**: Automatically starts playing when a new song is queued
- **Error Handling**: Graceful handling of audio loading and playback errors
- **Cross-browser Support**: Works across all modern browsers

## Real-Time User Reactions

The radio player includes an interactive reaction system that allows users to express themselves while listening:

- **Reaction Bar**: A selection of emotes (❤️, 🔥, 🚀, 👏, 💃, 🎉, ⭐, 👍) for users to choose from
- **Real-time Broadcasting**: Reactions are instantly shared with all connected users via WebSocket
- **Animated Display**: Emotes appear with smooth floating animations across the screen
- **Automatic Cleanup**: Reactions automatically disappear after 3 seconds
- **Connection Status**: Visual feedback when WebSocket connection is lost

### Reaction Features

- **8 Different Emotes**: Heart, Fire, Rocket, Clap, Dance, Party, Star, and Thumbs Up
- **Random Positioning**: Emotes appear at random positions on screen for visual variety
- **Smooth Animations**: CSS animations provide smooth floating and fade effects
- **Real-time Sync**: All users see reactions simultaneously
- **Connection Handling**: Graceful degradation when WebSocket is disconnected

## Server-Synchronized Playback

The radio player implements advanced synchronization to ensure all clients are at the same point in the song:

- **Server Authority**: All playback timing is controlled by the server for consistency
- **Real-time Updates**: 10 FPS WebSocket updates provide smooth progress tracking
- **Latency Compensation**: Network latency is automatically compensated for accurate timing
- **Client Requests**: Clients can request current playback state for immediate synchronization
- **Timestamp-based Sync**: Millisecond-precision timestamps ensure accurate synchronization
- **Automatic Recovery**: Clients automatically resync if they fall out of sync

### Synchronization Features

- **High-frequency Updates**: WebSocket broadcasts at 100ms intervals for smooth progress
- **State Requests**: Clients can request current playback state via WebSocket messages
- **Latency Adjustment**: Elapsed time is adjusted for network latency automatically
- **Immediate Sync**: New clients receive current state immediately upon connection
- **Periodic Resync**: Clients request state updates every 5 seconds to prevent drift

## Development

### Prerequisites
- Go 1.21+
- Node.js 18+
- Docker (for containerized development)
- SQLite
- Atlas (for database migrations)

### Local Development
```bash
# Set up development environment
make dev-setup

# Start services
make dev-compose

# View logs
make logs-backend
make logs-frontend

# Run tests
make test
make test-coverage
```

### Backend Setup
```bash
# Install dependencies
go mod download

# Run database migrations
make migrate-up

# Start the server
cd cmd/server && go run main.go
```

### Frontend Setup
```bash
cd client

# Install dependencies
npm install

# Start development server
npm run dev
```

## 🏗️ Architecture

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

## 📋 Useful Commands

### Docker Compose
```bash
# Start services
make compose-up

# Stop services
make compose-down

# View logs
make compose-logs

# Restart services
make compose-restart

# Clean up
make compose-clean
```

### GitHub Actions
```bash
# Deploy to Fly.io
make github-deploy

# Create release
make github-release
```

### Database
```bash
# Create backup
make db-backup

# Restore from backup
make db-restore

# Run migrations
make migrate-up
```

## 🔧 Configuration

### Environment Variables
Create a `.env` file with your configuration:

```bash
# Required
JWT_SECRET=your-secret-key-here
AWS_ACCESS_KEY_ID=your-aws-access-key
AWS_SECRET_ACCESS_KEY=your-aws-secret-key
S3_BUCKET_NAME=your-s3-bucket
YOUTUBE_API_KEY=your-youtube-api-key

# Optional
VITE_API_BASE_URL=http://localhost:8080
```

### GitHub Secrets (for CI/CD)
Configure this secret in your GitHub repository:
- `FLY_API_TOKEN`: Fly.io API token

## 📚 Documentation

- **[Quick Deployment Guide](QUICK_DEPLOY.md)** - Get started in 5 minutes
- **[Deployment Guide](docs/DEPLOYMENT.md)** - Comprehensive deployment documentation
- **[GitHub Actions Guide](docs/GITHUB_ACTIONS.md)** - CI/CD pipeline documentation
- **[Event Bus Architecture](docs/EVENT_BUS_ARCHITECTURE.md)** - Backend architecture details
- **[Reaction System](docs/REACTION_SYSTEM.md)** - Real-time reaction system documentation

## API Endpoints

### Public Endpoints
- `GET /api/v1/health` - Health check
- `GET /api/v1/queue` - Get current queue information
- `GET /api/v1/now-playing` - Get currently playing song
- `GET /api/v1/playlists/{youtube_id}/file` - Stream MP3 audio file
- `POST /api/v1/reactions` - Send a user reaction

### Admin Endpoints
- `POST /api/v1/admin/play` - Start playback
- `POST /api/v1/admin/pause` - Pause playback
- `POST /api/v1/admin/skip` - Skip to next song
- `POST /api/v1/admin/rewind` - Rewind current song

## WebSocket

The application uses WebSocket for real-time updates:
- Endpoint: `/ws`
- Events: 
  - `playback_update` - Song and timing information
  - `user_reaction` - User reaction events
  - `song_change` - Song change notifications
  - `queue_update` - Queue updates

## License

MIT