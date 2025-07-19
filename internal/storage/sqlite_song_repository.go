package storage

import (
	"database/sql"
	"time"

	"github.com/feline-dis/go-radio-v2/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteSongRepository struct {
	db *sql.DB
}

func NewSQLiteSongRepository(dbPath string) (*SQLiteSongRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	repo := &SQLiteSongRepository{db: db}
	if err := repo.createTables(); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *SQLiteSongRepository) createTables() error {
	songTableSQL := `
	CREATE TABLE IF NOT EXISTS songs (
		youtube_id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		artist TEXT,
		album TEXT,
		duration INTEGER NOT NULL,
		file_path TEXT NOT NULL,
		last_played DATETIME,
		play_count INTEGER DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	
	CREATE INDEX IF NOT EXISTS idx_songs_play_count ON songs(play_count);
	CREATE INDEX IF NOT EXISTS idx_songs_last_played ON songs(last_played);
	`

	_, err := r.db.Exec(songTableSQL)
	return err
}

func (r *SQLiteSongRepository) Create(song *models.Song) error {
	query := `
		INSERT INTO songs (
			youtube_id, title, artist, album, duration, file_path,
			last_played, play_count, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	_, err := r.db.Exec(query,
		song.YouTubeID,
		song.Title,
		song.Artist,
		song.Album,
		song.Duration,
		song.FilePath, // Changed from S3Key to FilePath
		song.LastPlayed,
		song.PlayCount,
		now,
		now,
	)

	return err
}

func (r *SQLiteSongRepository) GetByYouTubeID(youtubeID string) (*models.Song, error) {
	query := `
		SELECT youtube_id, title, artist, album, duration, file_path,
			   last_played, play_count, created_at, updated_at
		FROM songs
		WHERE youtube_id = ?
	`

	song := &models.Song{}
	err := r.db.QueryRow(query, youtubeID).Scan(
		&song.YouTubeID,
		&song.Title,
		&song.Artist,
		&song.Album,
		&song.Duration,
		&song.FilePath,
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

func (r *SQLiteSongRepository) UpdatePlayStats(youtubeID string) error {
	query := `
		UPDATE songs
		SET last_played = ?,
			play_count = play_count + 1,
			updated_at = ?
		WHERE youtube_id = ?
	`

	now := time.Now()
	_, err := r.db.Exec(query, now, now, youtubeID)
	return err
}

func (r *SQLiteSongRepository) GetRandomSong() (*models.Song, error) {
	query := `
		SELECT youtube_id, title, artist, album, duration, file_path,
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
		&song.FilePath,
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

func (r *SQLiteSongRepository) GetLeastPlayedSong() (*models.Song, error) {
	query := `
		SELECT youtube_id, title, artist, album, duration, file_path,
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
		&song.FilePath,
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

func (r *SQLiteSongRepository) GetAll() ([]*models.Song, error) {
	query := `
		SELECT youtube_id, title, artist, album, duration, file_path,
			   last_played, play_count, created_at, updated_at
		FROM songs
		ORDER BY title ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var songs []*models.Song
	for rows.Next() {
		song := &models.Song{}
		err := rows.Scan(
			&song.YouTubeID,
			&song.Title,
			&song.Artist,
			&song.Album,
			&song.Duration,
			&song.FilePath,
			&song.LastPlayed,
			&song.PlayCount,
			&song.CreatedAt,
			&song.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		songs = append(songs, song)
	}

	return songs, nil
}

func (r *SQLiteSongRepository) Delete(youtubeID string) error {
	query := `DELETE FROM songs WHERE youtube_id = ?`
	_, err := r.db.Exec(query, youtubeID)
	return err
}

func (r *SQLiteSongRepository) Close() error {
	return r.db.Close()
}