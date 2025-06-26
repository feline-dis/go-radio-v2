# Event Bus Architecture

## Overview

This document describes the event-driven architecture implemented to solve the circular dependency issue between the `RadioService` and `WebSocketHandler`.

## Problem

Previously, there was a circular dependency:
- `RadioService` depended on `WebSocketHandlerInterface` to broadcast song changes
- `WebSocketHandler` depended on `RadioServiceInterface` to get playback state

This created tight coupling and made testing difficult.

## Solution: Event-Driven Architecture

We implemented an event bus pattern that decouples the services:

### Components

1. **EventBus** (`internal/events/event_bus.go`)
   - Central event management system
   - Handles event subscription and publishing
   - Provides type-safe event publishing methods

2. **Events**
   - `SongChangeEvent`: Published when a song changes
   - `QueueUpdateEvent`: Published when the queue updates
   - `PlaybackUpdateEvent`: Published for real-time playback updates

3. **Event Handlers**
   - `WebSocketHandler`: Subscribes to events and broadcasts to clients
   - Future handlers can easily be added (logging, metrics, etc.)

### Architecture Flow

```
RadioService → EventBus → WebSocketHandler → Clients
     ↓           ↓              ↓
  Publishes   Routes        Subscribes
  Events      Events        to Events
```

### Benefits

1. **Decoupling**: Services no longer directly depend on each other
2. **Testability**: Easy to mock the event bus for testing
3. **Extensibility**: New event handlers can be added without modifying existing code
4. **Maintainability**: Clear separation of concerns
5. **Reliability**: Panic recovery in event handlers

### Usage

#### Publishing Events

```go
// In RadioService
eventBus.PublishSongChange(currentSong, nextSong)
eventBus.PublishQueueUpdate(queueInfo)
```

#### Subscribing to Events

```go
// In WebSocketHandler
eventBus.Subscribe(events.EventSongChange, handler.handleSongChangeEvent)
eventBus.Subscribe(events.EventQueueUpdate, handler.handleQueueUpdateEvent)
```

#### Adding New Event Types

1. Add event type constant in `internal/events/event_bus.go`
2. Create event struct
3. Add publishing method to EventBus
4. Subscribe handlers as needed

### Testing

The event bus is fully testable with mock implementations:

```go
type MockEventBus struct{}

func (m *MockEventBus) PublishSongChange(currentSong, nextSong *models.Song) {
    // Mock implementation for tests
}
```

### Future Enhancements

1. **Event Persistence**: Store events for replay/debugging
2. **Event Filtering**: Subscribe to specific event attributes
3. **Event Metrics**: Track event frequency and performance
4. **Event Validation**: Validate event payloads
5. **Event Versioning**: Handle event schema changes

## Migration Notes

The migration from the old architecture was straightforward:

1. Created the event bus system
2. Updated `RadioService` to use event bus instead of direct WebSocket calls
3. Updated `WebSocketHandler` to subscribe to events
4. Updated dependency injection in `main.go`
5. Updated tests to use mock event bus

All existing functionality is preserved while improving the architecture. 