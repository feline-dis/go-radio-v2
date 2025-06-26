package services

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/feline-dis/go-radio-v2/internal/models"
	"github.com/feline-dis/go-radio-v2/internal/repositories"
)

type PlaylistService struct {
	playlistRepo *repositories.PlaylistRepository
	songRepo     *repositories.SongRepository
	youtubeSvc   *YouTubeService
}

func NewPlaylistService(
	playlistRepo *repositories.PlaylistRepository,
	songRepo *repositories.SongRepository,
	youtubeSvc *YouTubeService,
) *PlaylistService {
	return &PlaylistService{
		playlistRepo: playlistRepo,
		songRepo:     songRepo,
		youtubeSvc:   youtubeSvc,
	}
}

// CreatePlaylist creates a new playlist with the given songs
func (s *PlaylistService) CreatePlaylist(name, description string, songIDs []string) (*models.Playlist, error) {
	// Create the playlist
	playlist := &models.Playlist{
		Name:        name,
		Description: description,
	}

	if err := s.playlistRepo.Create(playlist); err != nil {
		return nil, err
	}

	// Process songs in batches to avoid hitting YouTube API limits
	batchSize := 10
	for i := 0; i < len(songIDs); i += batchSize {
		end := i + batchSize
		if end > len(songIDs) {
			end = len(songIDs)
		}
		batch := songIDs[i:end]

		// Get song details from YouTube
		ids := strings.Join(batch, ",")
		detailsURL := fmt.Sprintf(
			"https://www.googleapis.com/youtube/v3/videos?part=snippet,contentDetails&id=%s&key=%s",
			ids,
			s.youtubeSvc.apiKey,
		)

		resp, err := s.youtubeSvc.httpClient.Get(detailsURL)
		if err != nil {
			log.Printf("Error getting video details: %v", err)
			continue
		}

		var videoResp struct {
			Items []struct {
				ID      string `json:"id"`
				Snippet struct {
					Title       string `json:"title"`
					Description string `json:"description"`
				} `json:"snippet"`
				ContentDetails struct {
					Duration string `json:"duration"`
				} `json:"contentDetails"`
			} `json:"items"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&videoResp); err != nil {
			resp.Body.Close()
			log.Printf("Error decoding video details: %v", err)
			continue
		}
		resp.Body.Close()

		// Create songs and add them to playlist
		for j, item := range videoResp.Items {
			// Parse duration (format: PT1H2M10S)
			duration := parseDuration(item.ContentDetails.Duration)
			if duration == 0 {
				log.Printf("Warning: Could not parse duration for video %s", item.ID)
				continue
			}

			// Create song entry
			song := &models.Song{
				YouTubeID: item.ID,
				Title:     item.Snippet.Title,
				Artist:    "Unknown", // We could try to extract this from title/description
				Album:     "Unknown",
				Duration:  int(duration.Seconds()),
				S3Key:     fmt.Sprintf("songs/%s.mp3", item.ID), // Assuming this is the format
			}

			// Check if song already exists
			existingSong, err := s.songRepo.GetByYouTubeID(song.YouTubeID)
			if err != nil {
				log.Printf("Error checking existing song: %v", err)
				continue
			}

			if existingSong == nil {
				// Create new song
				if err := s.songRepo.Create(song); err != nil {
					log.Printf("Error creating song: %v", err)
					continue
				}
			}

			// Add song to playlist
			position := i + j
			if err := s.playlistRepo.AddSong(playlist.ID, song.YouTubeID, position); err != nil {
				log.Printf("Error adding song to playlist: %v", err)
				continue
			}
		}

		// Sleep briefly to avoid hitting rate limits
		time.Sleep(100 * time.Millisecond)
	}

	return playlist, nil
}

// parseDuration parses a YouTube duration string (e.g., "PT1H2M10S") into a time.Duration
func parseDuration(duration string) time.Duration {
	var hours, minutes, seconds int
	var err error

	// Remove PT prefix
	duration = strings.TrimPrefix(duration, "PT")

	// Parse hours
	if strings.Contains(duration, "H") {
		parts := strings.Split(duration, "H")
		hours, err = strconv.Atoi(parts[0])
		if err != nil {
			return 0
		}
		duration = parts[1]
	}

	// Parse minutes
	if strings.Contains(duration, "M") {
		parts := strings.Split(duration, "M")
		minutes, err = strconv.Atoi(parts[0])
		if err != nil {
			return 0
		}
		duration = parts[1]
	}

	// Parse seconds
	if strings.Contains(duration, "S") {
		parts := strings.Split(duration, "S")
		seconds, err = strconv.Atoi(parts[0])
		if err != nil {
			return 0
		}
	}

	return time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second
}

// GetAllPlaylists returns all playlists
func (s *PlaylistService) GetAllPlaylists() ([]*models.Playlist, error) {
	return s.playlistRepo.GetAll()
}

// GetPlaylistByID returns a playlist by its ID
func (s *PlaylistService) GetPlaylistByID(id int) (*models.Playlist, error) {
	return s.playlistRepo.GetByID(id)
}

// GetPlaylistSongs returns all songs in a playlist
func (s *PlaylistService) GetPlaylistSongs(playlistID int) ([]*models.Song, error) {
	return s.playlistRepo.GetSongs(playlistID)
}

// AddSongToPlaylist adds a song to a playlist at the specified position
func (s *PlaylistService) AddSongToPlaylist(playlistID int, songID string, position int) error {
	return s.playlistRepo.AddSong(playlistID, songID, position)
}

// RemoveSongFromPlaylist removes a song from a playlist
func (s *PlaylistService) RemoveSongFromPlaylist(playlistID int, songID string) error {
	return s.playlistRepo.RemoveSong(playlistID, songID)
}

// UpdateSongPosition updates the position of a song in a playlist
func (s *PlaylistService) UpdateSongPosition(playlistID int, songID string, newPosition int) error {
	return s.playlistRepo.UpdateSongPosition(playlistID, songID, newPosition)
}
