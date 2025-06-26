# Reaction System Documentation

## Overview

The Go Radio V2 reaction system allows users to express themselves while listening to music together. Users can select from a variety of emotes that appear as animated overlays on all connected clients' screens.

## Architecture

### Backend Components

#### Event Bus Integration
- **File**: `internal/events/event_bus.go`
- **Event Type**: `EventUserReaction`
- **Payload**: `UserReactionEvent` with user_id, emote, and timestamp

#### WebSocket Handler
- **File**: `internal/websocket/handler.go`
- **Message Type**: `user_reaction`
- **Handling**: Broadcasts reaction events to all connected clients

#### API Controller
- **File**: `internal/controllers/reaction_controller.go`
- **Endpoint**: `POST /api/v1/reactions`
- **Validation**: Ensures all required fields are present

### Frontend Components

#### Reaction Context
- **File**: `client/src/contexts/ReactionContext.tsx`
- **Purpose**: Manages reaction state and WebSocket connection
- **Features**: Auto-cleanup, reconnection logic, connection status

#### Reaction Bar
- **File**: `client/src/components/ReactionBar.tsx`
- **Purpose**: UI for selecting and sending reactions
- **Features**: 8 emote buttons, connection status, user feedback

#### Animated Emotes
- **File**: `client/src/components/AnimatedEmotes.tsx`
- **Purpose**: Displays floating emote animations
- **Features**: Random positioning, auto-cleanup

## Emote System

### Supported Emotes
1. **Heart** (❤️) - `heart`
2. **Fire** (🔥) - `fire`
3. **Rocket** (🚀) - `rocket`
4. **Clap** (👏) - `clap`
5. **Dance** (💃) - `dance`
6. **Party** (🎉) - `party`
7. **Star** (⭐) - `star`
8. **Thumbs Up** (👍) - `thumbsup`

### Emote Mapping
```typescript
const EMOTE_MAP: Record<string, string> = {
  heart: "❤️",
  fire: "🔥",
  rocket: "🚀",
  clap: "👏",
  dance: "💃",
  party: "🎉",
  star: "⭐",
  thumbsup: "👍",
};
```

## Data Flow

### Sending a Reaction
1. User clicks emote button in ReactionBar
2. Component generates temporary user ID
3. `sendReaction()` function called with user ID and emote
4. WebSocket message sent to server: `{"type": "reaction", "user_id": "...", "emote": "..."}`
5. Server validates and publishes `EventUserReaction`
6. EventBus broadcasts to all connected clients
7. All clients receive `user_reaction` WebSocket message
8. AnimatedEmotes component displays floating emote

### Receiving a Reaction
1. WebSocket receives `user_reaction` message
2. ReactionContext processes payload
3. New reaction added to reactions state array
4. AnimatedEmotes component renders floating emote
5. CSS animation plays for 3 seconds
6. Automatic cleanup removes reaction from state

## Animation System

### CSS Animation
```css
@keyframes emote-float {
  0% {
    opacity: 0;
    transform: translateY(20px) scale(0.8);
  }
  20% {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
  80% {
    opacity: 1;
    transform: translateY(-40px) scale(1.1);
  }
  100% {
    opacity: 0;
    transform: translateY(-60px) scale(0.9);
  }
}
```

### Animation Features
- **Duration**: 3 seconds total
- **Positioning**: Random X (10-90%) and Y (20-80%) coordinates
- **Effects**: Fade in/out, scale, and vertical movement
- **Cleanup**: Automatic removal after animation completes

## API Reference

### WebSocket Messages

#### Send Reaction
```json
{
  "type": "reaction",
  "user_id": "user_abc123",
  "emote": "heart"
}
```

#### Receive Reaction
```json
{
  "type": "user_reaction",
  "payload": {
    "user_id": "user_abc123",
    "emote": "heart",
    "timestamp": 1703123456789
  }
}
```

### HTTP Endpoints

#### POST /api/v1/reactions
**Request Body:**
```json
{
  "user_id": "user_abc123",
  "emote": "heart"
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Reaction sent successfully"
}
```

## Configuration

### Frontend Configuration
- **Animation Duration**: 3 seconds (configurable in CSS)
- **Reconnection Delay**: 5 seconds
- **Max Reactions**: No limit (managed by cleanup)
- **Position Range**: X: 10-90%, Y: 20-80%

### Backend Configuration
- **WebSocket Buffer Size**: 100 messages
- **Event Handler Recovery**: Automatic panic recovery
- **Validation**: Required fields: user_id, emote

## Error Handling

### Frontend Errors
- **WebSocket Disconnected**: Show "DISCONNECTED - REACTIONS UNAVAILABLE"
- **Invalid Emote**: Fallback to thumbs up emoji
- **Animation Failures**: Graceful degradation
- **Memory Issues**: Automatic cleanup prevents memory leaks

### Backend Errors
- **Invalid Payload**: Return 400 Bad Request
- **Missing Fields**: Return 400 Bad Request
- **Event Bus Errors**: Log error, continue operation
- **WebSocket Errors**: Automatic reconnection

## Performance Considerations

### Frontend
- **Memory Management**: Automatic cleanup prevents memory leaks
- **Animation Performance**: CSS animations are GPU-accelerated
- **State Updates**: Efficient React state management
- **WebSocket Handling**: Non-blocking message processing

### Backend
- **Event Broadcasting**: Goroutines for non-blocking operations
- **Memory Usage**: Minimal memory footprint for reaction data
- **WebSocket Scaling**: Efficient client management
- **Event Processing**: Panic recovery ensures stability

## Testing

### Frontend Testing
- Test reaction sending and receiving
- Verify animation display and cleanup
- Test connection status handling
- Validate emote mapping
- Test multiple simultaneous reactions

### Backend Testing
- Test reaction API endpoint
- Verify WebSocket message handling
- Test event bus integration
- Validate payload validation
- Test multiple client scenarios

## Future Enhancements

### Potential Features
- **Custom Emotes**: User-defined emote support
- **Reaction History**: Persistent reaction logs
- **Emote Categories**: Grouped emote selection
- **Reaction Limits**: Rate limiting per user
- **Emote Effects**: Sound effects or haptic feedback
- **User Profiles**: Persistent user identification
- **Reaction Analytics**: Usage statistics and trends

### Technical Improvements
- **WebRTC Integration**: Direct peer-to-peer reactions
- **Optimistic Updates**: Immediate local reaction display
- **Batch Processing**: Group multiple reactions
- **Caching**: Cache frequently used emotes
- **Compression**: Optimize WebSocket payload size 