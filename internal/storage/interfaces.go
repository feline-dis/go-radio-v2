package storage

import (
	"context"
	"io"
	"time"

	"github.com/feline-dis/go-radio-v2/internal/models"
)

// SongRepository defines the interface for song metadata storage
type SongRepository interface {
	Create(song *models.Song) error
	GetByYouTubeID(youtubeID string) (*models.Song, error)
	UpdatePlayStats(youtubeID string) error
	GetRandomSong() (*models.Song, error)
	GetLeastPlayedSong() (*models.Song, error)
	GetAll() ([]*models.Song, error)
	Delete(youtubeID string) error
}

// PlaylistRepository defines the interface for playlist storage
type PlaylistRepository interface {
	Create(playlist *models.Playlist) error
	GetByID(id string) (*models.Playlist, error)
	GetByName(name string) (*models.Playlist, error)
	GetAll() ([]*models.Playlist, error)
	Update(playlist *models.Playlist) error
	Delete(id string) error
	GetFirstPlaylist() (*models.Playlist, error)
	
	// Song management in playlists
	AddSong(playlistID string, youtubeID string, position int) error
	RemoveSong(playlistID string, youtubeID string) error
	GetSongs(playlistID string) ([]*models.Song, error)
	UpdateSongPosition(playlistID string, youtubeID string, newPosition int) error
}

// FileStorage defines the interface for audio file storage
type FileStorage interface {
	UploadFile(ctx context.Context, key string, body io.Reader) error
	GetFile(ctx context.Context, key string) (io.ReadCloser, error)
	GetFilePath(key string) (string, error) // For local storage, returns file path
	GetPresignedURL(ctx context.Context, key string, expires time.Duration) (string, error) // For S3
	DeleteFile(ctx context.Context, key string) error
	FileExists(ctx context.Context, key string) (bool, error)
}

// StorageType defines the available storage backends
type StorageType string

const (
	StorageTypeLocal  StorageType = "local"
	StorageTypeS3     StorageType = "s3"
	StorageTypeSQLite StorageType = "sqlite"
	StorageTypeJSON   StorageType = "json"
)

// StorageConfig holds configuration for different storage backends
type StorageConfig struct {
	// File storage
	FileStorageType StorageType
	LocalDataDir    string
	S3Config        *S3Config
	
	// Metadata storage
	MetadataStorageType StorageType
	SQLiteDBPath        string
	JSONDataDir         string
}

type S3Config struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
}