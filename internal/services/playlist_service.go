package services

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/feline-dis/go-radio-v2/internal/models"
	"github.com/feline-dis/go-radio-v2/internal/repositories"
)

type PlaylistService struct {
	playlistRepo *repositories.PlaylistRepository
	songRepo     *repositories.SongRepository
	youtubeSvc   *YouTubeService
}

// songProcessingResult holds the result of processing a song
type songProcessingResult struct {
	song     *models.Song
	position int
	err      error
}

// batchJob represents a batch of songs to be processed
type batchJob struct {
	songIDs    []string
	startIndex int
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

// CreatePlaylist creates a new playlist with the given songs using concurrent processing
func (s *PlaylistService) CreatePlaylist(name, description string, songIDs []string) (*models.Playlist, error) {
	// Create the playlist
	playlist := &models.Playlist{
		Name:        name,
		Description: description,
	}

	if err := s.playlistRepo.Create(playlist); err != nil {
		return nil, err
	}

	// Process songs concurrently if there are any
	if len(songIDs) > 0 {
		if err := s.processSongsConcurrently(playlist.ID, songIDs); err != nil {
			log.Printf("Error processing songs concurrently: %v", err)
			// Don't return error here as playlist was created successfully
		}
	}

	return playlist, nil
}

// processSongsConcurrently processes songs using concurrent workers
func (s *PlaylistService) processSongsConcurrently(playlistID string, songIDs []string) error {
	const (
		batchSize  = 10
		maxWorkers = 3 // Limit concurrent API calls to avoid rate limits
		rateLimit  = 100 * time.Millisecond
	)

	// Create batches
	batches := make([]batchJob, 0)
	for i := 0; i < len(songIDs); i += batchSize {
		end := i + batchSize
		if end > len(songIDs) {
			end = len(songIDs)
		}
		batches = append(batches, batchJob{
			songIDs:    songIDs[i:end],
			startIndex: i,
		})
	}

	// Create channels for job distribution and result collection
	jobChan := make(chan batchJob, len(batches))
	resultChan := make(chan []songProcessingResult, len(batches))

	// Rate limiter channel
	rateLimiter := make(chan struct{}, maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		rateLimiter <- struct{}{}
	}

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.processBatchWorker(jobChan, resultChan, rateLimiter, rateLimit)
		}()
	}

	// Send jobs to workers
	go func() {
		defer close(jobChan)
		for _, batch := range batches {
			jobChan <- batch
		}
	}()

	// Wait for workers to complete and close result channel
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results and add songs to playlist
	allResults := make([]songProcessingResult, 0)
	for batchResults := range resultChan {
		allResults = append(allResults, batchResults...)
	}

	// Sort results by position to maintain order
	sortedResults := make([]songProcessingResult, len(allResults))
	for _, result := range allResults {
		if result.err == nil && result.position < len(sortedResults) {
			sortedResults[result.position] = result
		}
	}

	// Add songs to playlist in order
	var addErrors []error
	for _, result := range sortedResults {
		if result.err != nil {
			addErrors = append(addErrors, result.err)
			continue
		}

		if result.song != nil {
			if err := s.playlistRepo.AddSong(playlistID, result.song.YouTubeID, result.position); err != nil {
				log.Printf("Error adding song to playlist: %v", err)
				addErrors = append(addErrors, err)
			}
		}
	}

	if len(addErrors) > 0 {
		log.Printf("Encountered %d errors while adding songs to playlist", len(addErrors))
	}

	return nil
}

// processBatchWorker processes batches of songs concurrently
func (s *PlaylistService) processBatchWorker(
	jobChan <-chan batchJob,
	resultChan chan<- []songProcessingResult,
	rateLimiter chan struct{},
	rateLimit time.Duration,
) {
	for job := range jobChan {
		// Wait for rate limiter
		<-rateLimiter

		// Process the batch
		results := s.processBatch(job.songIDs, job.startIndex)

		// Send results
		resultChan <- results

		// Return rate limiter token after delay
		go func() {
			time.Sleep(rateLimit)
			rateLimiter <- struct{}{}
		}()
	}
}

// processBatch processes a batch of songs and returns results
func (s *PlaylistService) processBatch(songIDs []string, startIndex int) []songProcessingResult {
	// Get song details from YouTube
	ids := strings.Join(songIDs, ",")
	detailsURL := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/videos?part=snippet,contentDetails&id=%s&key=%s",
		ids,
		s.youtubeSvc.apiKey,
	)

	resp, err := s.youtubeSvc.httpClient.Get(detailsURL)
	if err != nil {
		log.Printf("Error getting video details: %v", err)
		// Return errors for all songs in this batch
		results := make([]songProcessingResult, len(songIDs))
		for i := range songIDs {
			results[i] = songProcessingResult{
				position: startIndex + i,
				err:      err,
			}
		}
		return results
	}
	defer resp.Body.Close()

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
		log.Printf("Error decoding video details: %v", err)
		// Return errors for all songs in this batch
		results := make([]songProcessingResult, len(songIDs))
		for i := range songIDs {
			results[i] = songProcessingResult{
				position: startIndex + i,
				err:      err,
			}
		}
		return results
	}

	// Process each video item concurrently
	results := make([]songProcessingResult, len(videoResp.Items))
	var wg sync.WaitGroup

	for i, item := range videoResp.Items {
		wg.Add(1)
		go func(i int, item struct {
			ID      string `json:"id"`
			Snippet struct {
				Title       string `json:"title"`
				Description string `json:"description"`
			} `json:"snippet"`
			ContentDetails struct {
				Duration string `json:"duration"`
			} `json:"contentDetails"`
		}) {
			defer wg.Done()

			// Parse duration (format: PT1H2M10S)
			duration := parseDuration(item.ContentDetails.Duration)
			if duration == 0 {
				log.Printf("Warning: Could not parse duration for video %s", item.ID)
				results[i] = songProcessingResult{
					position: startIndex + i,
					err:      fmt.Errorf("could not parse duration for video %s", item.ID),
				}
				return
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
				results[i] = songProcessingResult{
					position: startIndex + i,
					err:      err,
				}
				return
			}

			if existingSong == nil {
				// Create new song
				if err := s.songRepo.Create(song); err != nil {
					log.Printf("Error creating song: %v", err)
					results[i] = songProcessingResult{
						position: startIndex + i,
						err:      err,
					}
					return
				}
			} else {
				// Use existing song
				song = existingSong
			}

			results[i] = songProcessingResult{
				song:     song,
				position: startIndex + i,
				err:      nil,
			}
		}(i, item)
	}

	wg.Wait()
	return results
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
func (s *PlaylistService) GetPlaylistByID(id string) (*models.Playlist, error) {
	return s.playlistRepo.GetByID(id)
}

// GetPlaylistSongs returns all songs in a playlist
func (s *PlaylistService) GetPlaylistSongs(playlistID string) ([]*models.Song, error) {
	return s.playlistRepo.GetSongs(playlistID)
}

// AddSongToPlaylist adds a song to a playlist at the specified position
func (s *PlaylistService) AddSongToPlaylist(playlistID string, songID string, position int) error {
	return s.playlistRepo.AddSong(playlistID, songID, position)
}

// RemoveSongFromPlaylist removes a song from a playlist
func (s *PlaylistService) RemoveSongFromPlaylist(playlistID string, songID string) error {
	return s.playlistRepo.RemoveSong(playlistID, songID)
}

// UpdateSongPosition updates the position of a song in a playlist
func (s *PlaylistService) UpdateSongPosition(playlistID string, songID string, newPosition int) error {
	return s.playlistRepo.UpdateSongPosition(playlistID, songID, newPosition)
}
