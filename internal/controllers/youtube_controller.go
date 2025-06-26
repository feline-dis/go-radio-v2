package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/feline-dis/go-radio-v2/internal/services"
	"github.com/gorilla/mux"
)

type YouTubeController struct {
	youtubeSvc *services.YouTubeService
}

func NewYouTubeController(youtubeSvc *services.YouTubeService) *YouTubeController {
	return &YouTubeController{
		youtubeSvc: youtubeSvc,
	}
}

func (c *YouTubeController) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/v1/youtube/search", c.SearchVideos).Methods("GET")
}

func (c *YouTubeController) SearchVideos(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Missing query parameter", http.StatusBadRequest)
		return
	}

	results, err := c.youtubeSvc.SearchVideos(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
