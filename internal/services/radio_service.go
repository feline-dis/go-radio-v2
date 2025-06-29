package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/feline-dis/go-radio-v2/internal/models"
)

// Test configuration - set to 5 seconds for faster song transitions during testing
const TestSongDuration = 5 * time.Second

// Interfaces for dependency injection and testing
type SongRepositoryInterface interface {
	GetRandomSong() (*models.Song, error)
	GetLeastPlayedSong() (*models.Song, error)
	UpdatePlayStats(youtubeID string) error
}

type PlaylistRepositoryInterface interface {
	GetFirstPlaylist() (*models.Playlist, error)
	GetSongs(playlistID string) ([]*models.Song, error)
}

type S3ServiceInterface interface {
	GetPresignedURL(ctx context.Context, key string, expires time.Duration) (string, error)
	UploadFile(ctx context.Context, key string, body io.Reader) error
	DeleteFile(ctx context.Context, key string) error
}

type EventBusInterface interface {
	PublishSongChange(currentSong, nextSong *models.Song, queueInfo *models.QueueInfo)
	PublishQueueUpdate(queueInfo *models.QueueInfo)
	PublishPlaybackUpdate(song *models.Song, elapsed, remaining float64, paused bool)
}

type RadioService struct {
	songRepo     SongRepositoryInterface
	playlistRepo PlaylistRepositoryInterface
	s3Service    S3ServiceInterface
	eventBus     EventBusInterface
	state        *models.PlaybackState
	mu           sync.RWMutex
	randMu       sync.Mutex // For thread-safe random number generation
}

func NewRadioService(
	songRepo SongRepositoryInterface,
	playlistRepo PlaylistRepositoryInterface,
	s3Service S3ServiceInterface,
	eventBus EventBusInterface,
) *RadioService {
	// Initialize with a non-nil state
	state := &models.PlaybackState{
		Queue: make([]*models.Song, 0),
	}
	return &RadioService{
		songRepo:     songRepo,
		playlistRepo: playlistRepo,
		s3Service:    s3Service,
		eventBus:     eventBus,
		state:        state,
	}
}

func (s *RadioService) GetCurrentSong() *models.Song {
	s.mu.RLock()
	defer func() {
		s.mu.RUnlock()
	}()

	if s.state.CurrentSong == nil {
		return nil
	}

	return s.state.CurrentSong
}

func (s *RadioService) GetPlaybackState() *models.PlaybackState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.state
}

func (s *RadioService) Play() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state.Paused {
		s.state.Paused = false
		s.state.StartTime = time.Now().Add(-s.state.PauseTime.Sub(s.state.StartTime))
		return nil
	}

	// Get a random song from the least played songs
	song, err := s.songRepo.GetLeastPlayedSong()
	if err != nil {
		return err
	}

	if song == nil {
		return nil
	}

	s.state.CurrentSong = song
	s.state.StartTime = time.Now()
	s.state.Paused = false

	// Update play stats
	return s.songRepo.UpdatePlayStats(song.YouTubeID)
}

func (s *RadioService) Pause() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.state.Paused {
		s.state.Paused = true
		s.state.PauseTime = time.Now()
	}
}

func (s *RadioService) Skip() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get a random song
	song, err := s.songRepo.GetRandomSong()
	if err != nil {
		return err
	}

	if song == nil {
		return nil
	}

	s.state.CurrentSong = song
	s.state.StartTime = time.Now()
	s.state.Paused = false

	// Update play stats
	return s.songRepo.UpdatePlayStats(song.YouTubeID)
}

func (s *RadioService) Rewind(seconds int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state.CurrentSong == nil {
		return
	}

	s.state.StartTime = time.Now().Add(-time.Duration(seconds) * time.Second)
}

func (s *RadioService) GetElapsedTime() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.state.CurrentSong == nil || s.state.Paused {
		return 0
	}

	return time.Since(s.state.StartTime)
}

func (s *RadioService) GetRemainingTime() time.Duration {
	s.mu.RLock()
	defer func() {
		s.mu.RUnlock()
	}()

	if s.state.CurrentSong == nil || s.state.Paused {
		return 0
	}

	elapsed := time.Since(s.state.StartTime)
	remaining := time.Duration(s.state.CurrentSong.Duration)*time.Second - elapsed

	if remaining < 0 {
		return 0
	}
	return remaining
}

func (s *RadioService) GetQueueInfo() *models.QueueInfo {
	s.mu.RLock()
	defer func() {
		s.mu.RUnlock()
	}()

	if s.state == nil {
		return &models.QueueInfo{
			CurrentSong: nil,
			NextSong:    nil,
			Queue:       []*models.Song{},
			Playlist:    nil,
		}
	}

	// Calculate remaining time directly to avoid deadlock
	var remaining float64
	if s.state.CurrentSong != nil && !s.state.Paused {
		elapsed := time.Since(s.state.StartTime)
		remainingDuration := time.Duration(s.state.CurrentSong.Duration)*time.Second - elapsed
		if remainingDuration > 0 {
			remaining = remainingDuration.Seconds()
		}
	}

	return &models.QueueInfo{
		CurrentSong: s.state.CurrentSong,
		NextSong:    s.state.NextSong,
		Queue:       s.state.Queue,
		Playlist:    s.state.CurrentPlaylist,
		Remaining:   remaining,
		StartTime:   s.state.StartTime,
	}
}

func (s *RadioService) StartPlaybackLoop() error {
	// Get the first playlist without holding the lock
	playlist, err := s.playlistRepo.GetFirstPlaylist()
	if err != nil {
		log.Printf("[ERROR] StartPlaybackLoop: Failed to get first playlist: %v", err)
		return fmt.Errorf("failed to get first playlist: %w", err)
	}
	if playlist == nil {
		log.Printf("[ERROR] StartPlaybackLoop: No playlists found")
		return fmt.Errorf("no playlists found")
	}

	// Get songs from the playlist without holding the lock
	songs, err := s.playlistRepo.GetSongs(playlist.ID)
	if err != nil {
		log.Printf("[ERROR] StartPlaybackLoop: Failed to get playlist songs: %v", err)
		return fmt.Errorf("failed to get playlist songs: %w", err)
	}
	if len(songs) == 0 {
		log.Printf("[ERROR] StartPlaybackLoop: Playlist %s is empty", playlist.ID)
		return fmt.Errorf("playlist %s is empty", playlist.ID)
	}

	// Verify songs data
	for i, song := range songs {
		log.Printf("[DEBUG] StartPlaybackLoop: Song %d - ID: %s, Title: %s, Duration: %d",
			i, song.YouTubeID, song.Title, song.Duration)
	}

	// Create new state before acquiring lock
	newState := &models.PlaybackState{
		CurrentPlaylist:  playlist,
		CurrentSongIndex: 0,
		CurrentSong:      songs[0],
		StartTime:        time.Now(),
		Paused:           false,
		NextSong:         songs[1%len(songs)],
		Queue:            make([]*models.Song, 0, 5),
		ShuffledSongs:    s.shuffleSongs(songs),
		IsShuffled:       false, // Start with original order
	}

	// Build initial queue
	queueSize := 5
	if len(songs) < queueSize {
		queueSize = len(songs)
	}
	for i := 0; i < queueSize; i++ {
		newState.Queue = append(newState.Queue, songs[i%len(songs)])
	}

	log.Printf("[DEBUG] StartPlaybackLoop: Created new state with song: %s and queue size: %d",
		newState.CurrentSong.Title, len(newState.Queue))

	// Set state with proper synchronization
	s.mu.Lock()
	s.state = newState
	s.mu.Unlock()

	// Send initial song change notification
	s.notifySongChange(songs[0], songs[1%len(songs)])

	// Verify state after initialization
	s.mu.RLock()
	state := s.state
	s.mu.RUnlock()

	if state == nil {
		log.Printf("[ERROR] StartPlaybackLoop: State is nil after initialization")
		return fmt.Errorf("state is nil after initialization")
	}
	if state.CurrentSong == nil {
		log.Printf("[ERROR] StartPlaybackLoop: CurrentSong is nil after initialization")
		return fmt.Errorf("CurrentSong is nil after initialization")
	}
	if len(state.Queue) == 0 {
		log.Printf("[ERROR] StartPlaybackLoop: Queue is empty after initialization")
		return fmt.Errorf("Queue is empty after initialization")
	}

	log.Printf("[DEBUG] StartPlaybackLoop: State verification passed - CurrentSong: %s, Queue size: %d",
		state.CurrentSong.Title, len(state.Queue))

	// Start the playback loop in a goroutine
	log.Printf("[DEBUG] StartPlaybackLoop: Starting playback loop goroutine")
	loopStarted := make(chan struct{})

	// Make a copy of songs to avoid race conditions
	songsCopy := make([]*models.Song, len(songs))
	copy(songsCopy, songs)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[ERROR] playbackLoop: Panic recovered: %v", r)
			}
		}()
		log.Printf("[DEBUG] playbackLoop: Goroutine started")
		close(loopStarted)
		s.playbackLoop(songsCopy)
	}()

	// Wait for goroutine to start
	select {
	case <-loopStarted:
		log.Printf("[DEBUG] StartPlaybackLoop: Playback loop goroutine confirmed started")
	case <-time.After(time.Second):
		log.Printf("[ERROR] StartPlaybackLoop: Playback loop goroutine failed to start within 1 second")
		return fmt.Errorf("playback loop goroutine failed to start")
	}

	// Verify state one more time after a short delay
	time.Sleep(100 * time.Millisecond)
	s.mu.RLock()
	state = s.state
	s.mu.RUnlock()
	log.Printf("[DEBUG] StartPlaybackLoop: Final state check - CurrentSong: %v, Queue size: %d",
		state.CurrentSong, len(state.Queue))

	return nil
}

func (s *RadioService) playbackLoop(songs []*models.Song) {
	log.Printf("[DEBUG] playbackLoop: Starting with %d songs", len(songs))

	// Create a ticker for periodic state updates
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// Log initial state
	s.mu.RLock()
	if s.state == nil {
		log.Printf("[ERROR] playbackLoop: State is nil at start")
		return
	}
	if s.state.CurrentSong == nil {
		log.Printf("[ERROR] playbackLoop: CurrentSong is nil at start")
		return
	}
	s.mu.RUnlock()

	for range ticker.C {
		// Get remaining time without holding the lock
		remaining := s.GetRemainingTime()

		// Song has finished playing
		if remaining <= 0 {
			// Only lock during the state update
			s.mu.Lock()

			// Check if we've reached the end of the playlist
			if s.state.CurrentSongIndex >= len(songs)-1 {
				// Playlist completed, shuffle and restart
				s.state.ShuffledSongs = s.shuffleSongs(songs)
				s.state.IsShuffled = true
				s.state.CurrentSongIndex = 0
				s.state.CurrentSong = s.state.ShuffledSongs[0]
				s.state.NextSong = s.state.ShuffledSongs[1%len(s.state.ShuffledSongs)]

				// Update queue with shuffled songs
				s.state.Queue = make([]*models.Song, 0, 5)
				queueSize := 5
				if len(s.state.ShuffledSongs) < queueSize {
					queueSize = len(s.state.ShuffledSongs)
				}
				for i := 0; i < queueSize; i++ {
					s.state.Queue = append(s.state.Queue, s.state.ShuffledSongs[i%len(s.state.ShuffledSongs)])
				}
			} else {
				// Move to next song in current playlist
				s.state.CurrentSongIndex++
				if s.state.IsShuffled {
					s.state.CurrentSong = s.state.ShuffledSongs[s.state.CurrentSongIndex]
					s.state.NextSong = s.state.ShuffledSongs[(s.state.CurrentSongIndex+1)%len(s.state.ShuffledSongs)]
				} else {
					s.state.CurrentSong = songs[s.state.CurrentSongIndex]
					s.state.NextSong = songs[(s.state.CurrentSongIndex+1)%len(songs)]
				}
			}

			s.state.StartTime = time.Now()
			s.state.Paused = false

			// Update queue
			if !s.state.IsShuffled {
				s.state.Queue = make([]*models.Song, 0, 5)
				queueSize := 5
				if len(songs) < queueSize {
					queueSize = len(songs)
				}
				for i := 0; i < queueSize; i++ {
					s.state.Queue = append(s.state.Queue, songs[(s.state.CurrentSongIndex+1+i)%len(songs)])
				}
			}

			currentSongID := s.state.CurrentSong.YouTubeID
			nextSong := s.state.NextSong
			currentSong := s.state.CurrentSong

			s.mu.Unlock()

			// Notify clients about song change
			s.notifySongChange(currentSong, nextSong)

			// Update play stats without holding the lock
			if err := s.songRepo.UpdatePlayStats(currentSongID); err != nil {
				log.Printf("[ERROR] playbackLoop: Failed to update play stats for song %s: %v", currentSongID, err)
			}
		}
	}
}

func (s *RadioService) shuffleSongs(songs []*models.Song) []*models.Song {
	s.randMu.Lock()
	defer s.randMu.Unlock()

	shuffled := make([]*models.Song, len(songs))
	copy(shuffled, songs)

	// Use global rand package with mutex protection
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled
}

func (s *RadioService) notifySongChange(currentSong, nextSong *models.Song) {
	fmt.Println("Notifying song change:", currentSong.Title, nextSong.Title)
	if s.eventBus != nil {
		// Get queue info once and reuse it
		queueInfo := s.GetQueueInfo()
		s.eventBus.PublishSongChange(currentSong, nextSong, queueInfo)

		// Also publish queue update with the same info
		s.eventBus.PublishQueueUpdate(queueInfo)
	}
}
