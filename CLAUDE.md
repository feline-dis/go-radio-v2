# Go Radio v2 - Claude Context

## Project Overview
Go Radio v2 is a self-hosted streaming radio application built with Go backend and React frontend. It features real-time WebSocket communication, YouTube integration, flexible storage options (local files or S3), SQLite database, and a comprehensive playlist management system with TUI-based setup.

## Architecture

### Backend (Go)
- **Main Server**: `cmd/server/main.go`
- **Setup Wizard**: `cmd/setup/main.go` (TUI-based configuration)
- **Config Management**: `internal/config/config.go` (supports local and cloud storage)
- **Clean Architecture**: Controllers → Services → Storage Interfaces → Implementations
- **Database**: SQLite (default) or JSON files for metadata
- **File Storage**: Local filesystem (default) or AWS S3
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
- `mattn/go-sqlite3` - SQLite database driver
- `aws-sdk-go-v2` - AWS S3 integration (optional)
- `golang-jwt/jwt/v5` - JWT authentication
- `joho/godotenv` - Environment variable management
- `charmbracelet/bubbletea` - TUI framework for setup wizard
- `charmbracelet/lipgloss` - TUI styling
- `google/uuid` - UUID generation

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
│   └── setup/main.go           # TUI-based setup wizard
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
│   ├── repositories/           # Legacy data access layer (PostgreSQL)
│   │   ├── playlist_repository.go
│   │   └── song_repository.go
│   ├── storage/                # New storage abstraction layer
│   │   ├── interfaces.go       # Storage interfaces
│   │   ├── factory.go          # Storage factory for creating implementations
│   │   ├── sqlite_song_repository.go      # SQLite song storage
│   │   ├── sqlite_playlist_repository.go  # SQLite playlist storage
│   │   ├── local_file_storage.go          # Local file storage
│   │   └── s3_file_storage.go             # S3 file storage
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
├── data/                       # Local data directory
│   ├── audio/                  # Audio files (local storage)
│   │   └── songs/              # Individual song files
│   └── radio.db                # SQLite database
├── scripts/                    # Setup and utility scripts
│   └── setup.sh                # Automated setup script
├── migrations/                 # Legacy database migrations (PostgreSQL)
├── schema.hcl                  # Legacy Atlas database schema
├── atlas.hcl                   # Legacy Atlas configuration
├── docker-compose.yml          # Local development setup
├── Dockerfile                  # Backend container
├── Makefile                    # Build automation
└── fly.toml                    # Fly.io deployment config
```

## Database Schema (SQLite)

### Tables
1. **songs** - Core song metadata
   - `youtube_id` (PK) - YouTube video identifier
   - `title`, `artist`, `album` - Metadata
   - `duration` - Song length in seconds
   - `file_path` - File location (local path or S3 key)
   - `last_played`, `play_count` - Usage tracking
   - `created_at`, `updated_at` - Timestamps
   - Indexes on `play_count` and `last_played`

2. **playlists** - Playlist definitions
   - `id` (UUID PK) - Unique identifier
   - `name` (unique), `description` - Playlist info
   - `created_at`, `updated_at` - Timestamps

3. **playlist_songs** - Many-to-many relationship
   - `playlist_id`, `youtube_id` (composite PK)
   - `position` - Song order in playlist
   - `created_at` - Timestamp
   - Foreign keys with CASCADE delete

## Configuration

### Environment Variables
- **Storage**: `FILE_STORAGE_TYPE` (local/s3), `LOCAL_DATA_DIR`, `METADATA_STORAGE_TYPE` (sqlite/json)
- **Database**: `SQLITE_DB_PATH` for SQLite database location
- **AWS**: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `S3_BUCKET_NAME` (if using S3)
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
- `GET /api/v1/health` - Health check
- `GET /api/v1/now-playing` - Current song with timing
- `GET /api/v1/queue` - Current queue and playback state
- `GET /api/v1/debug/playback-state` - Debug playback information
- `POST /api/v1/admin/skip` - Skip to next song (admin)
- `POST /api/v1/admin/previous` - Go to previous song (admin)
- `POST /api/v1/admin/playlist/set-active` - Set active playlist (admin)

### Playlists
- `GET /api/v1/playlists` - List all playlists
- `POST /api/v1/playlists` - Create new playlist
- `GET /api/v1/playlists/{id}` - Get playlist details
- `GET /api/v1/playlists/{id}/songs` - Get songs in playlist
- `PUT /api/v1/playlists/{id}` - Update playlist
- `DELETE /api/v1/playlists/{id}` - Delete playlist

### Audio Files
- `GET /api/v1/songs/{youtube_id}/file` - Stream audio file
- `GET /api/v1/playlists/{youtube_id}/file` - Stream audio file (legacy endpoint)

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
- `make setup` - Complete setup (dependencies + build + TUI config)
- `make config` - Run TUI configuration wizard
- `make build` - Build Go binaries (server + setup)
- `make build-frontend` - Build React frontend
- `make run` - Build and run server
- `make dev` - Start development server
- `make test` - Run Go tests
- `make test-coverage` - Run tests with coverage
- `make clean` - Clean build artifacts
- `make start` - Complete setup and start (one command)

### Frontend Commands
- `yarn dev` - Start development server
- `yarn build` - Build for production
- `yarn lint` - Run ESLint

### Storage Management
- **Local Storage**: Files stored in `./data/audio/songs/`
- **SQLite Database**: Single file at `./data/radio.db`
- **Configuration**: Environment variables in `.env` file
- **Setup**: TUI wizard at `./bin/go-radio-setup`

### Docker Development (Legacy)
- `docker-compose up` - Full stack with PostgreSQL (legacy)
- Services: backend (8080), frontend (5173), postgres (5432)
- Health checks for all services
- Persistent volume for database

## Key Features

### Radio Service
- Automatic playback loop with configurable duration
- Song selection algorithms (random, least played)
- Playlist-based playback with shuffle support
- Real-time status broadcasting via WebSocket
- Local file streaming or S3 integration for audio files
- SQLite-based metadata storage with play count tracking

### Frontend Features
- Real-time radio player with audio visualizations (Butterchurn, Vissonance)
- Playlist management interface with drag-and-drop
- YouTube video integration and search
- User reactions system with real-time WebSocket updates
- Authentication with JWT and protected routes
- Responsive design with Tailwind CSS
- Audio streaming with Web Audio API integration

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
1. Run `make setup` for complete setup with TUI configuration
2. Backend runs on port 8080 with SQLite database
3. Frontend built files served by backend (React SPA)
4. External access via third-party tunneling services (see docs/TUNNELING.md)
5. Audio files stored in `./data/audio/songs/`

### Quick Setup for Testing
1. `make setup` - Full automated setup
2. `make config` - TUI configuration wizard
3. `make run` - Start the radio server
4. Visit http://localhost:8080 to access the radio

### Database Management
- SQLite database automatically created on first run
- Schema creation handled by storage interfaces
- Database file stored at `./data/radio.db`
- Simple SQL commands for debugging and testing

### Code Organization
- Clean architecture with clear separation of concerns
- Interface-based design for testability and multiple storage backends
- Consistent error handling and logging
- Environment-based configuration management
- Storage abstraction layer supporting local files and cloud storage

## Current Implementation Status (2025-07-19)

✅ **Completed & Working:**
- Self-hosted deployment with local SQLite storage
- TUI-based setup wizard with Charm Bubbletea
- Audio file streaming via `/api/v1/songs/{id}/file` endpoint
- Real-time radio playback with automatic song transitions
- WebSocket integration for live updates
- React frontend with audio visualizations
- External access via tunneling services (documentation provided)
- Comprehensive playlist and song management APIs
- Local file storage with configurable data directory

🔧 **Configuration:**
- Default storage: SQLite + Local files
- Optional: S3 + Cloud storage
- Environment variables with sensible defaults
- Automatic setup via `make setup` command

🎵 **Audio Features:**
- MP3 file streaming with proper headers
- Play count and last played tracking
- Playlist-based playback with shuffle
- YouTube integration for adding songs
- Real-time queue management

⚙️ **Development Setup:**
- Single binary deployment (`go-radio-server`)
- Frontend built into static files served by backend
- No external database dependencies required
- Works completely offline (no cloud dependencies)