package services

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/feline-dis/go-radio-v2/internal/models"
)

// Mock repositories for testing
type MockSongRepository struct {
	songs           map[string]*models.Song
	randomSong      *models.Song
	leastPlayedSong *models.Song
	updateStatsErr  error
}

func NewMockSongRepository() *MockSongRepository {
	return &MockSongRepository{
		songs: make(map[string]*models.Song),
	}
}

func (m *MockSongRepository) GetRandomSong() (*models.Song, error) {
	return m.randomSong, nil
}

func (m *MockSongRepository) GetLeastPlayedSong() (*models.Song, error) {
	return m.leastPlayedSong, nil
}

func (m *MockSongRepository) UpdatePlayStats(youtubeID string) error {
	return m.updateStatsErr
}

func (m *MockSongRepository) Create(song *models.Song) error {
	m.songs[song.YouTubeID] = song
	return nil
}

func (m *MockSongRepository) GetByYouTubeID(youtubeID string) (*models.Song, error) {
	song, exists := m.songs[youtubeID]
	if !exists {
		return nil, nil
	}
	return song, nil
}

type MockPlaylistRepository struct {
	playlists     map[string]*models.Playlist
	songs         map[string][]*models.Song
	firstPlaylist *models.Playlist
}

func NewMockPlaylistRepository() *MockPlaylistRepository {
	return &MockPlaylistRepository{
		playlists: make(map[string]*models.Playlist),
		songs:     make(map[string][]*models.Song),
	}
}

func (m *MockPlaylistRepository) GetFirstPlaylist() (*models.Playlist, error) {
	return m.firstPlaylist, nil
}

func (m *MockPlaylistRepository) GetSongs(playlistID string) ([]*models.Song, error) {
	songs, exists := m.songs[playlistID]
	if !exists {
		return []*models.Song{}, nil
	}
	return songs, nil
}

func (m *MockPlaylistRepository) Create(playlist *models.Playlist) error {
	m.playlists[playlist.ID] = playlist
	return nil
}

func (m *MockPlaylistRepository) GetByID(id string) (*models.Playlist, error) {
	playlist, exists := m.playlists[id]
	if !exists {
		return nil, nil
	}
	return playlist, nil
}

func (m *MockPlaylistRepository) GetAll() ([]*models.Playlist, error) {
	var playlists []*models.Playlist
	for _, playlist := range m.playlists {
		playlists = append(playlists, playlist)
	}
	return playlists, nil
}

func (m *MockPlaylistRepository) AddSong(playlistID string, youtubeID string, position int) error {
	return nil
}

func (m *MockPlaylistRepository) RemoveSong(playlistID string, youtubeID string) error {
	return nil
}

func (m *MockPlaylistRepository) UpdateSongPosition(playlistID string, youtubeID string, newPosition int) error {
	return nil
}

func (m *MockPlaylistRepository) GetByName(name string) (*models.Playlist, error) {
	return nil, nil
}

type MockS3Service struct{}

func (m *MockS3Service) GetPresignedURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	return "https://example.com/signed-url", nil
}

func (m *MockS3Service) UploadFile(ctx context.Context, key string, body io.Reader) error {
	return nil
}

func (m *MockS3Service) DeleteFile(ctx context.Context, key string) error {
	return nil
}

type MockEventBus struct{}

func (m *MockEventBus) PublishSongChange(currentSong, nextSong *models.Song, queueInfo *models.QueueInfo) {
	// Mock implementation - do nothing for tests
}

func (m *MockEventBus) PublishQueueUpdate(queueInfo *models.QueueInfo) {
	// Mock implementation - do nothing for tests
}

func (m *MockEventBus) PublishPlaybackUpdate(song *models.Song, elapsed, remaining float64, paused bool) {
	// Mock implementation - do nothing for tests
}

func (m *MockEventBus) PublishSkip(song *models.Song, nextSong *models.Song, state *models.PlaybackState) {
	// Mock implementation - do nothing for tests
}

func (m *MockEventBus) PublishPrevious(song *models.Song, nextSong *models.Song, state *models.PlaybackState) {
	// Mock implementation - do nothing for tests
}

func (m *MockEventBus) PublishPlaylistChange(song *models.Song, nextSong *models.Song, playlist *models.Playlist, state *models.PlaybackState) {
	// Mock implementation - do nothing for tests  
}

// Helper function to create test songs
func createTestSong(id, title, artist string, duration int) *models.Song {
	return &models.Song{
		YouTubeID: id,
		Title:     title,
		Artist:    artist,
		Duration:  duration,
		PlayCount: 0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Helper function to create test playlist
func createTestPlaylist(id string, name string) *models.Playlist {
	return &models.Playlist{
		ID:          id,
		Name:        name,
		Description: "Test playlist",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func TestNewRadioService(t *testing.T) {
	songRepo := NewMockSongRepository()
	playlistRepo := NewMockPlaylistRepository()
	s3Service := &MockS3Service{}
	eventBus := &MockEventBus{}

	service := NewRadioService(songRepo, playlistRepo, s3Service, eventBus)

	if service == nil {
		t.Fatal("Expected RadioService to be created, got nil")
	}

	if service.state == nil {
		t.Fatal("Expected state to be initialized, got nil")
	}

	if service.state.Queue == nil {
		t.Fatal("Expected queue to be initialized, got nil")
	}
}

func TestGetCurrentSong(t *testing.T) {
	songRepo := NewMockSongRepository()
	playlistRepo := NewMockPlaylistRepository()
	s3Service := &MockS3Service{}
	eventBus := &MockEventBus{}

	service := NewRadioService(songRepo, playlistRepo, s3Service, eventBus)

	// Test when no song is playing
	song := service.GetCurrentSong()
	if song != nil {
		t.Errorf("Expected nil when no song is playing, got %v", song)
	}

	// Test when a song is playing
	testSong := createTestSong("test123", "Test Song", "Test Artist", 180)
	service.state.CurrentSong = testSong

	song = service.GetCurrentSong()
	if song == nil {
		t.Fatal("Expected song to be returned, got nil")
	}

	if song.YouTubeID != testSong.YouTubeID {
		t.Errorf("Expected YouTubeID %s, got %s", testSong.YouTubeID, song.YouTubeID)
	}
}

func TestGetPlaybackState(t *testing.T) {
	songRepo := NewMockSongRepository()
	playlistRepo := NewMockPlaylistRepository()
	s3Service := &MockS3Service{}
	eventBus := &MockEventBus{}

	service := NewRadioService(songRepo, playlistRepo, s3Service, eventBus)

	state := service.GetPlaybackState()
	if state == nil {
		t.Fatal("Expected playback state to be returned, got nil")
	}

	if state.Queue == nil {
		t.Fatal("Expected queue to be initialized in state")
	}
}

func TestSkip(t *testing.T) {
	songRepo := NewMockSongRepository()
	playlistRepo := NewMockPlaylistRepository()
	s3Service := &MockS3Service{}
	eventBus := &MockEventBus{}

	service := NewRadioService(songRepo, playlistRepo, s3Service, eventBus)

	// Test skip with no playlist
	err := service.Skip()
	if err == nil {
		t.Error("Expected error when no playlist available")
	}

	// Test skip with playlist
	playlist := createTestPlaylist("1", "Test Playlist")
	songs := []*models.Song{
		createTestSong("song1", "Song 1", "Artist 1", 180),
		createTestSong("song2", "Song 2", "Artist 2", 200),
		createTestSong("song3", "Song 3", "Artist 3", 160),
	}
	
	service.state.CurrentPlaylist = playlist
	service.state.CurrentSongIndex = 0 // Start at first song
	service.state.CurrentSong = songs[0]
	
	playlistRepo.songs["1"] = songs

	err = service.Skip()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if service.state.CurrentSongIndex != 1 {
		t.Errorf("Expected current song index to be 1, got %d", service.state.CurrentSongIndex)
	}

	if service.state.CurrentSong.YouTubeID != "song2" {
		t.Errorf("Expected current song to be song2, got %s", service.state.CurrentSong.YouTubeID)
	}
}

func TestSetActivePlaylist(t *testing.T) {
	songRepo := NewMockSongRepository()
	playlistRepo := NewMockPlaylistRepository()
	s3Service := &MockS3Service{}
	eventBus := &MockEventBus{}

	service := NewRadioService(songRepo, playlistRepo, s3Service, eventBus)

	// Test with non-existent playlist
	err := service.SetActivePlaylist("non-existent")
	if err == nil {
		t.Error("Expected error when playlist not found")
	}

	// Set up playlists
	playlist1 := createTestPlaylist("1", "Test Playlist 1")
	playlist2 := createTestPlaylist("2", "Test Playlist 2")
	songs1 := []*models.Song{
		createTestSong("song1", "Song 1", "Artist 1", 180),
		createTestSong("song2", "Song 2", "Artist 2", 200),
	}
	songs2 := []*models.Song{
		createTestSong("song3", "Song 3", "Artist 3", 160),
		createTestSong("song4", "Song 4", "Artist 4", 220),
	}

	playlistRepo.playlists["1"] = playlist1
	playlistRepo.playlists["2"] = playlist2
	playlistRepo.songs["1"] = songs1
	playlistRepo.songs["2"] = songs2

	// Test setting active playlist
	err = service.SetActivePlaylist("2")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if service.state.CurrentPlaylist.ID != "2" {
		t.Errorf("Expected current playlist ID to be '2', got %s", service.state.CurrentPlaylist.ID)
	}

	if service.state.CurrentSong.YouTubeID != "song3" {
		t.Errorf("Expected current song to be 'song3', got %s", service.state.CurrentSong.YouTubeID)
	}

	if service.state.CurrentSongIndex != 0 {
		t.Errorf("Expected current song index to be 0, got %d", service.state.CurrentSongIndex)
	}

	// Test with empty playlist
	emptyPlaylist := createTestPlaylist("empty", "Empty Playlist")
	playlistRepo.playlists["empty"] = emptyPlaylist
	playlistRepo.songs["empty"] = []*models.Song{}

	err = service.SetActivePlaylist("empty")
	if err == nil {
		t.Error("Expected error when playlist is empty")
	}
}

func TestPrevious(t *testing.T) {
	songRepo := NewMockSongRepository()
	playlistRepo := NewMockPlaylistRepository()
	s3Service := &MockS3Service{}
	eventBus := &MockEventBus{}

	service := NewRadioService(songRepo, playlistRepo, s3Service, eventBus)

	// Test previous with no playlist
	err := service.Previous()
	if err == nil {
		t.Error("Expected error when no playlist available")
	}

	// Test previous with playlist
	playlist := createTestPlaylist("1", "Test Playlist")
	songs := []*models.Song{
		createTestSong("song1", "Song 1", "Artist 1", 180),
		createTestSong("song2", "Song 2", "Artist 2", 200),
		createTestSong("song3", "Song 3", "Artist 3", 160),
	}
	
	service.state.CurrentPlaylist = playlist
	service.state.CurrentSongIndex = 1 // Start at second song
	service.state.CurrentSong = songs[1]
	
	playlistRepo.songs["1"] = songs

	err = service.Previous()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if service.state.CurrentSongIndex != 0 {
		t.Errorf("Expected current song index to be 0, got %d", service.state.CurrentSongIndex)
	}
}

func TestGetElapsedTime(t *testing.T) {
	songRepo := NewMockSongRepository()
	playlistRepo := NewMockPlaylistRepository()
	s3Service := &MockS3Service{}
	eventBus := &MockEventBus{}

	service := NewRadioService(songRepo, playlistRepo, s3Service, eventBus)

	// Test with no current song
	elapsed := service.GetElapsedTime()
	if elapsed != 0 {
		t.Errorf("Expected 0 elapsed time with no song, got %v", elapsed)
	}

	// Test with playing song
	service.state.CurrentSong = createTestSong("test123", "Test Song", "Test Artist", 180)
	service.state.StartTime = time.Now().Add(-time.Second)
	elapsed = service.GetElapsedTime()
	if elapsed <= 0 {
		t.Errorf("Expected positive elapsed time, got %v", elapsed)
	}
}

func TestGetRemainingTime(t *testing.T) {
	songRepo := NewMockSongRepository()
	playlistRepo := NewMockPlaylistRepository()
	s3Service := &MockS3Service{}
	eventBus := &MockEventBus{}

	service := NewRadioService(songRepo, playlistRepo, s3Service, eventBus)

	// Test with no current song
	remaining := service.GetRemainingTime()
	if remaining != 0 {
		t.Errorf("Expected 0 remaining time with no song, got %v", remaining)
	}

	// Test with playing song
	service.state.CurrentSong = createTestSong("test123", "Test Song", "Test Artist", 180)
	service.state.StartTime = time.Now().Add(-time.Second)
	remaining = service.GetRemainingTime()
	if remaining <= 0 {
		t.Errorf("Expected positive remaining time, got %v", remaining)
	}

	// Test with song that has finished
	service.state.StartTime = time.Now().Add(-time.Duration(service.state.CurrentSong.Duration+1) * time.Second)
	remaining = service.GetRemainingTime()
	if remaining != 0 {
		t.Errorf("Expected 0 remaining time for finished song, got %v", remaining)
	}
}

func TestGetQueueInfo(t *testing.T) {
	songRepo := NewMockSongRepository()
	playlistRepo := NewMockPlaylistRepository()
	s3Service := &MockS3Service{}
	eventBus := &MockEventBus{}

	service := NewRadioService(songRepo, playlistRepo, s3Service, eventBus)

	// Test with empty state
	queueInfo := service.GetQueueInfo()
	if queueInfo == nil {
		t.Fatal("Expected queue info to be returned, got nil")
	}

	if queueInfo.CurrentSong != nil {
		t.Errorf("Expected nil current song, got %v", queueInfo.CurrentSong)
	}

	if len(queueInfo.Queue) != 0 {
		t.Errorf("Expected empty queue, got %d items", len(queueInfo.Queue))
	}

	// Test with populated state
	testSong := createTestSong("test123", "Test Song", "Test Artist", 180)
	testPlaylist := createTestPlaylist("1", "Test Playlist")
	testQueue := []*models.Song{testSong}

	service.state.CurrentSong = testSong
	service.state.NextSong = testSong
	service.state.CurrentPlaylist = testPlaylist
	service.state.Queue = testQueue

	queueInfo = service.GetQueueInfo()
	if queueInfo.CurrentSong == nil {
		t.Fatal("Expected current song to be returned, got nil")
	}

	if queueInfo.CurrentSong.YouTubeID != testSong.YouTubeID {
		t.Errorf("Expected current song ID %s, got %s", testSong.YouTubeID, queueInfo.CurrentSong.YouTubeID)
	}

	if len(queueInfo.Queue) != 1 {
		t.Errorf("Expected queue with 1 item, got %d", len(queueInfo.Queue))
	}

	if queueInfo.Playlist == nil {
		t.Fatal("Expected playlist to be returned, got nil")
	}
}

func TestStartPlaybackLoop(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*MockPlaylistRepository, *MockSongRepository)
		expectedError bool
	}{
		{
			name: "Start playback loop successfully",
			setupMocks: func(playlistRepo *MockPlaylistRepository, songRepo *MockSongRepository) {
				playlist := createTestPlaylist("1", "Test Playlist")
				songs := []*models.Song{
					createTestSong("song1", "Song 1", "Artist 1", 180),
					createTestSong("song2", "Song 2", "Artist 2", 200),
					createTestSong("song3", "Song 3", "Artist 3", 160),
				}
				playlistRepo.firstPlaylist = playlist
				playlistRepo.songs["1"] = songs
			},
			expectedError: false,
		},
		{
			name: "No playlists available",
			setupMocks: func(playlistRepo *MockPlaylistRepository, songRepo *MockSongRepository) {
				playlistRepo.firstPlaylist = nil
			},
			expectedError: true,
		},
		{
			name: "Empty playlist",
			setupMocks: func(playlistRepo *MockPlaylistRepository, songRepo *MockSongRepository) {
				playlist := createTestPlaylist("1", "Test Playlist")
				playlistRepo.firstPlaylist = playlist
				playlistRepo.songs["1"] = []*models.Song{}
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			songRepo := NewMockSongRepository()
			playlistRepo := NewMockPlaylistRepository()
			s3Service := &MockS3Service{}
			eventBus := &MockEventBus{}

			service := NewRadioService(songRepo, playlistRepo, s3Service, eventBus)

			tt.setupMocks(playlistRepo, songRepo)

			err := service.StartPlaybackLoop()

			if tt.expectedError && err == nil {
				t.Errorf("Expected error, got nil")
			}

			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if !tt.expectedError {
				// Give the goroutine a moment to start
				time.Sleep(100 * time.Millisecond)

				state := service.GetPlaybackState()
				if state.CurrentSong == nil {
					t.Error("Expected current song to be set after successful start")
				}

				if len(state.Queue) == 0 {
					t.Error("Expected queue to be populated after successful start")
				}
			}
		})
	}
}

func TestPlaybackLoopStateTransitions(t *testing.T) {
	songRepo := NewMockSongRepository()
	playlistRepo := NewMockPlaylistRepository()
	s3Service := &MockS3Service{}
	eventBus := &MockEventBus{}

	service := NewRadioService(songRepo, playlistRepo, s3Service, eventBus)

	// Set up a playlist with short songs for testing
	playlist := createTestPlaylist("1", "Test Playlist")
	songs := []*models.Song{
		createTestSong("song1", "Song 1", "Artist 1", 1), // 1 second duration
		createTestSong("song2", "Song 2", "Artist 2", 1),
		createTestSong("song3", "Song 3", "Artist 3", 1),
	}

	playlistRepo.firstPlaylist = playlist
	playlistRepo.songs["1"] = songs

	// Start playback loop
	err := service.StartPlaybackLoop()
	if err != nil {
		t.Fatalf("Failed to start playback loop: %v", err)
	}

	// Wait for initial state to be set
	time.Sleep(200 * time.Millisecond)

	initialSong := service.GetCurrentSong()
	if initialSong == nil {
		t.Fatal("Expected initial song to be set")
	}

	// Wait for song to finish and transition to next song
	time.Sleep(2 * time.Second)

	newSong := service.GetCurrentSong()
	if newSong == nil {
		t.Fatal("Expected new song to be set after transition")
	}

	if newSong.YouTubeID == initialSong.YouTubeID {
		t.Error("Expected song to change after duration elapsed")
	}
}

func TestConcurrentAccess(t *testing.T) {
	songRepo := NewMockSongRepository()
	playlistRepo := NewMockPlaylistRepository()
	s3Service := &MockS3Service{}
	eventBus := &MockEventBus{}

	service := NewRadioService(songRepo, playlistRepo, s3Service, eventBus)

	// Set up some state
	service.state.CurrentSong = createTestSong("test123", "Test Song", "Test Artist", 180)
	service.state.Queue = []*models.Song{service.state.CurrentSong}

	// Test concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				service.GetCurrentSong()
				service.GetQueueInfo()
				service.GetElapsedTime()
				service.GetRemainingTime()
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestUpdatePlayStatsError(t *testing.T) {
	songRepo := NewMockSongRepository()
	playlistRepo := NewMockPlaylistRepository()
	s3Service := &MockS3Service{}
	eventBus := &MockEventBus{}

	service := NewRadioService(songRepo, playlistRepo, s3Service, eventBus)

	// Set up error in song repository
	songRepo.updateStatsErr = errors.New("database error")

	// Set up playlist and songs
	playlist := createTestPlaylist("1", "Test Playlist")
	songs := []*models.Song{
		createTestSong("song1", "Song 1", "Artist 1", 1),
		createTestSong("song2", "Song 2", "Artist 2", 1),
	}

	playlistRepo.firstPlaylist = playlist
	playlistRepo.songs["1"] = songs

	// Start playback loop - should not fail due to stats update error
	err := service.StartPlaybackLoop()
	if err != nil {
		t.Fatalf("Expected playback loop to start despite stats update error: %v", err)
	}

	// Wait for song transition
	time.Sleep(2 * time.Second)

	// Verify playback continues despite stats update error
	currentSong := service.GetCurrentSong()
	if currentSong == nil {
		t.Fatal("Expected playback to continue despite stats update error")
	}
}
