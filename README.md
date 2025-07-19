# Go Radio v2

A self-hosted streaming radio application built with Go and React. Host your own radio service with local file storage or cloud storage, featuring real-time WebSocket communication, YouTube integration, and optional public tunneling.

## Features

- **Self-hosted** - Run your own radio service locally
- **Flexible Storage** - Choose between local files or AWS S3 for audio storage
- **Multiple Database Options** - SQLite (default) or JSON files for metadata
- **Real-time Streaming** - WebSocket-based live radio with automatic song transitions
- **YouTube Integration** - Download and stream audio YouTube videos
- **Public Access** - Optional ngrok tunneling for external access
- **Modern UI** - React frontend with real-time visualizations
- **Easy Setup** - TUI-based configuration wizard

## Tech Stack

### Backend
- **Go 1.21+** - Main application server
- **Gorilla Mux** - HTTP router
- **Gorilla WebSocket** - Real-time communication
- **SQLite** - Local database (default) or JSON files
- **AWS SDK** - Optional S3 integration
- **Ngrok** - Optional public tunneling
- **JWT** - Authentication

### Frontend
- **React 19** - UI framework
- **TypeScript** - Type safety
- **Vite** - Build tool
- **Tailwind CSS** - Styling
- **TanStack Query** - Server state management
- **Audio Visualizations** - Butterchurn and custom visualizers

## Quick Start

### Prerequisites

- **Go 1.21+** - Download from [golang.org](https://golang.org/dl/)
- **Node.js 18+** - Download from [nodejs.org](https://nodejs.org/)
- **Git** - For cloning the repository

### Installation

1. **Clone and setup**
   ```bash
   git clone https://github.com/feline-dis/go-radio-v2.git
   cd go-radio-v2
   make setup
   ```

2. **Configure your radio** (optional - has sensible defaults)
   ```bash
   make config
   ```
   This opens a TUI wizard to configure:
   - Storage backend (local files or S3)
   - Data directory location
   - Metadata storage (SQLite or JSON)
   - Public tunneling (ngrok)

3. **Start your radio**
   ```bash
   make run
   ```

4. **Access your radio**
   - Local: http://localhost:8080
   - Public (if tunneling enabled): Your ngrok URL

### One-command setup
```bash
make start  # Complete setup and start
```

### Quick Start (Minimal)
```bash
git clone https://github.com/feline-dis/go-radio-v2.git
cd go-radio-v2
go mod download
cd client && yarn install && yarn build && cd ..
go build -o bin/go-radio-server ./cmd/server
./bin/go-radio-server
```

## Storage Options

### File Storage
- **Local Files** (Default) - Store audio files in a local directory
- **AWS S3** - Store audio files in Amazon S3 bucket

### Metadata Storage
- **SQLite** (Default) - Single file database, recommended for most users
- **JSON Files** - Store metadata in JSON files for simple setups

### Configuration Examples

**Local Setup (Default)**
```bash
FILE_STORAGE_TYPE=local
LOCAL_DATA_DIR=./data
METADATA_STORAGE_TYPE=sqlite
SQLITE_DB_PATH=./data/radio.db
```

**Cloud Setup**
```bash
FILE_STORAGE_TYPE=s3
METADATA_STORAGE_TYPE=sqlite
AWS_ACCESS_KEY_ID=your_key
AWS_SECRET_ACCESS_KEY=your_secret
S3_BUCKET_NAME=your_bucket
```

## Public Access

Enable public access to your radio using ngrok:

```bash
TUNNEL_ENABLED=true
TUNNEL_PROVIDER=ngrok
NGROK_AUTH_TOKEN=your_token  # Optional, for custom domains
NGROK_DOMAIN=your_domain     # Optional custom domain
```

## API Endpoints

### Radio Control
- `GET /api/v1/health` - Health check
- `GET /api/v1/now-playing` - Get currently playing song
- `GET /api/v1/queue` - Get current queue and playback state
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
- `GET /api/v1/playlists/{youtube_id}/file` - Stream audio file (legacy)

### YouTube Integration
- `POST /api/v1/youtube/add` - Add YouTube video to playlist
- `GET /api/v1/youtube/search` - Search YouTube videos

### WebSocket
- `WS /ws` - Real-time updates for playback status, queue changes, and reactions

### Authentication
All admin endpoints require JWT authentication via the `Authorization` header.

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `FILE_STORAGE_TYPE` | File storage backend (`local` or `s3`) | `local` |
| `LOCAL_DATA_DIR` | Local data directory | `./data` |
| `METADATA_STORAGE_TYPE` | Metadata storage (`sqlite` or `json`) | `sqlite` |
| `SQLITE_DB_PATH` | SQLite database file path | `./data/radio.db` |
| `TUNNEL_ENABLED` | Enable public tunneling | `false` |
| `NGROK_AUTH_TOKEN` | Ngrok authentication token | - |
| `JWT_SECRET` | JWT signing secret | Required |
| `YOUTUBE_API_KEY` | YouTube API key | Required |
| `AWS_ACCESS_KEY_ID` | AWS access key (if using S3) | - |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key (if using S3) | - |
| `S3_BUCKET_NAME` | S3 bucket name (if using S3) | - |

### Directory Structure

```
go-radio-v2/
├── data/                   # Local data directory
│   ├── audio/             # Audio files (if using local storage)
│   │   └── songs/         # Individual song files (youtube_id.mp3)
│   └── radio.db           # SQLite database
├── bin/                   # Built binaries
│   ├── go-radio-server    # Main server
│   └── go-radio-setup     # Setup wizard (TUI)
├── client/                # React frontend
│   ├── dist/              # Built frontend files
│   ├── src/               # Source code
│   └── package.json       # Node.js dependencies
├── cmd/                   # Go applications
│   ├── server/            # Main server application
│   └── setup/             # TUI setup wizard
├── internal/              # Go backend code
│   ├── storage/           # Storage abstractions and implementations
│   │   ├── interfaces.go  # Storage interfaces
│   │   ├── sqlite_*.go    # SQLite implementations
│   │   ├── local_*.go     # Local file storage
│   │   └── s3_*.go        # S3 storage implementations
│   ├── services/          # Business logic layer
│   ├── controllers/       # HTTP route handlers
│   ├── models/            # Data models
│   └── config/            # Configuration management
├── scripts/               # Setup and build scripts
│   └── setup.sh           # Automated setup script
├── .env                   # Environment configuration
└── .env.example           # Example configuration
```

## Development

### Commands

```bash
# Full setup
make setup

# Install dependencies only
make dev-deps

# Build applications
make build

# Build frontend only
make build-frontend

# Run configuration wizard
make config

# Start development server
make dev

# Run tests
make test

# Clean build artifacts
make clean
```

### Manual Setup

If you prefer manual setup:

```bash
# Install Go dependencies
go mod download

# Install frontend dependencies
cd client && yarn install

# Build frontend
cd client && yarn build

# Build backend
go build -o bin/go-radio-server cmd/server/main.go

# Run setup wizard
go run cmd/setup/main.go

# Start server
./bin/go-radio-server
```

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   React Client  │    │   Go Backend    │    │ Local Storage   │
│                 │    │                 │    │                 │
│ - Radio Player  │◄──►│ - HTTP API      │◄──►│ - SQLite/JSON   │
│ - Playlist Mgmt │    │ - WebSocket     │    │ - Audio Files   │
│ - Real-time UI  │    │ - Radio Service │    │ - Metadata      │
│ - Visualizations│    │ - Tunnel Service│    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │                         │
                              ▼                         ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │   Ngrok Tunnel  │    │   AWS S3        │
                       │                 │    │   (Optional)    │
                       │ - Public Access │    │ - Audio Files   │
                       │ - Custom Domain │    │ - Cloud Storage │
                       └─────────────────┘    └─────────────────┘
```

## Use Cases

### Personal Radio
- Host your music collection locally
- Access from anywhere with tunneling
- Share with friends and family

### Small Organization
- Internal radio for office/team
- Shared playlists and music discovery
- Real-time listening together

## Troubleshooting

### Common Issues

**"No playlists found" error**
```bash
# Create a test playlist
curl -X POST http://localhost:8080/api/v1/playlists \
  -H "Content-Type: application/json" \
  -d '{"name": "My Playlist", "description": "Test playlist", "song_ids": []}'
```

**Database locked error (SQLite)**
```bash
# Stop the server and remove database lock
rm ./data/radio.db-wal ./data/radio.db-shm
```

**Permission denied on data directory**
```bash
# Fix directory permissions
chmod 755 ./data
chmod 644 ./data/radio.db
```

**Frontend not loading**
```bash
# Rebuild frontend
cd client && yarn build && cd ..
```

**ngrok authentication error**
```bash
# Sign up at ngrok.com and get auth token
ngrok authtoken YOUR_TOKEN
```

**Port already in use**
```bash
# Change port in .env file
PORT=8081
```

**Audio files not playing**
```bash
# Check audio file format and location
ls -la ./data/audio/songs/
file ./data/audio/songs/*.mp3

# Verify endpoint responds
curl -I http://localhost:8080/api/v1/songs/YOUTUBE_ID/file
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests if applicable
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Gorilla Toolkit](https://github.com/gorilla) for WebSocket and HTTP routing
- [Ngrok](https://ngrok.com/) for tunneling capabilities
- [Charm](https://charm.sh/) for beautiful TUI components
