# Go Radio Client

A React-based web client for the Go Radio streaming service.

## Features

- **Synchronized Playback**: All users listen to the same song at the same time
- **Real-time Updates**: WebSocket connection for live song changes and queue updates
- **User Interaction Required**: Audio context initialization only happens after user interaction (browser requirement)
- **Volume Control**: Adjustable volume with mute functionality
- **Queue Display**: Shows current and upcoming songs
- **Responsive Design**: Works on desktop and mobile devices

## How It Works

### Initialization Flow

1. **Page Load**: Client connects to WebSocket and fetches queue information
2. **User Interaction**: User clicks "ENTER RADIO" button to initialize audio context
3. **Audio Setup**: Audio context is created and current song file is loaded
4. **Synchronized Playback**: Music starts playing from the correct position based on server start time
5. **Real-time Sync**: Client stays synchronized with server via WebSocket events

### Synchronization

- The client calculates elapsed time based on the server's `StartTime`
- WebSocket events notify clients of song changes
- Progress bar updates in real-time to match server playback
- Multiple users hear the same song at the same position

### Audio Handling

- Uses Web Audio API for high-quality audio playback
- Audio context is only initialized after user interaction (browser requirement)
- Volume changes are applied smoothly without audio artifacts
- Automatic cleanup when component unmounts

## Development

### Prerequisites

- Node.js 18+
- Yarn or npm

### Setup

```bash
cd client
yarn install
```

### Development Server

```bash
yarn dev
```

The client will connect to `ws://localhost:8080/ws` in development mode.

### Production Build

```bash
yarn build
```

## Architecture

### Components

- **RadioProvider**: Context provider managing audio state and WebSocket connection
- **RadioInitButton**: Initial entry point requiring user interaction
- **RadioPlayer**: Main player interface with controls and queue display

### State Management

- **React Context**: Centralized state management for radio functionality
- **React Query**: Server state management for queue and song data
- **WebSocket**: Real-time communication with server

### Key Features

- **User Interaction Detection**: Ensures audio context is only initialized after user interaction
- **Server Time Synchronization**: Calculates playback position based on server start time
- **Song Change Handling**: Automatically switches to new songs when server changes tracks
- **Error Handling**: Graceful handling of connection issues and audio errors
- **Cleanup**: Proper cleanup of audio resources and WebSocket connections

## Browser Compatibility

- Modern browsers with Web Audio API support
- Requires user interaction before audio playback (browser security requirement)
- WebSocket support for real-time updates
