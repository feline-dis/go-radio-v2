package storage

import (
	"database/sql"
	"time"

	"github.com/feline-dis/go-radio-v2/internal/models"
	"github.com/google/uuid"
)

type SQLitePlaylistRepository struct {
	db *sql.DB
}

func NewSQLitePlaylistRepository(dbPath string) (*SQLitePlaylistRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	repo := &SQLitePlaylistRepository{db: db}
	if err := repo.createTables(); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *SQLitePlaylistRepository) createTables() error {
	playlistTablesSQL := `
	CREATE TABLE IF NOT EXISTS playlists (
		id TEXT PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		description TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS playlist_songs (
		playlist_id TEXT NOT NULL,
		youtube_id TEXT NOT NULL,
		position INTEGER NOT NULL,
		created_at DATETIME NOT NULL,
		PRIMARY KEY (playlist_id, youtube_id),
		FOREIGN KEY (playlist_id) REFERENCES playlists(id) ON DELETE CASCADE,
		FOREIGN KEY (youtube_id) REFERENCES songs(youtube_id) ON DELETE CASCADE
	);
	
	CREATE INDEX IF NOT EXISTS idx_playlist_songs_position ON playlist_songs(playlist_id, position);
	`

	_, err := r.db.Exec(playlistTablesSQL)
	return err
}

func (r *SQLitePlaylistRepository) Create(playlist *models.Playlist) error {
	query := `
		INSERT INTO playlists (id, name, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`

	now := time.Now()
	id := uuid.New().String()
	_, err := r.db.Exec(query, id, playlist.Name, playlist.Description, now, now)
	if err != nil {
		return err
	}

	playlist.ID = id
	playlist.CreatedAt = now
	playlist.UpdatedAt = now
	return nil
}

func (r *SQLitePlaylistRepository) GetByID(id string) (*models.Playlist, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM playlists
		WHERE id = ?
	`

	playlist := &models.Playlist{}
	err := r.db.QueryRow(query, id).Scan(
		&playlist.ID,
		&playlist.Name,
		&playlist.Description,
		&playlist.CreatedAt,
		&playlist.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return playlist, nil
}

func (r *SQLitePlaylistRepository) GetByName(name string) (*models.Playlist, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM playlists
		WHERE name = ?
	`

	playlist := &models.Playlist{}
	err := r.db.QueryRow(query, name).Scan(
		&playlist.ID,
		&playlist.Name,
		&playlist.Description,
		&playlist.CreatedAt,
		&playlist.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return playlist, nil
}

func (r *SQLitePlaylistRepository) GetAll() ([]*models.Playlist, error) {
	query := `
		SELECT p.id, p.name, p.description, p.created_at, p.updated_at, 
		       COALESCE(COUNT(ps.playlist_id), 0) as song_count
		FROM playlists p
		LEFT JOIN playlist_songs ps ON p.id = ps.playlist_id
		GROUP BY p.id, p.name, p.description, p.created_at, p.updated_at
		ORDER BY p.name
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var playlists []*models.Playlist
	for rows.Next() {
		playlist := &models.Playlist{}
		err := rows.Scan(
			&playlist.ID,
			&playlist.Name,
			&playlist.Description,
			&playlist.CreatedAt,
			&playlist.UpdatedAt,
			&playlist.SongCount,
		)
		if err != nil {
			return nil, err
		}
		playlists = append(playlists, playlist)
	}

	return playlists, nil
}

func (r *SQLitePlaylistRepository) Update(playlist *models.Playlist) error {
	query := `
		UPDATE playlists
		SET name = ?, description = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	_, err := r.db.Exec(query, playlist.Name, playlist.Description, now, playlist.ID)
	if err != nil {
		return err
	}

	playlist.UpdatedAt = now
	return nil
}

func (r *SQLitePlaylistRepository) Delete(id string) error {
	// Due to CASCADE, playlist_songs will be deleted automatically
	query := `DELETE FROM playlists WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *SQLitePlaylistRepository) GetFirstPlaylist() (*models.Playlist, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM playlists
		ORDER BY created_at ASC
		LIMIT 1
	`

	playlist := &models.Playlist{}
	err := r.db.QueryRow(query).Scan(
		&playlist.ID,
		&playlist.Name,
		&playlist.Description,
		&playlist.CreatedAt,
		&playlist.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return playlist, nil
}

func (r *SQLitePlaylistRepository) AddSong(playlistID string, youtubeID string, position int) error {
	query := `
		INSERT INTO playlist_songs (playlist_id, youtube_id, position, created_at)
		VALUES (?, ?, ?, ?)
	`

	_, err := r.db.Exec(query, playlistID, youtubeID, position, time.Now())
	return err
}

func (r *SQLitePlaylistRepository) RemoveSong(playlistID string, youtubeID string) error {
	query := `
		DELETE FROM playlist_songs
		WHERE playlist_id = ? AND youtube_id = ?
	`

	_, err := r.db.Exec(query, playlistID, youtubeID)
	return err
}

func (r *SQLitePlaylistRepository) GetSongs(playlistID string) ([]*models.Song, error) {
	query := `
		SELECT s.youtube_id, s.title, s.artist, s.album, s.duration, s.file_path, 
		       s.last_played, s.play_count, s.created_at, s.updated_at
		FROM playlist_songs ps
		JOIN songs s ON ps.youtube_id = s.youtube_id
		WHERE ps.playlist_id = ?
		ORDER BY ps.position ASC
	`

	rows, err := r.db.Query(query, playlistID)
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

func (r *SQLitePlaylistRepository) UpdateSongPosition(playlistID string, youtubeID string, newPosition int) error {
	query := `
		UPDATE playlist_songs
		SET position = ?
		WHERE playlist_id = ? AND youtube_id = ?
	`

	_, err := r.db.Exec(query, newPosition, playlistID, youtubeID)
	return err
}

func (r *SQLitePlaylistRepository) Close() error {
	return r.db.Close()
}