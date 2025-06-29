package controllers

import (
	"encoding/json"
	"log"
	"net/http"
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
	admin.HandleFunc("/skip", c.Skip).Methods("POST")
	admin.HandleFunc("/previous", c.Previous).Methods("POST")
	admin.HandleFunc("/playlist/set-active", c.SetActivePlaylist).Methods("POST")
}

func (c *RadioController) GetNowPlaying(w http.ResponseWriter, r *http.Request) {
	song := c.radioSvc.GetCurrentSong()
	if song == nil {
		http.Error(w, "No song is currently playing", http.StatusNotFound)
		return
	}

	response := struct {
		Song      *models.Song `json:"song"`
		Elapsed   float64      `json:"elapsed"`
		Remaining float64      `json:"remaining"`
	}{
		Song:      song,
		Elapsed:   c.radioSvc.GetElapsedTime().Seconds(),
		Remaining: c.radioSvc.GetRemainingTime().Seconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (c *RadioController) Skip(w http.ResponseWriter, r *http.Request) {
	c.radioSvc.Next()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"action": "skip",
	})
}

func (c *RadioController) Previous(w http.ResponseWriter, r *http.Request) {
	c.radioSvc.Previous()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"action": "previous",
	})
}

func (c *RadioController) GetQueue(w http.ResponseWriter, r *http.Request) {
	log.Printf("[DEBUG] GetQueue: Starting request handling")

	queueInfo := c.radioSvc.GetQueueInfo()
	log.Printf("[DEBUG] GetQueue: Got queue info: %+v", queueInfo)

	if queueInfo == nil {
		log.Printf("[DEBUG] GetQueue: Queue info is nil, creating empty queue")
		// Return empty queue info instead of error
		queueInfo = &models.QueueInfo{
			Queue:            []*models.Song{},
			Playlist:         nil,
			Remaining:        0,
			StartTime:        time.Time{},
			CurrentSongIndex: 0,
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
	elapsed := c.radioSvc.GetElapsedTime().Seconds()
	remaining := c.radioSvc.GetRemainingTime().Seconds()

	response := struct {
		CurrentSong *models.Song `json:"current_song"`
		Elapsed     float64      `json:"elapsed"`
		Remaining   float64      `json:"remaining"`
		Timestamp   int64        `json:"timestamp"`
	}{
		CurrentSong: c.radioSvc.GetCurrentSong(),
		Elapsed:     elapsed,
		Remaining:   remaining,
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

func (c *RadioController) SetActivePlaylist(w http.ResponseWriter, r *http.Request) {
	var request struct {
		PlaylistID string `json:"playlist_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.PlaylistID == "" {
		http.Error(w, "playlist_id is required", http.StatusBadRequest)
		return
	}

	if err := c.radioSvc.SetActivePlaylist(request.PlaylistID); err != nil {
		log.Printf("[ERROR] SetActivePlaylist: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":      "success",
		"action":      "playlist_changed",
		"playlist_id": request.PlaylistID,
	})
}
