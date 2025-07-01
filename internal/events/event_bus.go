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
	EventSkip           = "skip"
	EventPrevious       = "previous"
	EventPlaylistChange = "playlist_change"
)

// Event represents a generic event
type Event struct {
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}

// SongChangeEvent represents a song change event
type SongChangeEvent struct {
	CurrentSong      *models.Song     `json:"current_song"`
	NextSong         *models.Song     `json:"next_song"`
	Queue            []*models.Song   `json:"queue"`
	Playlist         *models.Playlist `json:"playlist"`
	Remaining        float64          `json:"remaining"`
	StartTime        time.Time        `json:"start_time"`
	Timestamp        int64            `json:"timestamp"`
	CurrentSongIndex int              `json:"current_song_index"`
}

// QueueUpdateEvent represents a queue update event
type QueueUpdateEvent struct {
	CurrentSong      *models.Song     `json:"current_song"`
	NextSong         *models.Song     `json:"next_song"`
	Queue            []*models.Song   `json:"queue"`
	Playlist         *models.Playlist `json:"playlist"`
	Remaining        float64          `json:"remaining"`
	StartTime        time.Time        `json:"start_time"`
	CurrentSongIndex int              `json:"current_song_index"`
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

// SkipEvent represents a skip event
type SkipEvent struct {
	Song      *models.Song          `json:"song"`
	NextSong  *models.Song          `json:"next_song"`
	State     *models.PlaybackState `json:"state"`
	Timestamp int64                 `json:"timestamp"`
}

// PreviousEvent represents a previous event
type PreviousEvent struct {
	Song      *models.Song          `json:"song"`
	NextSong  *models.Song          `json:"next_song"`
	State     *models.PlaybackState `json:"state"`
	Timestamp int64                 `json:"timestamp"`
}

// PlaylistChangeEvent represents a playlist change event
type PlaylistChangeEvent struct {
	Song      *models.Song          `json:"song"`
	NextSong  *models.Song          `json:"next_song"`
	Playlist  *models.Playlist      `json:"playlist"`
	State     *models.PlaybackState `json:"state"`
	Timestamp int64                 `json:"timestamp"`
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
			CurrentSong:      currentSong,
			NextSong:         nextSong,
			Queue:            queueInfo.Queue,
			Playlist:         queueInfo.Playlist,
			Remaining:        queueInfo.Remaining,
			StartTime:        queueInfo.StartTime,
			CurrentSongIndex: queueInfo.CurrentSongIndex,
			Timestamp:        time.Now().UnixMilli(),
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

	// Get current and next songs safely
	var currentSong, nextSong *models.Song
	if len(queueInfo.Queue) > 0 && queueInfo.CurrentSongIndex >= 0 && queueInfo.CurrentSongIndex < len(queueInfo.Queue) {
		currentSong = queueInfo.Queue[queueInfo.CurrentSongIndex]
		nextIndex := (queueInfo.CurrentSongIndex + 1) % len(queueInfo.Queue)
		if nextIndex < len(queueInfo.Queue) {
			nextSong = queueInfo.Queue[nextIndex]
		}
	}

	event := Event{
		Type: EventQueueUpdate,
		Payload: QueueUpdateEvent{
			CurrentSong:      currentSong,
			NextSong:         nextSong,
			Queue:            queueInfo.Queue,
			Playlist:         queueInfo.Playlist,
			Remaining:        queueInfo.Remaining,
			StartTime:        queueInfo.StartTime,
			CurrentSongIndex: queueInfo.CurrentSongIndex,
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

// PublishSkip publishes a skip event
func (eb *EventBus) PublishSkip(song *models.Song, nextSong *models.Song, state *models.PlaybackState) {
	event := Event{
		Type: EventSkip,
		Payload: SkipEvent{
			Song:      song,
			NextSong:  nextSong,
			State:     state,
			Timestamp: time.Now().UnixMilli(),
		},
		Timestamp: time.Now(),
	}
	eb.Publish(event)
}

// PublishPrevious publishes a previous event
func (eb *EventBus) PublishPrevious(song *models.Song, nextSong *models.Song, state *models.PlaybackState) {
	event := Event{
		Type: EventPrevious,
		Payload: PreviousEvent{
			Song:      song,
			NextSong:  nextSong,
			State:     state,
			Timestamp: time.Now().UnixMilli(),
		},
		Timestamp: time.Now(),
	}
	eb.Publish(event)
}

// PublishPlaylistChange publishes a playlist change event
func (eb *EventBus) PublishPlaylistChange(song *models.Song, nextSong *models.Song, playlist *models.Playlist, state *models.PlaybackState) {
	event := Event{
		Type: EventPlaylistChange,
		Payload: PlaylistChangeEvent{
			Song:      song,
			NextSong:  nextSong,
			Playlist:  playlist,
			State:     state,
			Timestamp: time.Now().UnixMilli(),
		},
		Timestamp: time.Now(),
	}
	eb.Publish(event)
}
