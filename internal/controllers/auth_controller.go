package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/feline-dis/go-radio-v2/internal/config"
	"github.com/feline-dis/go-radio-v2/internal/middleware"
	"github.com/feline-dis/go-radio-v2/internal/services"
	"github.com/gorilla/mux"
)

type AuthController struct {
	jwtService *services.JWTService
	config     *config.Config
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token   string `json:"token"`
	Message string `json:"message"`
}

type RefreshRequest struct {
	Token string `json:"token"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewAuthController(jwtService *services.JWTService, cfg *config.Config) *AuthController {
	return &AuthController{
		jwtService: jwtService,
		config:     cfg,
	}
}

func (ac *AuthController) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/v1/auth/login", ac.Login).Methods("POST")
	r.HandleFunc("/api/v1/auth/refresh", ac.RefreshToken).Methods("POST")
	
	// Protected routes
	authRouter := r.PathPrefix("/api/v1/auth").Subrouter()
	authRouter.Use(middleware.AuthMiddleware(ac.jwtService))
	authRouter.HandleFunc("/me", ac.GetCurrentUser).Methods("GET")
}

// Login handles user authentication
func (ac *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Validate credentials against config
	if req.Username != ac.config.Admin.Username || req.Password != ac.config.Admin.Password {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := ac.jwtService.GenerateToken(req.Username)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to generate token"})
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{
		Token:   token,
		Message: "Login successful",
	})
}

// RefreshToken handles token refresh
func (ac *AuthController) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Refresh the token
	newToken, err := ac.jwtService.RefreshToken(req.Token)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid or expired token"})
		return
	}

	// Return new token
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{
		Token:   newToken,
		Message: "Token refreshed successfully",
	})
}

// GetCurrentUser returns information about the currently authenticated user
func (ac *AuthController) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	username, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "User not authenticated"})
		return
	}

	response := struct {
		Username string `json:"username"`
		Message  string `json:"message"`
	}{
		Username: username,
		Message:  "User authenticated",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
} 