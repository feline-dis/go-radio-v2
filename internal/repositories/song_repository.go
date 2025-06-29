package repositories

import (
	"database/sql"
	"time"

	"github.com/feline-dis/go-radio-v2/internal/models"
)

type SongRepository struct {
	db *sql.DB
}

func NewSongRepository(db *sql.DB) *SongRepository {
	return &SongRepository{db: db}
}

func (r *SongRepository) Create(song *models.Song) error {
	query := `
		INSERT INTO songs (
			youtube_id, title, artist, album, duration, s3_key,
			last_played, play_count, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	now := time.Now()
	_, err := r.db.Exec(query,
		song.YouTubeID,
		song.Title,
		song.Artist,
		song.Album,
		song.Duration,
		song.S3Key,
		song.LastPlayed,
		song.PlayCount,
		now,
		now,
	)

	return err
}

func (r *SongRepository) GetByYouTubeID(youtubeID string) (*models.Song, error) {
	query := `
		SELECT youtube_id, title, artist, album, duration, s3_key,
			   last_played, play_count, created_at, updated_at
		FROM songs
		WHERE youtube_id = $1
	`

	song := &models.Song{}
	err := r.db.QueryRow(query, youtubeID).Scan(
		&song.YouTubeID,
		&song.Title,
		&song.Artist,
		&song.Album,
		&song.Duration,
		&song.S3Key,
		&song.LastPlayed,
		&song.PlayCount,
		&song.CreatedAt,
		&song.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return song, nil
}

func (r *SongRepository) UpdatePlayStats(youtubeID string) error {
	query := `
		UPDATE songs
		SET last_played = $1,
			play_count = play_count + 1,
			updated_at = $2
		WHERE youtube_id = $3
	`

	now := time.Now()
	_, err := r.db.Exec(query, now, now, youtubeID)
	return err
}

func (r *SongRepository) GetRandomSong() (*models.Song, error) {
	query := `
		SELECT youtube_id, title, artist, album, duration, s3_key,
			   last_played, play_count, created_at, updated_at
		FROM songs
		ORDER BY RANDOM()
		LIMIT 1
	`

	song := &models.Song{}
	err := r.db.QueryRow(query).Scan(
		&song.YouTubeID,
		&song.Title,
		&song.Artist,
		&song.Album,
		&song.Duration,
		&song.S3Key,
		&song.LastPlayed,
		&song.PlayCount,
		&song.CreatedAt,
		&song.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return song, nil
}

func (r *SongRepository) GetLeastPlayedSong() (*models.Song, error) {
	query := `
		SELECT youtube_id, title, artist, album, duration, s3_key,
			   last_played, play_count, created_at, updated_at
		FROM songs
		ORDER BY play_count ASC, last_played ASC
		LIMIT 1
	`

	song := &models.Song{}
	err := r.db.QueryRow(query).Scan(
		&song.YouTubeID,
		&song.Title,
		&song.Artist,
		&song.Album,
		&song.Duration,
		&song.S3Key,
		&song.LastPlayed,
		&song.PlayCount,
		&song.CreatedAt,
		&song.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return song, nil
}
