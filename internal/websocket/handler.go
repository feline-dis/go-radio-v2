package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/feline-dis/go-radio-v2/internal/events"
	"github.com/feline-dis/go-radio-v2/internal/models"
	"github.com/gorilla/websocket"
)

// RadioServiceInterface defines the methods we need from the radio service
type RadioServiceInterface interface {
	GetPlaybackState() *models.PlaybackState
	GetElapsedTime() time.Duration
	GetRemainingTime() time.Duration
	GetQueueInfo() *models.QueueInfo
	GetCurrentSong() *models.Song
}

// EventBusInterface defines the methods we need from the event bus
type EventBusInterface interface {
	Subscribe(eventType string, handler events.EventHandler)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

type Client struct {
	conn     *websocket.Conn
	send     chan []byte
	radioSvc RadioServiceInterface
	handler  *Handler
}

type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type PlaybackUpdate struct {
	Song      *models.Song `json:"song"`
	Elapsed   float64      `json:"elapsed"`
	Remaining float64      `json:"remaining"`
	Paused    bool         `json:"paused"`
	TotalTime float64      `json:"total_time"`
	Timestamp int64        `json:"timestamp"` // Unix timestamp for sync
}

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

type PlaybackControlEvent struct {
	Action    string                `json:"action"`
	Song      *models.Song          `json:"song"`
	State     *models.PlaybackState `json:"state"`
	Timestamp int64                 `json:"timestamp"`
}

type SkipEvent struct {
	Song      *models.Song          `json:"song"`
	NextSong  *models.Song          `json:"next_song"`
	State     *models.PlaybackState `json:"state"`
	Timestamp int64                 `json:"timestamp"`
}

type PreviousEvent struct {
	Song      *models.Song          `json:"song"`
	NextSong  *models.Song          `json:"next_song"`
	State     *models.PlaybackState `json:"state"`
	Timestamp int64                 `json:"timestamp"`
}

type PlaylistChangeEvent struct {
	Song      *models.Song          `json:"song"`
	NextSong  *models.Song          `json:"next_song"`
	Playlist  *models.Playlist      `json:"playlist"`
	State     *models.PlaybackState `json:"state"`
	Timestamp int64                 `json:"timestamp"`
}

type QueueUpdate struct {
	CurrentSong      *models.Song     `json:"current_song"`
	NextSong         *models.Song     `json:"next_song"`
	Queue            []*models.Song   `json:"queue"`
	Playlist         *models.Playlist `json:"playlist"`
	Remaining        float64          `json:"remaining"`
	StartTime        time.Time        `json:"start_time"`
	CurrentSongIndex int              `json:"current_song_index"`
}

type UserReactionEvent struct {
	UserID    string `json:"user_id"`
	Emote     string `json:"emote"`
	Timestamp int64  `json:"timestamp"`
}

type ClientRequest struct {
	Type string `json:"type"`
}

type ReactionRequest struct {
	Type   string `json:"type"`
	UserID string `json:"user_id"`
	Emote  string `json:"emote"`
}

type Handler struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	radioSvc   RadioServiceInterface
	eventBus   EventBusInterface
	mu         sync.RWMutex
}

func NewHandler(radioSvc RadioServiceInterface, eventBus EventBusInterface) *Handler {
	handler := &Handler{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 100), // Buffer for broadcast messages
		register:   make(chan *Client, 10), // Buffer for client registrations
		unregister: make(chan *Client, 10), // Buffer for client unregistrations
		radioSvc:   radioSvc,
		eventBus:   eventBus,
	}

	// Subscribe to events
	if eventBus != nil {
		eventBus.Subscribe(events.EventSongChange, handler.handleSongChangeEvent)
		eventBus.Subscribe(events.EventQueueUpdate, handler.handleQueueUpdateEvent)
		eventBus.Subscribe(events.EventUserReaction, handler.handleUserReactionEvent)
		eventBus.Subscribe(events.EventSkip, handler.handleSkipEvent)
		eventBus.Subscribe(events.EventPrevious, handler.handlePreviousEvent)
		eventBus.Subscribe(events.EventPlaylistChange, handler.handlePlaylistChangeEvent)
	}

	return handler
}

func (h *Handler) SetRadioService(radioSvc RadioServiceInterface) {
	h.radioSvc = radioSvc
}

// handleSongChangeEvent handles song change events from the event bus
func (h *Handler) handleSongChangeEvent(event events.Event) {
	songChangeEvent, ok := event.Payload.(events.SongChangeEvent)
	if !ok {
		log.Printf("[ERROR] handleSongChangeEvent: Failed to cast payload to SongChangeEvent")
		return
	}

	wsEvent := SongChangeEvent{
		CurrentSong:      songChangeEvent.CurrentSong,
		NextSong:         songChangeEvent.NextSong,
		Queue:            songChangeEvent.Queue,
		Playlist:         songChangeEvent.Playlist,
		Remaining:        songChangeEvent.Remaining,
		StartTime:        songChangeEvent.StartTime,
		Timestamp:        songChangeEvent.Timestamp,
		CurrentSongIndex: songChangeEvent.CurrentSongIndex,
	}

	message := Message{
		Type:    "song_change",
		Payload: wsEvent,
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("[ERROR] handleSongChangeEvent: Failed to marshal event: %v", err)
		return
	}

	h.broadcast <- data
}

// handleQueueUpdateEvent handles queue update events from the event bus
func (h *Handler) handleQueueUpdateEvent(event events.Event) {
	queueUpdateEvent, ok := event.Payload.(events.QueueUpdateEvent)
	if !ok {
		log.Printf("[ERROR] handleQueueUpdateEvent: Failed to cast payload to QueueUpdateEvent")
		return
	}

	wsEvent := QueueUpdate{
		CurrentSong:      queueUpdateEvent.CurrentSong,
		NextSong:         queueUpdateEvent.NextSong,
		Queue:            queueUpdateEvent.Queue,
		Playlist:         queueUpdateEvent.Playlist,
		Remaining:        queueUpdateEvent.Remaining,
		StartTime:        queueUpdateEvent.StartTime,
		CurrentSongIndex: queueUpdateEvent.CurrentSongIndex,
	}

	message := Message{
		Type:    "queue_update",
		Payload: wsEvent,
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("[ERROR] handleQueueUpdateEvent: Failed to marshal event: %v", err)
		return
	}

	h.broadcast <- data
}

// handleUserReactionEvent handles user reaction events from the event bus
func (h *Handler) handleUserReactionEvent(event events.Event) {
	reactionEvent, ok := event.Payload.(events.UserReactionEvent)
	if !ok {
		log.Printf("[ERROR] handleUserReactionEvent: Failed to cast payload to UserReactionEvent")
		return
	}

	wsEvent := UserReactionEvent{
		UserID:    reactionEvent.UserID,
		Emote:     reactionEvent.Emote,
		Timestamp: reactionEvent.Timestamp,
	}

	message := Message{
		Type:    "user_reaction",
		Payload: wsEvent,
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("[ERROR] handleUserReactionEvent: Failed to marshal event: %v", err)
		return
	}

	h.broadcast <- data
}

// handleSkipEvent handles skip events from the event bus
func (h *Handler) handleSkipEvent(event events.Event) {
	skipEvent, ok := event.Payload.(events.SkipEvent)
	if !ok {
		log.Printf("[ERROR] handleSkipEvent: Failed to cast payload to SkipEvent")
		return
	}

	wsEvent := SkipEvent{
		Song:      skipEvent.Song,
		NextSong:  skipEvent.NextSong,
		State:     skipEvent.State,
		Timestamp: skipEvent.Timestamp,
	}

	message := Message{
		Type:    "skip",
		Payload: wsEvent,
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("[ERROR] handleSkipEvent: Failed to marshal event: %v", err)
		return
	}

	h.broadcast <- data
}

// handlePreviousEvent handles previous events from the event bus
func (h *Handler) handlePreviousEvent(event events.Event) {
	previousEvent, ok := event.Payload.(events.PreviousEvent)
	if !ok {
		log.Printf("[ERROR] handlePreviousEvent: Failed to cast payload to PreviousEvent")
		return
	}

	wsEvent := PreviousEvent{
		Song:      previousEvent.Song,
		NextSong:  previousEvent.NextSong,
		State:     previousEvent.State,
		Timestamp: previousEvent.Timestamp,
	}

	message := Message{
		Type:    "previous",
		Payload: wsEvent,
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("[ERROR] handlePreviousEvent: Failed to marshal event: %v", err)
		return
	}

	h.broadcast <- data
}

// handlePlaylistChangeEvent handles playlist change events from the event bus
func (h *Handler) handlePlaylistChangeEvent(event events.Event) {
	playlistChangeEvent, ok := event.Payload.(events.PlaylistChangeEvent)
	if !ok {
		log.Printf("[ERROR] handlePlaylistChangeEvent: Failed to cast payload to PlaylistChangeEvent")
		return
	}

	wsEvent := PlaylistChangeEvent{
		Song:      playlistChangeEvent.Song,
		NextSong:  playlistChangeEvent.NextSong,
		Playlist:  playlistChangeEvent.Playlist,
		State:     playlistChangeEvent.State,
		Timestamp: playlistChangeEvent.Timestamp,
	}

	message := Message{
		Type:    "playlist_change",
		Payload: wsEvent,
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("[ERROR] handlePlaylistChangeEvent: Failed to marshal event: %v", err)
		return
	}

	h.broadcast <- data
}

func (h *Handler) Run() {
	// Increase broadcast frequency for better synchronization
	ticker := time.NewTicker(100 * time.Millisecond) // 10 FPS for smooth updates
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

			// Send immediate state to new client
			go client.sendPlaybackState()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}

	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to websocket: %v", err)
		return
	}

	client := &Client{
		conn:     conn,
		send:     make(chan []byte, 256),
		radioSvc: h.radioSvc,
		handler:  h,
	}

	h.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) handleMessage(messageType int, data []byte) {
	var request ClientRequest
	if err := json.Unmarshal(data, &request); err != nil {
		log.Printf("[ERROR] handleMessage: Failed to unmarshal request: %v", err)
		return
	}

	switch request.Type {
	case "get_playback_state":
		c.sendPlaybackState()
	case "ping":
		// Send pong response
		response := Message{
			Type:    "pong",
			Payload: map[string]interface{}{"timestamp": time.Now().UnixMilli()},
		}
		if responseData, err := json.Marshal(response); err == nil {
			select {
			case c.send <- responseData:
			default:
			}
		}
	case "reaction":
		// Handle reaction request
		var reactionReq ReactionRequest
		if err := json.Unmarshal(data, &reactionReq); err != nil {
			log.Printf("[ERROR] handleMessage: Failed to unmarshal reaction request: %v", err)
			return
		}

		// Publish reaction to event bus
		if c.handler.eventBus != nil {
			c.handler.eventBus.(interface {
				PublishUserReaction(userID, emote string)
			}).PublishUserReaction(reactionReq.UserID, reactionReq.Emote)
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.handler.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		messageType, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error reading message: %v", err)
			}
			break
		}

		// Handle client messages
		if messageType == websocket.TextMessage {
			c.handleMessage(messageType, data)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) sendPlaybackState() {
	state := c.radioSvc.GetPlaybackState()
	if state == nil || c.radioSvc.GetCurrentSong() == nil {
		// Send empty state to indicate no song is playing
		update := PlaybackUpdate{
			Song:      nil,
			Elapsed:   0,
			Remaining: 0,
			Paused:    true,
			TotalTime: 0,
			Timestamp: time.Now().UnixMilli(),
		}

		message := Message{
			Type:    "playback_state",
			Payload: update,
		}

		data, err := json.Marshal(message)
		if err != nil {
			log.Printf("[ERROR] sendPlaybackState: Failed to marshal empty state: %v", err)
			return
		}

		select {
		case c.send <- data:
		default:
		}
		return
	}

	elapsed := c.radioSvc.GetElapsedTime().Seconds()
	remaining := c.radioSvc.GetRemainingTime().Seconds()

	currentSong := c.radioSvc.GetCurrentSong()

	update := PlaybackUpdate{
		Song:      currentSong,
		Elapsed:   elapsed,
		Remaining: remaining,
		Paused:    state.Paused,
		TotalTime: float64(currentSong.Duration),
		Timestamp: time.Now().UnixMilli(),
	}

	message := Message{
		Type:    "playback_state",
		Payload: update,
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("[ERROR] sendPlaybackState: Failed to marshal state: %v", err)
		return
	}

	select {
	case c.send <- data:
	default:
	}
}
