package events

import (
	"log"
	"sync"
	"time"

	"github.com/feline-dis/go-radio-v2/internal/models"
)

// Event types
const (
	EventSongChange     = "song_change"
	EventQueueUpdate    = "queue_update"
	EventPlaybackUpdate = "playback_update"
	EventUserReaction   = "user_reaction"
)

// Event represents a generic event
type Event struct {
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}

// SongChangeEvent represents a song change event
type SongChangeEvent struct {
	CurrentSong *models.Song     `json:"current_song"`
	NextSong    *models.Song     `json:"next_song"`
	Queue       []*models.Song   `json:"queue"`
	Playlist    *models.Playlist `json:"playlist"`
	Remaining   float64          `json:"remaining"`
	StartTime   time.Time        `json:"start_time"`
	Timestamp   int64            `json:"timestamp"`
}

// QueueUpdateEvent represents a queue update event
type QueueUpdateEvent struct {
	CurrentSong *models.Song     `json:"current_song"`
	NextSong    *models.Song     `json:"next_song"`
	Queue       []*models.Song   `json:"queue"`
	Playlist    *models.Playlist `json:"playlist"`
	Remaining   float64          `json:"remaining"`
	StartTime   time.Time        `json:"start_time"`
}

// PlaybackUpdateEvent represents a playback update event
type PlaybackUpdateEvent struct {
	Song      *models.Song `json:"song"`
	Elapsed   float64      `json:"elapsed"`
	Remaining float64      `json:"remaining"`
	Paused    bool         `json:"paused"`
	TotalTime float64      `json:"total_time"`
	Timestamp int64        `json:"timestamp"`
}

// UserReactionEvent represents a user reaction event
type UserReactionEvent struct {
	UserID    string `json:"user_id"`
	Emote     string `json:"emote"`
	Timestamp int64  `json:"timestamp"`
}

// EventHandler is a function that handles events
type EventHandler func(event Event)

// EventBus manages event subscriptions and publishing
type EventBus struct {
	handlers map[string][]EventHandler
	mu       sync.RWMutex
}

// NewEventBus creates a new event bus
func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[string][]EventHandler),
	}
}

// Subscribe registers a handler for a specific event type
func (eb *EventBus) Subscribe(eventType string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.handlers[eventType] == nil {
		eb.handlers[eventType] = make([]EventHandler, 0)
	}
	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

// Publish sends an event to all registered handlers
func (eb *EventBus) Publish(event Event) {
	eb.mu.RLock()
	handlers := make([]EventHandler, len(eb.handlers[event.Type]))
	copy(handlers, eb.handlers[event.Type])
	eb.mu.RUnlock()

	for _, handler := range handlers {
		go func(h EventHandler, e Event) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[ERROR] EventBus: Handler panicked: %v", r)
				}
			}()
			h(e)
		}(handler, event)
	}
}

// PublishSongChange publishes a song change event
func (eb *EventBus) PublishSongChange(currentSong, nextSong *models.Song, queueInfo *models.QueueInfo) {
	event := Event{
		Type: EventSongChange,
		Payload: SongChangeEvent{
			CurrentSong: currentSong,
			NextSong:    nextSong,
			Queue:       queueInfo.Queue,
			Playlist:    queueInfo.Playlist,
			Remaining:   queueInfo.Remaining,
			StartTime:   queueInfo.StartTime,
			Timestamp:   time.Now().UnixMilli(),
		},
		Timestamp: time.Now(),
	}
	eb.Publish(event)
}

// PublishQueueUpdate publishes a queue update event
func (eb *EventBus) PublishQueueUpdate(queueInfo *models.QueueInfo) {
	if queueInfo == nil {
		return
	}

	event := Event{
		Type: EventQueueUpdate,
		Payload: QueueUpdateEvent{
			CurrentSong: queueInfo.CurrentSong,
			NextSong:    queueInfo.NextSong,
			Queue:       queueInfo.Queue,
			Playlist:    queueInfo.Playlist,
			Remaining:   queueInfo.Remaining,
			StartTime:   queueInfo.StartTime,
		},
		Timestamp: time.Now(),
	}
	eb.Publish(event)
}

// PublishPlaybackUpdate publishes a playback update event
func (eb *EventBus) PublishPlaybackUpdate(song *models.Song, elapsed, remaining float64, paused bool) {
	event := Event{
		Type: EventPlaybackUpdate,
		Payload: PlaybackUpdateEvent{
			Song:      song,
			Elapsed:   elapsed,
			Remaining: remaining,
			Paused:    paused,
			TotalTime: float64(song.Duration),
			Timestamp: time.Now().UnixMilli(),
		},
		Timestamp: time.Now(),
	}
	eb.Publish(event)
}

// PublishUserReaction publishes a user reaction event
func (eb *EventBus) PublishUserReaction(userID, emote string) {
	event := Event{
		Type: EventUserReaction,
		Payload: UserReactionEvent{
			UserID:    userID,
			Emote:     emote,
			Timestamp: time.Now().UnixMilli(),
		},
		Timestamp: time.Now(),
	}
	eb.Publish(event)
}
