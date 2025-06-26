package models

import (
	"time"
)

// Song represents a song's metadata in the database
type Song struct {
	YouTubeID  string    `json:"youtube_id" db:"youtube_id"`
	Title      string    `json:"title" db:"title"`
	Artist     string    `json:"artist" db:"artist"`
	Album      string    `json:"album" db:"album"`
	Duration   int       `json:"duration" db:"duration"` // Duration in seconds
	S3Key      string    `json:"s3_key" db:"s3_key"`
	LastPlayed time.Time `json:"last_played" db:"last_played"`
	PlayCount  int       `json:"play_count" db:"play_count"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// Playlist represents a playlist in the database
type Playlist struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// PlaylistSong represents the many-to-many relationship between playlists and songs
type PlaylistSong struct {
	PlaylistID int       `json:"playlist_id" db:"playlist_id"`
	YouTubeID  string    `json:"youtube_id" db:"youtube_id"`
	Position   int       `json:"position" db:"position"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// PlaybackState represents the current playback state
type PlaybackState struct {
	CurrentSong      *Song
	NextSong         *Song
	StartTime        time.Time
	Paused           bool
	PauseTime        time.Time
	CurrentPlaylist  *Playlist
	CurrentSongIndex int
	Queue            []*Song
	ShuffledSongs    []*Song // Shuffled version of playlist songs
	IsShuffled       bool    // Whether we're currently playing shuffled
}

// QueueInfo represents the current queue information
type QueueInfo struct {
	CurrentSong *Song
	NextSong    *Song
	Queue       []*Song
	Playlist    *Playlist
	Remaining   float64 // Remaining time in seconds
	StartTime   time.Time
}
