package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type LocalFileStorage struct {
	dataDir string
}

func NewLocalFileStorage(dataDir string) (*LocalFileStorage, error) {
	// Create data directory if it doesn't exist
	audioDir := filepath.Join(dataDir, "audio")
	if err := os.MkdirAll(audioDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create audio directory: %w", err)
	}

	return &LocalFileStorage{
		dataDir: dataDir,
	}, nil
}

func (l *LocalFileStorage) UploadFile(ctx context.Context, key string, body io.Reader) error {
	filePath := l.getFilePath(key)
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()

	// Copy the body to the file
	_, err = io.Copy(file, body)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return nil
}

func (l *LocalFileStorage) GetFile(ctx context.Context, key string) (io.ReadCloser, error) {
	filePath := l.getFilePath(key)
	
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", key)
		}
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}

	return file, nil
}

func (l *LocalFileStorage) GetFilePath(key string) (string, error) {
	filePath := l.getFilePath(key)
	
	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file not found: %s", key)
		}
		return "", fmt.Errorf("failed to stat file %s: %w", filePath, err)
	}

	return filePath, nil
}

func (l *LocalFileStorage) GetPresignedURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	// For local storage, we return the file path
	// The HTTP server will need to serve these files directly
	return l.getFilePath(key), nil
}

func (l *LocalFileStorage) DeleteFile(ctx context.Context, key string) error {
	filePath := l.getFilePath(key)
	
	err := os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file %s: %w", filePath, err)
	}

	return nil
}

func (l *LocalFileStorage) FileExists(ctx context.Context, key string) (bool, error) {
	filePath := l.getFilePath(key)
	
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence %s: %w", filePath, err)
	}

	return true, nil
}

func (l *LocalFileStorage) getFilePath(key string) string {
	return filepath.Join(l.dataDir, "audio", key)
}

// GetAudioDir returns the directory where audio files are stored
func (l *LocalFileStorage) GetAudioDir() string {
	return filepath.Join(l.dataDir, "audio")
}