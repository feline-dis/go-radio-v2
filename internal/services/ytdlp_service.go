package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/feline-dis/go-radio-v2/internal/models"
)

// YtDlpServiceInterface defines the interface for YouTube download operations
type YtDlpServiceInterface interface {
	// DownloadAudio downloads audio from YouTube and returns the file path
	DownloadAudio(ctx context.Context, youtubeID string, outputDir string) (string, error)
	// GetVideoInfo gets metadata about a YouTube video without downloading
	GetVideoInfo(ctx context.Context, youtubeID string) (*models.Song, error)
	// IsVideoAvailable checks if a YouTube video is available for download
	IsVideoAvailable(ctx context.Context, youtubeID string) (bool, error)
}

// YtDlpService implements YouTube download functionality using yt-dlp
type YtDlpService struct {
	ytDlpPath string
	timeout   time.Duration
}

// NewYtDlpService creates a new YtDlpService instance
func NewYtDlpService() (*YtDlpService, error) {
	// Check if yt-dlp is available in PATH
	ytDlpPath, err := exec.LookPath("yt-dlp")
	if err != nil {
		return nil, fmt.Errorf("yt-dlp not found in PATH: %w", err)
	}

	return &YtDlpService{
		ytDlpPath: ytDlpPath,
		timeout:   5 * time.Minute, // 5 minute timeout for downloads
	}, nil
}

// DownloadAudio downloads audio from YouTube and returns the file path
func (s *YtDlpService) DownloadAudio(ctx context.Context, youtubeID string, outputDir string) (string, error) {
	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Define output template - just use the YouTube ID as filename
	outputTemplate := filepath.Join(outputDir, fmt.Sprintf("%s.%%(ext)s", youtubeID))
	url := fmt.Sprintf("https://www.youtube.com/watch?v=%s", youtubeID)

	// Create context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Build yt-dlp command
	cmd := exec.CommandContext(ctxWithTimeout, s.ytDlpPath,
		"--extract-audio",           // Extract audio only
		"--audio-format", "mp3",     // Convert to MP3
		"--audio-quality", "0",      // Best quality
		"--no-playlist",             // Don't download playlists
		"--output", outputTemplate,  // Output template
		"--no-warnings",             // Suppress warnings
		"--quiet",                   // Suppress most output
		url,
	)

	log.Printf("[DEBUG] YtDlpService: Running command: %s", strings.Join(cmd.Args, " "))

	// Run the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("yt-dlp failed: %w, output: %s", err, string(output))
	}

	// Determine the actual output file path
	expectedPath := filepath.Join(outputDir, fmt.Sprintf("%s.mp3", youtubeID))
	
	// Check if file exists
	if _, err := os.Stat(expectedPath); err != nil {
		return "", fmt.Errorf("downloaded file not found at expected path %s: %w", expectedPath, err)
	}

	log.Printf("[DEBUG] YtDlpService: Successfully downloaded %s to %s", youtubeID, expectedPath)
	return expectedPath, nil
}

// GetVideoInfo gets metadata about a YouTube video without downloading
func (s *YtDlpService) GetVideoInfo(ctx context.Context, youtubeID string) (*models.Song, error) {
	url := fmt.Sprintf("https://www.youtube.com/watch?v=%s", youtubeID)

	// Create context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Build yt-dlp command to get info
	cmd := exec.CommandContext(ctxWithTimeout, s.ytDlpPath,
		"--print", "%(title)s",     // Print title
		"--print", "%(uploader)s",  // Print uploader (artist)
		"--print", "%(duration)s",  // Print duration in seconds
		"--no-warnings",            // Suppress warnings
		"--quiet",                  // Suppress most output
		"--no-playlist",            // Don't process playlists
		url,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get video info: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 3 {
		return nil, fmt.Errorf("unexpected output format from yt-dlp")
	}

	title := strings.TrimSpace(lines[0])
	artist := strings.TrimSpace(lines[1])
	durationStr := strings.TrimSpace(lines[2])

	// Parse duration
	duration, err := strconv.Atoi(durationStr)
	if err != nil {
		log.Printf("[WARN] YtDlpService: Failed to parse duration '%s', defaulting to 0: %v", durationStr, err)
		duration = 0
	}

	// Clean up the title and artist
	title = s.cleanMetadata(title)
	artist = s.cleanMetadata(artist)

	song := &models.Song{
		YouTubeID: youtubeID,
		Title:     title,
		Artist:    artist,
		Album:     "", // Not available from yt-dlp
		Duration:  duration,
		FilePath:  fmt.Sprintf("songs/%s.mp3", youtubeID), // Will be set when downloaded
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return song, nil
}

// IsVideoAvailable checks if a YouTube video is available for download
func (s *YtDlpService) IsVideoAvailable(ctx context.Context, youtubeID string) (bool, error) {
	url := fmt.Sprintf("https://www.youtube.com/watch?v=%s", youtubeID)

	// Create context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Build yt-dlp command to check availability
	cmd := exec.CommandContext(ctxWithTimeout, s.ytDlpPath,
		"--simulate",       // Don't download, just simulate
		"--quiet",          // Suppress most output
		"--no-warnings",    // Suppress warnings
		"--no-playlist",    // Don't process playlists
		url,
	)

	err := cmd.Run()
	return err == nil, nil
}

// cleanMetadata removes common unwanted patterns from metadata
func (s *YtDlpService) cleanMetadata(text string) string {
	// Remove common patterns like (Official Video), [HD], etc.
	patterns := []string{
		`\(Official Video\)`,
		`\(Official Music Video\)`,
		`\(Official Audio\)`,
		`\[Official Video\]`,
		`\[Official Music Video\]`,
		`\[Official Audio\]`,
		`\(HD\)`,
		`\[HD\]`,
		`\(4K\)`,
		`\[4K\]`,
		`\(Lyrics\)`,
		`\[Lyrics\]`,
		`\(Lyric Video\)`,
		`\[Lyric Video\]`,
	}

	cleaned := text
	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		cleaned = re.ReplaceAllString(cleaned, "")
	}

	// Clean up extra whitespace
	cleaned = regexp.MustCompile(`\s+`).ReplaceAllString(cleaned, " ")
	cleaned = strings.TrimSpace(cleaned)

	return cleaned
}

// MockYtDlpService provides a mock implementation for testing
type MockYtDlpService struct {
	downloadDelay time.Duration
	shouldFail    bool
}

// NewMockYtDlpService creates a new mock service for testing
func NewMockYtDlpService(downloadDelay time.Duration, shouldFail bool) *MockYtDlpService {
	return &MockYtDlpService{
		downloadDelay: downloadDelay,
		shouldFail:    shouldFail,
	}
}

func (m *MockYtDlpService) DownloadAudio(ctx context.Context, youtubeID string, outputDir string) (string, error) {
	if m.shouldFail {
		return "", fmt.Errorf("mock download failed")
	}

	// Simulate download delay
	select {
	case <-time.After(m.downloadDelay):
	case <-ctx.Done():
		return "", ctx.Err()
	}

	// Create mock file
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", err
	}

	mockPath := filepath.Join(outputDir, fmt.Sprintf("%s.mp3", youtubeID))
	file, err := os.Create(mockPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Write some mock data
	_, err = io.WriteString(file, "mock audio data")
	if err != nil {
		return "", err
	}

	return mockPath, nil
}

func (m *MockYtDlpService) GetVideoInfo(ctx context.Context, youtubeID string) (*models.Song, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("mock get info failed")
	}

	return &models.Song{
		YouTubeID: youtubeID,
		Title:     fmt.Sprintf("Mock Song %s", youtubeID),
		Artist:    "Mock Artist",
		Album:     "",
		Duration:  180, // 3 minutes
		FilePath:  fmt.Sprintf("songs/%s.mp3", youtubeID),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (m *MockYtDlpService) IsVideoAvailable(ctx context.Context, youtubeID string) (bool, error) {
	if m.shouldFail {
		return false, fmt.Errorf("mock availability check failed")
	}
	return true, nil
}