package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/feline-dis/go-radio-v2/internal/config"
	"github.com/feline-dis/go-radio-v2/internal/repositories"
	"github.com/feline-dis/go-radio-v2/internal/services"
	_ "modernc.org/sqlite"
)

func main() {
	// Parse command line arguments
	playlistName := flag.String("playlist", "", "Name of the playlist to download")
	flag.Parse()

	if *playlistName == "" {
		log.Fatal("Please provide a playlist name using -playlist flag")
	}

	// Load configuration
	cfg := config.Load()

	// Open database connection
	db, err := sql.Open("sqlite", cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Initialize repositories and services
	playlistRepo := repositories.NewPlaylistRepository(db)
	s3Service, err := services.NewS3Service(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize S3 service: %v", err)
	}

	// Get playlist by name
	playlist, err := playlistRepo.GetByName(*playlistName)
	if err != nil {
		log.Fatalf("Failed to get playlist: %v", err)
	}
	if playlist == nil {
		log.Fatalf("Playlist '%s' not found", *playlistName)
	}

	// Get all songs in the playlist
	songs, err := playlistRepo.GetSongs(playlist.ID)
	if err != nil {
		log.Fatalf("Failed to get playlist songs: %v", err)
	}

	log.Printf("Found %d songs in playlist '%s'", len(songs), playlist.Name)

	// Create temporary directory for downloads
	tempDir, err := os.MkdirTemp("", "go-radio-downloads-*")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Process each song
	for i, song := range songs {
		log.Printf("[%d/%d] Processing %s - %s", i+1, len(songs), song.Artist, song.Title)

		// Skip if song already exists in S3
		exists, err := s3Service.FileExists(context.Background(), song.S3Key)
		if err != nil {
			log.Printf("Error checking if song exists in S3: %v", err)
			continue
		}
		if exists {
			log.Printf("Song already exists in S3, skipping")
			continue
		}

		// Download song using yt-dlp
		outputPath := filepath.Join(tempDir, fmt.Sprintf("%s.mp3", song.YouTubeID))
		fmt.Println("Running command: ", "yt-dlp",
			"-x", // Extract audio
			"--audio-format", "mp3",
			"--audio-quality", "0", // Best quality
			"-o", outputPath,
			"https://www.youtube.com/watch?v="+song.YouTubeID,
		)
		downloadCmd := exec.Command("yt-dlp",
			"-x", // Extract audio
			"--audio-format", "mp3",
			"--audio-quality", "0", // Best quality
			"-o", outputPath,
			"https://www.youtube.com/watch?v="+song.YouTubeID,
		)

		if err := downloadCmd.Run(); err != nil {
			log.Printf("Failed to download song: %v", err)
			continue
		}

		// Check if the file was created with the exact name we specified
		downloadedFile := outputPath
		if _, err := os.Stat(downloadedFile); os.IsNotExist(err) {
			// If not found, try to find it with a different extension
			matches, err := filepath.Glob(filepath.Join(tempDir, song.YouTubeID+".*"))
			if err != nil || len(matches) == 0 {
				log.Printf("Failed to find downloaded file")
				continue
			}
			downloadedFile = matches[0]
		}

		// Normalize audio using ffmpeg
		normalizedFile := filepath.Join(tempDir, song.YouTubeID+"_normalized.mp3")
		normalizeCmd := exec.Command("ffmpeg",
			"-i", downloadedFile,
			"-af", "loudnorm=I=-16:TP=-1.5:LRA=11", // Normalize to -16 LUFS
			"-ar", "44100", // Set sample rate to 44.1kHz
			"-y", // Overwrite output file if it exists
			normalizedFile,
		)

		if err := normalizeCmd.Run(); err != nil {
			log.Printf("Failed to normalize audio: %v", err)
			continue
		}

		// Upload to S3
		file, err := os.Open(normalizedFile)
		if err != nil {
			log.Printf("Failed to open normalized file: %v", err)
			continue
		}

		if err := s3Service.UploadFile(context.Background(), song.S3Key, file); err != nil {
			file.Close()
			log.Printf("Failed to upload to S3: %v", err)
			continue
		}
		file.Close()

		// Clean up downloaded files
		os.Remove(downloadedFile)
		os.Remove(normalizedFile)

		log.Printf("Successfully processed song")
	}

	log.Printf("Finished processing playlist '%s'", playlist.Name)
}
