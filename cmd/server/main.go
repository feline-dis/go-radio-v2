package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"github.com/feline-dis/go-radio-v2/internal/config"
	"github.com/feline-dis/go-radio-v2/internal/controllers"
	"github.com/feline-dis/go-radio-v2/internal/events"
	"github.com/feline-dis/go-radio-v2/internal/middleware"
	"github.com/feline-dis/go-radio-v2/internal/services"
	"github.com/feline-dis/go-radio-v2/internal/storage"
	"github.com/feline-dis/go-radio-v2/internal/websocket"
)

func main() {
	cfg := config.Load()

	fmt.Println("Config:", cfg)

	// Initialize storage factory
	storageFactory := storage.NewStorageFactory(cfg)

	// Validate storage configuration
	if err := storageFactory.ValidateConfig(); err != nil {
		log.Fatalf("Invalid storage configuration: %v", err)
	}

	// Initialize repositories using storage factory
	songRepo, err := storageFactory.CreateSongRepository()
	if err != nil {
		log.Fatalf("Failed to initialize song repository: %v", err)
	}

	playlistRepo, err := storageFactory.CreatePlaylistRepository()
	if err != nil {
		log.Fatalf("Failed to initialize playlist repository: %v", err)
	}

	// Initialize file storage
	fileStorage, err := storageFactory.CreateFileStorage()
	if err != nil {
		log.Fatalf("Failed to initialize file storage: %v", err)
	}

	// Initialize YouTube service
	youtubeService, err := services.NewYouTubeService()
	if err != nil {
		log.Fatalf("Failed to initialize YouTube service: %v", err)
	}

	// Initialize yt-dlp service
	var ytdlpService services.YtDlpServiceInterface
	realService, err := services.NewYtDlpService()
	if err != nil {
		log.Printf("Warning: Failed to initialize yt-dlp service (yt-dlp not available): %v", err)
		log.Printf("Songs will not be automatically downloaded. Please install yt-dlp or add songs manually.")
		// Use mock service for testing/development without yt-dlp
		ytdlpService = services.NewMockYtDlpService(1*time.Second, false)
	} else {
		ytdlpService = realService
	}

	// Initialize event bus
	eventBus := events.NewEventBus()


	// Initialize services
	playlistService := services.NewPlaylistService(playlistRepo, songRepo, youtubeService)
	radioService := services.NewRadioService(songRepo, playlistRepo, fileStorage, eventBus, ytdlpService, cfg.Storage.LocalDataDir)

	// Initialize WebSocket handler with radio service and event bus
	wsHandler := websocket.NewHandler(radioService, eventBus)
	// Start WebSocket handler in a goroutine
	go wsHandler.Run()

	// Initialize JWT service
	jwtService := services.NewJWTService(cfg)

	// Initialize controllers
	radioController := controllers.NewRadioController(radioService)
	youtubeController := controllers.NewYouTubeController(youtubeService)
	playlistController := controllers.NewPlaylistController(playlistService, fileStorage)
	reactionController := controllers.NewReactionController(eventBus)
	authController := controllers.NewAuthController(jwtService, cfg)

	// Create router
	router := mux.NewRouter()

	// Add CORS middleware for cross-origin requests
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Allow requests from the React development server
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// WebSocket endpoint - register directly on the main router
	router.Handle("/ws", wsHandler)

	// Create a subrouter for all other routes that will use the logging middleware
	apiRouter := router.PathPrefix("").Subrouter()

	// Register all routes on the apiRouter instead of the main router
	radioController.RegisterRoutes(apiRouter)
	youtubeController.RegisterRoutes(apiRouter)
	playlistController.RegisterRoutes(apiRouter)
	authController.RegisterRoutes(apiRouter)
	
	// Register reaction routes
	apiRouter.HandleFunc("/api/v1/reactions", reactionController.SendReaction).Methods("POST")
	
	// Add server info endpoint
	apiRouter.HandleFunc("/api/v1/server-info", func(w http.ResponseWriter, r *http.Request) {
		info := map[string]interface{}{
			"server_port": cfg.Server.Port,
			"local_url": fmt.Sprintf("http://localhost:%s", cfg.Server.Port),
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
	}).Methods("GET")

	// Admin routes with JWT authentication middleware
	adminRouter := apiRouter.PathPrefix("/api/v1/admin").Subrouter()
	adminRouter.Use(middleware.AuthMiddleware(jwtService))

	// Serve static files for the frontend
	// Check for local development (client/dist) first, then Docker path
	staticDir := "./client/dist"
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		staticDir = "/app/static" // Docker production path
	}
	
	fs := http.FileServer(http.Dir(staticDir))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
	router.PathPrefix("/assets/").Handler(fs)
	router.PathPrefix("/favicon.ico").Handler(fs)
	router.PathPrefix("/manifest.json").Handler(fs)

	// Check if static directory exists and log its contents
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		log.Printf("Warning: Static directory %s does not exist", staticDir)
		log.Printf("Please build the frontend with: cd client && yarn build")
	} else {
		log.Printf("Static directory %s exists", staticDir)
		// List contents of static directory
		if entries, err := os.ReadDir(staticDir); err == nil {
			log.Printf("Static directory contents:")
			for _, entry := range entries {
				log.Printf("  - %s", entry.Name())
			}
		}
	}

	// Handle client-side routing - serve index.html for all non-API routes
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Serving request: %s", r.URL.Path)
		
		// Don't serve index.html for API routes or WebSocket
		if strings.HasPrefix(r.URL.Path, "/api") || strings.HasPrefix(r.URL.Path, "/ws") {
			http.NotFound(w, r)
			return
		}
		
		// For all other routes, serve index.html to support client-side routing
		indexPath := staticDir + "/index.html"
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			http.Error(w, "Frontend not built. Please run: cd client && yarn build", http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, indexPath)
	})

	// Create server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Create a channel to signal when the server is ready
	serverReady := make(chan struct{})

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", cfg.Server.Port)
		// Signal that the server is ready to accept connections
		close(serverReady)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Wait for server to be ready
	<-serverReady
	
	// Display server status prominently
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🎵 GO RADIO SERVER STARTED")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("🏠 Local URL:  http://localhost:%s\n", cfg.Server.Port)
	fmt.Println("")
	fmt.Println("🎧 Your radio is ready! Open the URL above in your browser.")
	fmt.Println("💡 Tip: Use a tunnel service like ngrok for external access")
	fmt.Println(strings.Repeat("=", 80) + "\n")
	
	log.Println("Server is ready to accept connections")

	// Start the playback loop
	if err := radioService.StartPlaybackLoop(); err != nil {
		log.Printf("Error starting playback loop: %v", err)
	} else {
		// Show URL reminder after successful startup
		fmt.Printf("\n🎵 Radio is now playing! Access URL:\n")
		fmt.Printf("   🏠 Local:   http://localhost:%s\n", cfg.Server.Port)
		fmt.Printf("   📡 Info:    http://localhost:%s/api/v1/server-info\n\n", cfg.Server.Port)
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()


	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
