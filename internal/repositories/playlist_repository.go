package repositories

import (
	"database/sql"
	"time"

	"github.com/feline-dis/go-radio-v2/internal/models"
)

type PlaylistRepository struct {
	db *sql.DB
}

func NewPlaylistRepository(db *sql.DB) *PlaylistRepository {
	return &PlaylistRepository{db: db}
}

func (r *PlaylistRepository) Create(playlist *models.Playlist) error {
	query := `
		INSERT INTO playlists (name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	now := time.Now()
	var id string
	err := r.db.QueryRow(query,
		playlist.Name,
		playlist.Description,
		now,
		now,
	).Scan(&id)
	if err != nil {
		return err
	}

	playlist.ID = id
	playlist.CreatedAt = now
	playlist.UpdatedAt = now
	return nil
}

func (r *PlaylistRepository) GetByID(id string) (*models.Playlist, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM playlists
		WHERE id = $1
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

func (r *PlaylistRepository) GetAll() ([]*models.Playlist, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM playlists
		ORDER BY name
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
		)
		if err != nil {
			return nil, err
		}
		playlists = append(playlists, playlist)
	}

	return playlists, nil
}

func (r *PlaylistRepository) AddSong(playlistID string, youtubeID string, position int) error {
	query := `
		INSERT INTO playlist_songs (playlist_id, youtube_id, position, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(query, playlistID, youtubeID, position, time.Now())
	return err
}

func (r *PlaylistRepository) GetSongs(playlistID string) ([]*models.Song, error) {
	query := `
		SELECT s.youtube_id, s.title, s.artist, s.album, s.duration, s.s3_key, s.last_played, s.play_count, s.created_at, s.updated_at
		FROM playlist_songs ps
		JOIN songs s ON ps.youtube_id = s.youtube_id
		WHERE ps.playlist_id = $1
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
			&song.S3Key,
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

func (r *PlaylistRepository) RemoveSong(playlistID string, youtubeID string) error {
	query := `
		DELETE FROM playlist_songs
		WHERE playlist_id = $1 AND youtube_id = $2
	`

	_, err := r.db.Exec(query, playlistID, youtubeID)
	return err
}

func (r *PlaylistRepository) UpdateSongPosition(playlistID string, youtubeID string, newPosition int) error {
	query := `
		UPDATE playlist_songs
		SET position = $1
		WHERE playlist_id = $2 AND youtube_id = $3
	`

	_, err := r.db.Exec(query, newPosition, playlistID, youtubeID)
	return err
}

func (r *PlaylistRepository) GetByName(name string) (*models.Playlist, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM playlists
		WHERE name = $1
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

func (r *PlaylistRepository) GetFirstPlaylist() (*models.Playlist, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM playlists
		ORDER BY id
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
