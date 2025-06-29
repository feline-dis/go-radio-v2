# Go Radio v2

A modern radio streaming application built with Go and React, featuring real-time WebSocket communication, S3 file storage, and YouTube integration.

## Features

- Real-time radio streaming with WebSocket support
- YouTube video integration and download
- S3 file storage for audio files
- PostgreSQL database with migrations
- React frontend with modern UI
- Docker containerization
- Fly.io deployment ready

## Tech Stack

### Backend
- **Go 1.21+** - Main application server
- **Gorilla Mux** - HTTP router
- **Gorilla WebSocket** - Real-time communication
- **PostgreSQL** - Database
- **Atlas** - Database migrations
- **AWS SDK** - S3 integration
- **JWT** - Authentication

### Frontend
- **React 18** - UI framework
- **TypeScript** - Type safety
- **Vite** - Build tool
- **Tailwind CSS** - Styling
- **Axios** - HTTP client

### Infrastructure
- **Docker** - Containerization
- **Docker Compose** - Local development
- **Fly.io** - Cloud deployment
- **GitHub Actions** - CI/CD

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+
- Docker and Docker Compose
- PostgreSQL (or use Docker)

### Local Development

1. **Clone the repository**
   ```bash
   git clone https://github.com/feline-dis/go-radio-v2.git
   cd go-radio-v2
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start with Docker Compose**
   ```bash
   make dev-compose
   ```

4. **Access the application**
   - Frontend: http://localhost:5173
   - Backend API: http://localhost:8080
   - Metrics: http://localhost:9090

### Manual Setup

1. **Install dependencies**
   ```bash
   # Backend
   go mod download
   
   # Frontend
   cd client
   yarn install
   ```

2. **Set up PostgreSQL**
   ```bash
   # Create database
   createdb go_radio
   
   # Run migrations
   atlas migrate apply --env local
   ```

3. **Run the application**
   ```bash
   # Backend
   go run cmd/server/main.go
   
   # Frontend (in another terminal)
   cd client
   yarn dev
   ```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `POSTGRES_HOST` | PostgreSQL host | `localhost` |
| `POSTGRES_PORT` | PostgreSQL port | `5432` |
| `POSTGRES_USER` | PostgreSQL user | `postgres` |
| `POSTGRES_PASSWORD` | PostgreSQL password | `postgres` |
| `POSTGRES_DB` | PostgreSQL database | `go_radio` |
| `POSTGRES_SSLMODE` | PostgreSQL SSL mode | `disable` |
| `JWT_SECRET` | JWT signing secret | Required |
| `AWS_ACCESS_KEY_ID` | AWS access key | Required |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key | Required |
| `S3_BUCKET_NAME` | S3 bucket name | Required |
| `YOUTUBE_API_KEY` | YouTube API key | Required |

### Database Schema

The application uses PostgreSQL with the following main tables:

- **songs** - Song metadata and playback statistics
- **playlists** - Playlist definitions
- **playlist_songs** - Many-to-many relationship between playlists and songs

## API Endpoints

### Radio Control
- `GET /api/v1/radio/status` - Get current playback status
- `POST /api/v1/radio/play` - Start playback
- `POST /api/v1/radio/pause` - Pause playback
- `POST /api/v1/radio/next` - Play next song
- `POST /api/v1/radio/shuffle` - Toggle shuffle mode

### Playlists
- `GET /api/v1/playlists` - List all playlists
- `POST /api/v1/playlists` - Create new playlist
- `GET /api/v1/playlists/{id}` - Get playlist details
- `PUT /api/v1/playlists/{id}` - Update playlist
- `DELETE /api/v1/playlists/{id}` - Delete playlist

### YouTube Integration
- `POST /api/v1/youtube/add` - Add YouTube video to playlist
- `GET /api/v1/youtube/search` - Search YouTube videos

### WebSocket
- `WS /ws` - Real-time updates for playback status, queue changes, and user reactions

## Development

### Database Migrations

```bash
# Create new migration
atlas migrate diff migration_name

# Apply migrations
atlas migrate apply --env local

# Rollback migrations
atlas migrate down --env local
```

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage
```

### Building

```bash
# Build binary
make build

# Build Docker image
make docker-build
```

## Deployment

### Fly.io

1. **Install Fly CLI**
   ```bash
   curl -L https://fly.io/install.sh | sh
   ```

2. **Deploy**
   ```bash
   make fly-deploy
   ```

### Docker

```bash
# Build and run
make docker-build
docker run -p 8080:8080 go-radio
```

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   React Client  │    │   Go Backend    │    │   PostgreSQL    │
│                 │    │                 │    │                 │
│ - Radio Player  │◄──►│ - HTTP API      │◄──►│ - Songs         │
│ - Playlist Mgmt │    │ - WebSocket     │    │ - Playlists     │
│ - Real-time UI  │    │ - Radio Service │    │ - User Data     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │   AWS S3        │
                       │                 │
                       │ - Audio Files   │
                       │ - Static Assets │
                       └─────────────────┘
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.