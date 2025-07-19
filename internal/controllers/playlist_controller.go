package controllers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/feline-dis/go-radio-v2/internal/services"
	"github.com/feline-dis/go-radio-v2/internal/storage"
	"github.com/gorilla/mux"
)

type PlaylistController struct {
	playlistSvc *services.PlaylistService
	fileStorage storage.FileStorage
}

func NewPlaylistController(playlistSvc *services.PlaylistService, fileStorage storage.FileStorage) *PlaylistController {
	return &PlaylistController{
		playlistSvc: playlistSvc,
		fileStorage: fileStorage,
	}
}

func (c *PlaylistController) RegisterRoutes(r *mux.Router) {
	// Public endpoints
	r.HandleFunc("/api/v1/playlists", c.GetPlaylists).Methods("GET")
	r.HandleFunc("/api/v1/playlists", c.CreatePlaylist).Methods("POST")
	r.HandleFunc("/api/v1/playlists/{id}", c.GetPlaylist).Methods("GET")
	r.HandleFunc("/api/v1/playlists/{id}/songs", c.GetPlaylistSongs).Methods("GET")
	r.HandleFunc("/api/v1/songs/{youtube_id}/file", c.GetSongFile).Methods("GET")
	r.HandleFunc("/api/v1/playlists/{youtube_id}/file", c.GetSongFile).Methods("GET") // Legacy endpoint for frontend compatibility

	// Admin endpoints
	admin := r.PathPrefix("/api/v1/admin/playlists").Subrouter()
	admin.HandleFunc("/{id}/songs", c.AddSongToPlaylist).Methods("POST")
	admin.HandleFunc("/{id}/songs/{songId}", c.RemoveSongFromPlaylist).Methods("DELETE")
	admin.HandleFunc("/{id}/songs/{songId}/position", c.UpdateSongPosition).Methods("PUT")
}

func (c *PlaylistController) GetPlaylists(w http.ResponseWriter, r *http.Request) {
	playlists, err := c.playlistSvc.GetAllPlaylists()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(playlists)
}

func (c *PlaylistController) GetPlaylist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Missing playlist ID", http.StatusBadRequest)
		return
	}

	playlist, err := c.playlistSvc.GetPlaylistByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if playlist == nil {
		http.Error(w, "Playlist not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(playlist)
}

func (c *PlaylistController) CreatePlaylist(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Songs       []string `json:"songs"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	playlist, err := c.playlistSvc.CreatePlaylist(request.Name, request.Description, request.Songs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(playlist)
}

func (c *PlaylistController) GetPlaylistSongs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Missing playlist ID", http.StatusBadRequest)
		return
	}

	songs, err := c.playlistSvc.GetPlaylistSongs(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(songs)
}

func (c *PlaylistController) AddSongToPlaylist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Missing playlist ID", http.StatusBadRequest)
		return
	}

	var request struct {
		SongID   string `json:"song_id"`
		Position int    `json:"position"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.playlistSvc.AddSongToPlaylist(id, request.SongID, request.Position); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (c *PlaylistController) RemoveSongFromPlaylist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Missing playlist ID", http.StatusBadRequest)
		return
	}

	songID := vars["songId"]
	if songID == "" {
		http.Error(w, "Missing song ID", http.StatusBadRequest)
		return
	}

	if err := c.playlistSvc.RemoveSongFromPlaylist(id, songID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (c *PlaylistController) UpdateSongPosition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Missing playlist ID", http.StatusBadRequest)
		return
	}

	songID := vars["songId"]
	if songID == "" {
		http.Error(w, "Missing song ID", http.StatusBadRequest)
		return
	}

	var request struct {
		Position int `json:"position"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.playlistSvc.UpdateSongPosition(id, songID, request.Position); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (c *PlaylistController) GetSongFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	youtubeID := vars["youtube_id"]
	if youtubeID == "" {
		http.Error(w, "Missing YouTube ID", http.StatusBadRequest)
		return
	}

	exists, err := c.fileStorage.FileExists(r.Context(), "songs/"+youtubeID+".mp3")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Set proper headers for audio streaming
	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Cache-Control", "public, max-age=31536000")

	file, err := c.fileStorage.GetFile(r.Context(), "songs/"+youtubeID+".mp3")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
