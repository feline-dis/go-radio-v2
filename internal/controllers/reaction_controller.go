package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/feline-dis/go-radio-v2/internal/events"
)

type ReactionController struct {
	eventBus *events.EventBus
}

type ReactionRequest struct {
	UserID string `json:"user_id"`
	Emote  string `json:"emote"`
}

func NewReactionController(eventBus *events.EventBus) *ReactionController {
	return &ReactionController{
		eventBus: eventBus,
	}
}

// SendReaction handles POST requests to send a reaction
func (rc *ReactionController) SendReaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ReactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.UserID == "" || req.Emote == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Publish reaction to event bus
	rc.eventBus.PublishUserReaction(req.UserID, req.Emote)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Reaction sent successfully",
	})
}
