package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/feline-dis/go-radio-v2/internal/models"
	"github.com/feline-dis/go-radio-v2/internal/services"
	"github.com/gorilla/mux"
)

type RadioController struct {
	radioSvc *services.RadioService
}

func NewRadioController(radioSvc *services.RadioService) *RadioController {
	return &RadioController{
		radioSvc: radioSvc,
	}
}

func (c *RadioController) RegisterRoutes(r *mux.Router) {
	// Public endpoints
	r.HandleFunc("/api/v1/health", c.HealthCheck).Methods("GET")
	r.HandleFunc("/api/v1/now-playing", c.GetNowPlaying).Methods("GET")
	r.HandleFunc("/api/v1/queue", c.GetQueue).Methods("GET")
	r.HandleFunc("/api/v1/debug/playback-state", c.GetDebugPlaybackState).Methods("GET")

	// Admin endpoints
	admin := r.PathPrefix("/api/v1/admin").Subrouter()
	admin.HandleFunc("/play", c.Play).Methods("POST")
	admin.HandleFunc("/pause", c.Pause).Methods("POST")
	admin.HandleFunc("/skip", c.Skip).Methods("POST")
	admin.HandleFunc("/rewind", c.Rewind).Methods("POST")
}

func (c *RadioController) GetNowPlaying(w http.ResponseWriter, r *http.Request) {
	song := c.radioSvc.GetCurrentSong()
	if song == nil {
		http.Error(w, "No song is currently playing", http.StatusNotFound)
		return
	}

	state := c.radioSvc.GetPlaybackState()
	response := struct {
		Song      *models.Song `json:"song"`
		Elapsed   float64      `json:"elapsed"`
		Remaining float64      `json:"remaining"`
		Paused    bool         `json:"paused"`
	}{
		Song:      song,
		Elapsed:   c.radioSvc.GetElapsedTime().Seconds(),
		Remaining: c.radioSvc.GetRemainingTime().Seconds(),
		Paused:    state.Paused,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (c *RadioController) Play(w http.ResponseWriter, r *http.Request) {
	if err := c.radioSvc.Play(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (c *RadioController) Pause(w http.ResponseWriter, r *http.Request) {
	c.radioSvc.Pause()
	w.WriteHeader(http.StatusOK)
}

func (c *RadioController) Skip(w http.ResponseWriter, r *http.Request) {
	if err := c.radioSvc.Skip(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (c *RadioController) Rewind(w http.ResponseWriter, r *http.Request) {
	seconds := 10 // Default rewind duration
	if s := r.URL.Query().Get("seconds"); s != "" {
		if sec, err := strconv.Atoi(s); err == nil && sec > 0 {
			seconds = sec
		}
	}

	c.radioSvc.Rewind(seconds)
	w.WriteHeader(http.StatusOK)
}

func (c *RadioController) GetQueue(w http.ResponseWriter, r *http.Request) {
	log.Printf("[DEBUG] GetQueue: Starting request handling")

	queueInfo := c.radioSvc.GetQueueInfo()
	log.Printf("[DEBUG] GetQueue: Got queue info: %+v", queueInfo)

	if queueInfo == nil {
		log.Printf("[DEBUG] GetQueue: Queue info is nil, creating empty queue")
		// Return empty queue info instead of error
		queueInfo = &models.QueueInfo{
			CurrentSong: nil,
			NextSong:    nil,
			Queue:       []*models.Song{},
			Playlist:    nil,
		}
	}

	log.Printf("[DEBUG] GetQueue: Setting content type header")
	w.Header().Set("Content-Type", "application/json")

	log.Printf("[DEBUG] GetQueue: Encoding response: %+v", queueInfo)
	if err := json.NewEncoder(w).Encode(queueInfo); err != nil {
		log.Printf("[ERROR] GetQueue: Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	log.Printf("[DEBUG] GetQueue: Response sent successfully")
}

func (c *RadioController) GetDebugPlaybackState(w http.ResponseWriter, r *http.Request) {
	state := c.radioSvc.GetPlaybackState()
	elapsed := c.radioSvc.GetElapsedTime().Seconds()
	remaining := c.radioSvc.GetRemainingTime().Seconds()

	response := struct {
		CurrentSong *models.Song `json:"current_song"`
		Elapsed     float64      `json:"elapsed"`
		Remaining   float64      `json:"remaining"`
		Paused      bool         `json:"paused"`
		Timestamp   int64        `json:"timestamp"`
	}{
		CurrentSong: state.CurrentSong,
		Elapsed:     elapsed,
		Remaining:   remaining,
		Paused:      state.Paused,
		Timestamp:   time.Now().UnixMilli(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (c *RadioController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := struct {
		Status    string `json:"status"`
		Timestamp int64  `json:"timestamp"`
	}{
		Status:    "healthy",
		Timestamp: time.Now().UnixMilli(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
