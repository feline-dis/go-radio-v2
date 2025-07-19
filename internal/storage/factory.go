package storage

import (
	"fmt"

	"github.com/feline-dis/go-radio-v2/internal/config"
)

type StorageFactory struct {
	config *config.Config
}

func NewStorageFactory(cfg *config.Config) *StorageFactory {
	return &StorageFactory{config: cfg}
}

func (f *StorageFactory) CreateSongRepository() (SongRepository, error) {
	switch f.config.Storage.MetadataStorageType {
	case "sqlite":
		return NewSQLiteSongRepository(f.config.Storage.SQLiteDBPath)
	case "json":
		// TODO: Implement JSON-based song repository
		return nil, fmt.Errorf("JSON storage not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported metadata storage type: %s", f.config.Storage.MetadataStorageType)
	}
}

func (f *StorageFactory) CreatePlaylistRepository() (PlaylistRepository, error) {
	switch f.config.Storage.MetadataStorageType {
	case "sqlite":
		return NewSQLitePlaylistRepository(f.config.Storage.SQLiteDBPath)
	case "json":
		// TODO: Implement JSON-based playlist repository
		return nil, fmt.Errorf("JSON storage not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported metadata storage type: %s", f.config.Storage.MetadataStorageType)
	}
}

func (f *StorageFactory) CreateFileStorage() (FileStorage, error) {
	switch f.config.Storage.FileStorageType {
	case "local":
		return NewLocalFileStorage(f.config.Storage.LocalDataDir)
	case "s3":
		// Create S3 storage directly without importing services
		return NewS3FileStorage(f.config)
	default:
		return nil, fmt.Errorf("unsupported file storage type: %s", f.config.Storage.FileStorageType)
	}
}

// ValidateConfig checks if the storage configuration is valid
func (f *StorageFactory) ValidateConfig() error {
	// Validate file storage type
	if f.config.Storage.FileStorageType != "local" && f.config.Storage.FileStorageType != "s3" {
		return fmt.Errorf("invalid file storage type: %s (must be 'local' or 's3')", f.config.Storage.FileStorageType)
	}

	// Validate metadata storage type
	if f.config.Storage.MetadataStorageType != "sqlite" && f.config.Storage.MetadataStorageType != "json" {
		return fmt.Errorf("invalid metadata storage type: %s (must be 'sqlite' or 'json')", f.config.Storage.MetadataStorageType)
	}

	// Validate S3 config if using S3
	if f.config.Storage.FileStorageType == "s3" {
		if f.config.AWS.BucketName == "" {
			return fmt.Errorf("S3 bucket name is required when using S3 storage")
		}
		if f.config.AWS.AccessKeyID == "" || f.config.AWS.SecretAccessKey == "" {
			return fmt.Errorf("AWS credentials are required when using S3 storage")
		}
	}

	return nil
}