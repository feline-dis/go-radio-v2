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
- SQLite
- Atlas (for database migrations)

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

4. Run database migrations:
   ```bash
   make migrate-up
   ```

5. Run the application:
   ```bash
   make dev
   ```

## Development

- Run tests: `make test`
- Run tests with coverage: `make test-coverage`
- Run linter: `make lint`
- Build binary: `make build`
- Run migrations: `make migrate-up`

## Deployment

1. Build and push Docker image:
   ```bash
   make docker-build
   make docker-push
   ```

2. Deploy to fly.io:
   ```bash
   make deploy
   ```

## Database Schema

The application uses SQLite for storing metadata. The main tables are:

- `songs`: Stores song metadata (title, artist, duration, etc.)
- `playlists`: Stores playlist information (name, description)
- `playlist_songs`: Manages the many-to-many relationship between playlists and songs

## License

MIT

## API Endpoints

### Public Endpoints
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

## Prerequisites

- Docker
- AWS S3 bucket
- fly.io account (for deployment)

## Project Structure

```
go-radio/
├── cmd/
│   ├── server/        # Main application entry point
│   └── migrate/       # Database migration tool
├── internal/
│   ├── config/        # Configuration management
│   ├── controllers/   # HTTP handlers
│   │   └── reaction_controller.go  # Reaction API endpoints
│   ├── events/        # Event bus system
│   ├── middleware/    # HTTP middleware
│   ├── models/        # Data models
│   ├── repositories/  # Data access layer
│   ├── services/      # Business logic
│   └── websocket/     # WebSocket handling
├── client/
│   ├── src/
│   │   ├── components/
│   │   │   ├── ReactionBar.tsx     # Reaction selection UI
│   │   │   └── AnimatedEmotes.tsx  # Reaction display
│   │   └── contexts/
│   │       └── ReactionContext.tsx # Reaction state management
│   └── ...
├── pkg/
│   ├── logger/        # Logging utilities
│   └── utils/         # Common utilities
├── migrations/        # SQLite migrations
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Getting Started

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/go-radio.git
   cd go-radio
   ```

2. Install dependencies:
   ```bash
   make deps
   ```

3. Set up environment variables:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. Run database migrations:
   ```bash
   make migrate-up
   ```

5. Run the application:
   ```bash
   make dev
   ```

## License

MIT