package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"

	"github.com/feline-dis/go-radio-v2/internal/config"
	"github.com/feline-dis/go-radio-v2/internal/controllers"
	"github.com/feline-dis/go-radio-v2/internal/events"
	"github.com/feline-dis/go-radio-v2/internal/middleware"
	"github.com/feline-dis/go-radio-v2/internal/repositories"
	"github.com/feline-dis/go-radio-v2/internal/services"
	"github.com/feline-dis/go-radio-v2/internal/websocket"
)

func runMigrations() error {
	// Check if atlas is installed
	if _, err := exec.LookPath("atlas"); err != nil {
		log.Printf("Warning: Atlas not found in PATH. Skipping migrations.")
		return nil
	}

	// Run atlas migrate apply
	cmd := exec.Command("atlas", "migrate", "apply", "--env", "local")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	cfg := config.Load()

	fmt.Println("Config:", cfg)

	// Run database migrations
	if err := runMigrations(); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	// Open PostgreSQL database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.DBName, cfg.Database.SSLMode)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(time.Minute * 5)

	// Test the connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Initialize repositories
	songRepo := repositories.NewSongRepository(db)
	playlistRepo := repositories.NewPlaylistRepository(db)

	// Initialize S3 service
	s3Service, err := services.NewS3Service(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize S3 service: %v", err)
	}

	// Initialize YouTube service
	youtubeService, err := services.NewYouTubeService()
	if err != nil {
		log.Fatalf("Failed to initialize YouTube service: %v", err)
	}

	// Initialize event bus
	eventBus := events.NewEventBus()

	// Initialize services
	playlistService := services.NewPlaylistService(playlistRepo, songRepo, youtubeService)
	radioService := services.NewRadioService(songRepo, playlistRepo, s3Service, eventBus)

	// Initialize WebSocket handler with radio service and event bus
	wsHandler := websocket.NewHandler(radioService, eventBus)
	// Start WebSocket handler in a goroutine
	go wsHandler.Run()

	// Initialize JWT service
	jwtService := services.NewJWTService(cfg)

	// Initialize controllers
	radioController := controllers.NewRadioController(radioService)
	youtubeController := controllers.NewYouTubeController(youtubeService)
	playlistController := controllers.NewPlaylistController(playlistService, s3Service)
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

	// Admin routes with JWT authentication middleware
	adminRouter := apiRouter.PathPrefix("/api/v1/admin").Subrouter()
	adminRouter.Use(middleware.AuthMiddleware(jwtService))

	// Serve static files for the frontend
	fs := http.FileServer(http.Dir("/app/static"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
	router.PathPrefix("/assets/").Handler(fs)
	router.PathPrefix("/favicon.ico").Handler(fs)
	router.PathPrefix("/manifest.json").Handler(fs)

	// Check if static directory exists and log its contents
	if _, err := os.Stat("/app/static"); os.IsNotExist(err) {
		log.Printf("Warning: Static directory /app/static does not exist")
	} else {
		log.Printf("Static directory /app/static exists")
		// List contents of static directory
		if entries, err := os.ReadDir("/app/static"); err == nil {
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
		http.ServeFile(w, r, "/app/static/index.html")
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
	log.Println("Server is ready to accept connections")

	// Start the playback loop
	if err := radioService.StartPlaybackLoop(); err != nil {
		log.Printf("Error starting playback loop: %v", err)
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
