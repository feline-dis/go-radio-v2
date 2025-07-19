package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
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
	GetByID(playlistID string) (*models.Playlist, error)
}

type FileStorageInterface interface {
	GetPresignedURL(ctx context.Context, key string, expires time.Duration) (string, error)
	UploadFile(ctx context.Context, key string, body io.Reader) error
	DeleteFile(ctx context.Context, key string) error
	FileExists(ctx context.Context, key string) (bool, error)
}

type EventBusInterface interface {
	PublishSongChange(currentSong, nextSong *models.Song, queueInfo *models.QueueInfo)
	PublishQueueUpdate(queueInfo *models.QueueInfo)
	PublishPlaybackUpdate(song *models.Song, elapsed, remaining float64, paused bool)
	PublishSkip(song *models.Song, nextSong *models.Song, state *models.PlaybackState)
	PublishPrevious(song *models.Song, nextSong *models.Song, state *models.PlaybackState)
	PublishPlaylistChange(song *models.Song, nextSong *models.Song, playlist *models.Playlist, state *models.PlaybackState)
}

type RadioService struct {
	songRepo     SongRepositoryInterface
	playlistRepo PlaylistRepositoryInterface
	fileStorage  FileStorageInterface
	eventBus     EventBusInterface
	ytdlpService YtDlpServiceInterface
	state        *models.PlaybackState
	mu           sync.RWMutex
	randMu       sync.Mutex // For thread-safe random number generation
	dataDir      string     // Base directory for audio files
}

func NewRadioService(
	songRepo SongRepositoryInterface,
	playlistRepo PlaylistRepositoryInterface,
	fileStorage FileStorageInterface,
	eventBus EventBusInterface,
	ytdlpService YtDlpServiceInterface,
	dataDir string,
) *RadioService {
	// Initialize with a non-nil state
	state := &models.PlaybackState{
		Queue: make([]*models.Song, 0),
	}
	return &RadioService{
		songRepo:     songRepo,
		playlistRepo: playlistRepo,
		fileStorage:  fileStorage,
		eventBus:     eventBus,
		ytdlpService: ytdlpService,
		state:        state,
		dataDir:      dataDir,
	}
}

func (s *RadioService) GetPlaybackState() *models.PlaybackState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.state
}

func (s *RadioService) Next() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state == nil || len(s.state.Queue) == 0 {
		return
	}

	// Move to next song
	s.state.CurrentSongIndex = s.state.CurrentSongIndex + 1

	// Handle wrap-around at end of playlist
	if s.state.CurrentSongIndex >= len(s.state.Queue) {
		s.state.CurrentSongIndex = 0
	}

	s.state.StartTime = time.Now()

	// Get current and next songs safely
	var currentSong, nextSong *models.Song
	if s.state.CurrentSongIndex < len(s.state.Queue) {
		currentSong = s.state.Queue[s.state.CurrentSongIndex]
	}
	nextIndex := (s.state.CurrentSongIndex + 1) % len(s.state.Queue)
	if nextIndex < len(s.state.Queue) {
		nextSong = s.state.Queue[nextIndex]
	}

	// Create queue info without additional locking
	queueInfo := &models.QueueInfo{
		Queue:            s.state.Queue,
		Playlist:         s.state.CurrentPlaylist,
		Remaining:        0, // Will be calculated by client
		StartTime:        s.state.StartTime,
		CurrentSongIndex: s.state.CurrentSongIndex,
	}

	s.eventBus.PublishSongChange(currentSong, nextSong, queueInfo)
}

func (s *RadioService) Previous() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state == nil || len(s.state.Queue) == 0 {
		return
	}

	// Move to previous song
	s.state.CurrentSongIndex = s.state.CurrentSongIndex - 1

	// Handle wrap-around at beginning of playlist
	if s.state.CurrentSongIndex < 0 {
		s.state.CurrentSongIndex = len(s.state.Queue) - 1
	}

	s.state.StartTime = time.Now()

	// Get current and next songs safely
	var currentSong, nextSong *models.Song
	if s.state.CurrentSongIndex < len(s.state.Queue) {
		currentSong = s.state.Queue[s.state.CurrentSongIndex]
	}
	nextIndex := (s.state.CurrentSongIndex + 1) % len(s.state.Queue)
	if nextIndex < len(s.state.Queue) {
		nextSong = s.state.Queue[nextIndex]
	}

	// Create queue info without additional locking
	queueInfo := &models.QueueInfo{
		Queue:            s.state.Queue,
		Playlist:         s.state.CurrentPlaylist,
		Remaining:        0, // Will be calculated by client
		StartTime:        s.state.StartTime,
		CurrentSongIndex: s.state.CurrentSongIndex,
	}

	s.eventBus.PublishSongChange(currentSong, nextSong, queueInfo)
}

func (s *RadioService) GetElapsedTime() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.state == nil || len(s.state.Queue) == 0 {
		return 0
	}

	if s.state.CurrentSongIndex < 0 || s.state.CurrentSongIndex >= len(s.state.Queue) {
		return 0
	}

	return time.Since(s.state.StartTime)
}

func (s *RadioService) GetRemainingTime() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.state == nil || len(s.state.Queue) == 0 {
		return 0
	}

	if s.state.CurrentSongIndex < 0 || s.state.CurrentSongIndex >= len(s.state.Queue) {
		return 0
	}

	currentSong := s.state.Queue[s.state.CurrentSongIndex]
	if currentSong == nil {
		return 0
	}

	elapsed := time.Since(s.state.StartTime)
	remaining := time.Duration(currentSong.Duration)*time.Second - elapsed

	if remaining < 0 {
		return 0
	}
	return remaining
}

func (s *RadioService) GetQueueInfo() *models.QueueInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.state == nil {
		return &models.QueueInfo{
			Queue:            []*models.Song{},
			Playlist:         nil,
			Remaining:        0,
			StartTime:        time.Time{},
			CurrentSongIndex: 0,
		}
	}

	// Get current song safely without additional locking
	var currentSong *models.Song
	if len(s.state.Queue) > 0 && s.state.CurrentSongIndex >= 0 && s.state.CurrentSongIndex < len(s.state.Queue) {
		currentSong = s.state.Queue[s.state.CurrentSongIndex]
	}

	// Calculate remaining time directly to avoid deadlock
	var remaining float64
	if currentSong != nil && !s.state.Paused {
		elapsed := time.Since(s.state.StartTime)
		remainingDuration := time.Duration(currentSong.Duration)*time.Second - elapsed
		if remainingDuration > 0 {
			remaining = remainingDuration.Seconds()
		}
	}

	return &models.QueueInfo{
		Queue:            s.state.Queue,
		Playlist:         s.state.CurrentPlaylist,
		Remaining:        remaining,
		StartTime:        s.state.StartTime,
		CurrentSongIndex: s.state.CurrentSongIndex,
	}
}

func (s *RadioService) GetCurrentSong() *models.Song {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.state == nil || len(s.state.Queue) == 0 {
		return nil
	}

	if s.state.CurrentSongIndex < 0 || s.state.CurrentSongIndex >= len(s.state.Queue) {
		return nil
	}

	return s.state.Queue[s.state.CurrentSongIndex]
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

	shuffledSongs := s.shuffleSongs(songs)
	numShuffledSongs := len(shuffledSongs)

	// Create new state before acquiring lock
	newState := &models.PlaybackState{
		CurrentPlaylist:  playlist,
		CurrentSongIndex: 0,
		StartTime:        time.Now(),
		Paused:           false,
		Queue:            make([]*models.Song, 0, numShuffledSongs),
	}

	for i := 0; i < numShuffledSongs; i++ {
		newState.Queue = append(newState.Queue, shuffledSongs[i])
	}

	// Set state with proper synchronization
	s.mu.Lock()
	s.state = newState
	s.mu.Unlock()

	// Download the first song before starting playback
	log.Printf("[DEBUG] StartPlaybackLoop: Ensuring first song is downloaded")
	ctx := context.Background()
	if err := s.checkAndDownloadCurrentSong(ctx); err != nil {
		log.Printf("[ERROR] StartPlaybackLoop: Failed to download first song: %v", err)
		return fmt.Errorf("failed to download first song: %w", err)
	}

	// Send initial song change notification
	s.notifySongChange(songs[0], songs[1%len(songs)])

	// Verify state after initialization
	s.mu.RLock()
	state := s.state
	s.mu.RUnlock()

	currentSong := s.GetCurrentSong()

	if state == nil {
		log.Printf("[ERROR] StartPlaybackLoop: State is nil after initialization")
		return fmt.Errorf("state is nil after initialization")
	}
	if currentSong == nil {
		log.Printf("[ERROR] StartPlaybackLoop: CurrentSong is nil after initialization")
		return fmt.Errorf("currentSong is nil after initialization")
	}
	if len(state.Queue) == 0 {
		log.Printf("[ERROR] StartPlaybackLoop: Queue is empty after initialization")
		return fmt.Errorf("queue is empty after initialization")
	}

	log.Printf("[DEBUG] StartPlaybackLoop: State verification passed - CurrentSong: %s, Queue size: %d",
		currentSong.Title, len(state.Queue))

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
		s.GetCurrentSong(), len(state.Queue))

	return nil
}

func (s *RadioService) playbackLoop(songs []*models.Song) {
	log.Printf("[DEBUG] playbackLoop: Starting with %d songs", len(songs))

	// Create a ticker for periodic state updates
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// Log initial state
	for range ticker.C {
		// Get remaining time without holding the lock
		remaining := s.GetRemainingTime()

		// Song has finished playing
		if remaining <= 0 {
			// Only lock during the state update
			s.mu.Lock()

			if s.state == nil || len(s.state.Queue) == 0 {
				s.mu.Unlock()
				continue
			}

			// Check if we've reached the end of the playlist
			if s.state.CurrentSongIndex >= len(s.state.Queue)-1 {
				// Playlist completed, shuffle and restart
				shuffledSongs := s.shuffleSongs(s.state.Queue)
				s.state.CurrentSongIndex = 0
				s.state.StartTime = time.Now()

				// Update queue with shuffled songs
				s.state.Queue = make([]*models.Song, 0, len(shuffledSongs))
				for i := 0; i < len(shuffledSongs); i++ {
					s.state.Queue = append(s.state.Queue, shuffledSongs[i%len(shuffledSongs)])
				}

				// Get songs for notification without additional locking
				var currentSong, nextSong *models.Song
				if len(s.state.Queue) > 0 {
					currentSong = s.state.Queue[0]
					if len(s.state.Queue) > 1 {
						nextSong = s.state.Queue[1]
					}
				}

				// Create queue info without additional locking
				queueInfo := &models.QueueInfo{
					Queue:            s.state.Queue,
					Playlist:         s.state.CurrentPlaylist,
					Remaining:        0,
					StartTime:        s.state.StartTime,
					CurrentSongIndex: s.state.CurrentSongIndex,
				}

				s.mu.Unlock()

				// Ensure the new current song is downloaded
				if currentSong != nil {
					ctx := context.Background()
					if err := s.checkAndDownloadCurrentSong(ctx); err != nil {
						log.Printf("[ERROR] playbackLoop: Failed to download restarted song %s: %v", currentSong.YouTubeID, err)
					}
				}

				// Notify outside of lock
				if s.eventBus != nil && currentSong != nil {
					s.eventBus.PublishSongChange(currentSong, nextSong, queueInfo)
				}
			} else {
				// Move to next song - increment index
				s.state.CurrentSongIndex = s.state.CurrentSongIndex + 1
				s.state.StartTime = time.Now()

				// Get songs for notification without additional locking
				var currentSong, nextSong *models.Song
				if s.state.CurrentSongIndex < len(s.state.Queue) {
					currentSong = s.state.Queue[s.state.CurrentSongIndex]
				}
				nextIndex := (s.state.CurrentSongIndex + 1) % len(s.state.Queue)
				if nextIndex < len(s.state.Queue) {
					nextSong = s.state.Queue[nextIndex]
				}

				// Create queue info without additional locking
				queueInfo := &models.QueueInfo{
					Queue:            s.state.Queue,
					Playlist:         s.state.CurrentPlaylist,
					Remaining:        0,
					StartTime:        s.state.StartTime,
					CurrentSongIndex: s.state.CurrentSongIndex,
				}

				s.mu.Unlock()

				// Ensure the new current song is downloaded
				if currentSong != nil {
					ctx := context.Background()
					if err := s.checkAndDownloadCurrentSong(ctx); err != nil {
						log.Printf("[ERROR] playbackLoop: Failed to download next song %s: %v", currentSong.YouTubeID, err)
					}
				}

				// Notify outside of lock
				if s.eventBus != nil && currentSong != nil {
					s.eventBus.PublishSongChange(currentSong, nextSong, queueInfo)
				}
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
	if currentSong != nil {
		fmt.Println("Notifying song change:", currentSong.Title)
	}
	if s.eventBus != nil {
		// Get queue info once and reuse it
		queueInfo := s.GetQueueInfo()
		s.eventBus.PublishSongChange(currentSong, nextSong, queueInfo)

		// Also publish queue update with the same info
		s.eventBus.PublishQueueUpdate(queueInfo)
	}
}

// SetActivePlaylist changes the current playlist and restarts playback
func (s *RadioService) SetActivePlaylist(playlistID string) error {
	// Get the new playlist without holding the lock
	playlist, err := s.playlistRepo.GetByID(playlistID)
	if err != nil {
		return fmt.Errorf("failed to get playlist: %w", err)
	}
	if playlist == nil {
		return fmt.Errorf("playlist not found")
	}

	// Get songs from the new playlist
	songs, err := s.playlistRepo.GetSongs(playlist.ID)
	if err != nil {
		return fmt.Errorf("failed to get playlist songs: %w", err)
	}
	if len(songs) == 0 {
		return fmt.Errorf("playlist %s is empty", playlist.ID)
	}

	log.Printf("[DEBUG] SetActivePlaylist: Switching to playlist %s with %d songs", playlist.Name, len(songs))

	shuffledSongs := s.shuffleSongs(songs)

	// Create new state with the new playlist
	newState := &models.PlaybackState{
		CurrentPlaylist:  playlist,
		CurrentSongIndex: 0,
		StartTime:        time.Now(),
		Paused:           false,
		Queue:            make([]*models.Song, 0, len(shuffledSongs)),
	}

	// Build new queue
	for i := 0; i < len(shuffledSongs); i++ {
		newState.Queue = append(newState.Queue, shuffledSongs[i%len(shuffledSongs)])
	}

	// Set state with proper synchronization
	s.mu.Lock()
	s.state = newState

	// Get songs for notification without additional locking
	var currentSong, nextSong *models.Song
	if len(s.state.Queue) > 0 {
		currentSong = s.state.Queue[0]
		if len(s.state.Queue) > 1 {
			nextSong = s.state.Queue[1]
		}
	}

	// Create queue info without additional locking
	queueInfo := &models.QueueInfo{
		Queue:            s.state.Queue,
		Playlist:         s.state.CurrentPlaylist,
		Remaining:        0,
		StartTime:        s.state.StartTime,
		CurrentSongIndex: s.state.CurrentSongIndex,
	}

	s.mu.Unlock()

	// Ensure the new current song is downloaded
	if currentSong != nil {
		ctx := context.Background()
		if err := s.checkAndDownloadCurrentSong(ctx); err != nil {
			log.Printf("[ERROR] SetActivePlaylist: Failed to download first song %s: %v", currentSong.YouTubeID, err)
			return fmt.Errorf("failed to download first song: %w", err)
		}
	}

	// Broadcast playlist change event outside of lock
	if s.eventBus != nil && currentSong != nil {
		s.eventBus.PublishSongChange(currentSong, nextSong, queueInfo)
	}

	log.Printf("[DEBUG] SetActivePlaylist: Successfully switched to playlist %s", playlist.Name)
	return nil
}

// ensureSongDownloaded checks if a song is downloaded and downloads it if necessary
func (s *RadioService) ensureSongDownloaded(ctx context.Context, song *models.Song) error {
	if song == nil {
		return fmt.Errorf("song is nil")
	}

	// Get the full file path
	audioDir := filepath.Join(s.dataDir, "audio", "songs")
	fullPath := filepath.Join(audioDir, fmt.Sprintf("%s.mp3", song.YouTubeID))

	// Check if file already exists
	if _, err := os.Stat(fullPath); err == nil {
		log.Printf("[DEBUG] ensureSongDownloaded: Song %s already downloaded at %s", song.YouTubeID, fullPath)
		return nil
	}

	log.Printf("[INFO] ensureSongDownloaded: Downloading song %s (%s)", song.YouTubeID, song.Title)

	// Download the audio
	downloadedPath, err := s.ytdlpService.DownloadAudio(ctx, song.YouTubeID, audioDir)
	if err != nil {
		return fmt.Errorf("failed to download song %s: %w", song.YouTubeID, err)
	}

	log.Printf("[INFO] ensureSongDownloaded: Successfully downloaded %s to %s", song.YouTubeID, downloadedPath)
	return nil
}

// predownloadNextSong downloads the next song in the background
func (s *RadioService) predownloadNextSong(ctx context.Context) {
	s.mu.RLock()
	if s.state == nil || len(s.state.Queue) == 0 {
		s.mu.RUnlock()
		return
	}

	// Get next song index
	nextIndex := (s.state.CurrentSongIndex + 1) % len(s.state.Queue)
	if nextIndex >= len(s.state.Queue) {
		s.mu.RUnlock()
		return
	}

	nextSong := s.state.Queue[nextIndex]
	s.mu.RUnlock()

	if nextSong == nil {
		return
	}

	// Download in background
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		if err := s.ensureSongDownloaded(ctx, nextSong); err != nil {
			log.Printf("[WARN] predownloadNextSong: Failed to predownload next song %s: %v", nextSong.YouTubeID, err)
		} else {
			log.Printf("[DEBUG] predownloadNextSong: Successfully predownloaded next song %s", nextSong.YouTubeID)
		}
	}()
}

// checkAndDownloadCurrentSong ensures the current song is downloaded before playback
func (s *RadioService) checkAndDownloadCurrentSong(ctx context.Context) error {
	currentSong := s.GetCurrentSong()
	if currentSong == nil {
		return fmt.Errorf("no current song")
	}

	if err := s.ensureSongDownloaded(ctx, currentSong); err != nil {
		return fmt.Errorf("failed to ensure current song is downloaded: %w", err)
	}

	// Start predownloading the next song
	s.predownloadNextSong(ctx)

	return nil
}
