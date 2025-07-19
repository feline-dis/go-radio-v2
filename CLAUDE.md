# Go Radio v2 - Claude Context

## Project Overview
Go Radio v2 is a modern streaming radio application built with Go backend and React frontend. It features real-time WebSocket communication, YouTube integration, S3 file storage, and a comprehensive playlist management system.

## Architecture

### Backend (Go)
- **Main Server**: `cmd/server/main.go`
- **Config Management**: `internal/config/config.go`
- **Clean Architecture**: Controllers → Services → Repositories
- **Database**: PostgreSQL with Atlas migrations
- **File Storage**: AWS S3 integration
- **Real-time**: WebSocket support via Gorilla WebSocket
- **Authentication**: JWT-based auth system

### Frontend (React)
- **Framework**: React 19 with TypeScript
- **Build Tool**: Vite
- **State Management**: Context API + TanStack Query
- **Routing**: React Router v7
- **UI**: Component-based architecture with Tailwind CSS
- **Real-time**: WebSocket client integration

## Key Technologies

### Backend Dependencies
- `gorilla/mux` - HTTP router
- `gorilla/websocket` - WebSocket support
- `lib/pq` - PostgreSQL driver
- `aws-sdk-go-v2` - AWS S3 integration
- `golang-jwt/jwt/v5` - JWT authentication
- `joho/godotenv` - Environment variable management

### Frontend Dependencies
- `react` v19 - Core framework
- `@tanstack/react-query` - Server state management
- `axios` - HTTP client
- `react-router-dom` v7 - Client-side routing
- `butterchurn` + `butterchurn-presets` - Audio visualizations
- `vissonance` - Custom visualizer
- `three` - 3D graphics for visualizations
- `socket.io-client` - WebSocket client
- `react-hook-form` + `zod` - Form handling and validation

## Project Structure

```
go-radio-v2/
├── cmd/
│   ├── server/main.go          # Main application entry point
│   └── download/main.go        # YouTube download utility
├── internal/
│   ├── config/config.go        # Configuration management
│   ├── controllers/            # HTTP route handlers
│   │   ├── auth_controller.go
│   │   ├── playlist_controller.go
│   │   ├── radio_controller.go
│   │   ├── reaction_controller.go
│   │   └── youtube_controller.go
│   ├── events/                 # Event bus system
│   │   ├── event_bus.go
│   │   └── event_bus_test.go
│   ├── middleware/             # HTTP middleware
│   │   ├── auth.go
│   │   └── logging.go
│   ├── models/                 # Data models
│   │   └── song.go
│   ├── repositories/           # Data access layer
│   │   ├── playlist_repository.go
│   │   └── song_repository.go
│   ├── services/               # Business logic
│   │   ├── jwt_service.go
│   │   ├── jwt_service_test.go
│   │   ├── playlist_service.go
│   │   ├── radio_service.go
│   │   ├── radio_service_test.go
│   │   ├── s3_service.go
│   │   └── youtube_service.go
│   └── websocket/              # WebSocket handler
│       └── handler.go
├── client/                     # React frontend
│   ├── src/
│   │   ├── components/         # Reusable UI components
│   │   ├── contexts/           # React context providers
│   │   ├── pages/              # Page components
│   │   ├── lib/                # Utility libraries
│   │   └── types/              # TypeScript type definitions
│   ├── package.json
│   └── vite.config.ts
├── migrations/                 # Database migrations
├── schema.hcl                  # Atlas database schema
├── atlas.hcl                   # Atlas configuration
├── docker-compose.yml          # Local development setup
├── Dockerfile                  # Backend container
├── Makefile                    # Build automation
└── fly.toml                    # Fly.io deployment config
```

## Database Schema

### Tables
1. **songs** - Core song metadata
   - `youtube_id` (PK) - YouTube video identifier
   - `title`, `artist`, `album` - Metadata
   - `duration` - Song length in seconds
   - `s3_key` - AWS S3 file location
   - `last_played`, `play_count` - Usage tracking
   - Indexes on `play_count` and `last_played`

2. **playlists** - Playlist definitions
   - `id` (UUID PK) - Unique identifier
   - `name` (unique), `description` - Playlist info
   - Timestamps for creation/updates

3. **playlist_songs** - Many-to-many relationship
   - `playlist_id`, `youtube_id` (composite PK)
   - `position` - Song order in playlist
   - Foreign keys with CASCADE delete

## Configuration

### Environment Variables
- **Database**: `POSTGRES_*` variables for connection
- **AWS**: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `S3_BUCKET_NAME`
- **Auth**: `JWT_SECRET`, `JWT_EXPIRATION`
- **YouTube**: `YOUTUBE_API_KEY`
- **Server**: `PORT`, timeout configurations
- **Metrics**: `ENABLE_METRICS`, `METRICS_PORT`

### Config Loading
- Attempts to load `.env` file from multiple locations
- Falls back to system environment variables
- Provides sensible defaults for development

## API Endpoints

### Radio Control
- `GET /api/v1/radio/status` - Current playback status
- `POST /api/v1/radio/play` - Start playback
- `POST /api/v1/radio/pause` - Pause playback
- `POST /api/v1/radio/next` - Skip to next song
- `POST /api/v1/radio/shuffle` - Toggle shuffle mode

### Playlists
- `GET /api/v1/playlists` - List playlists
- `POST /api/v1/playlists` - Create playlist
- `GET /api/v1/playlists/{id}` - Get playlist details
- `PUT /api/v1/playlists/{id}` - Update playlist
- `DELETE /api/v1/playlists/{id}` - Delete playlist

### YouTube Integration
- `POST /api/v1/youtube/add` - Add YouTube video
- `GET /api/v1/youtube/search` - Search YouTube

### Authentication
- JWT-based authentication system
- Admin routes protected by auth middleware

### WebSocket
- `/ws` - Real-time updates for playback, queue, reactions

## Development Workflow

### Build Commands (Makefile)
- `make build` - Build Go binary
- `make run` - Build and run server
- `make test` - Run Go tests
- `make test-coverage` - Run tests with coverage
- `make clean` - Clean build artifacts

### Database Migrations
- `make migrate` - Generate new migration
- `make migrate-up` - Apply migrations
- `make migrate-down` - Rollback migrations

### Frontend Commands
- `yarn dev` - Start development server
- `yarn build` - Build for production
- `yarn lint` - Run ESLint

### Docker Development
- `docker-compose up` - Full stack with PostgreSQL
- Services: backend (8080), frontend (5173), postgres (5432)
- Health checks for all services
- Persistent volume for database

## Key Features

### Radio Service
- Automatic playback loop with configurable duration
- Song selection algorithms (random, least played)
- Playlist-based playback
- Real-time status broadcasting via WebSocket
- S3 integration for audio file streaming

### Frontend Features
- Real-time radio player with visualizations
- Playlist management interface
- YouTube video integration
- User reactions system
- Authentication with protected routes
- Responsive design with Tailwind CSS

### Testing
- Go unit tests for core services
- Test coverage reporting
- Comprehensive service interfaces for mocking

## Deployment

### Fly.io Configuration
- Application: `go-radio-v2`
- Region: `iad` (US East)
- VM: 1 CPU, 1GB RAM
- Health checks on `/api/v1/health`
- Metrics endpoint on port 9090
- Volume mount for persistent data

### Docker Support
- Multi-stage builds for production
- Separate Dockerfiles for backend and frontend
- Production-ready nginx configuration for frontend

## Development Tips

### Testing
- Run `make test` for backend tests
- Service interfaces enable easy mocking
- Test configuration uses 5-second song duration for faster testing

### Local Development
1. Use `docker-compose up` for full stack
2. Backend runs on port 8080
3. Frontend runs on port 5173
4. PostgreSQL on port 5432

### Database Management
- Atlas handles schema migrations
- Local environment uses Docker PostgreSQL
- Schema defined in HCL format for type safety

### Code Organization
- Clean architecture with clear separation of concerns
- Interface-based design for testability
- Consistent error handling and logging
- Environment-based configuration management