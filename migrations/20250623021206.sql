-- Create "songs" table
CREATE TABLE `songs` (
  `youtube_id` text NOT NULL,
  `title` text NOT NULL,
  `artist` text NOT NULL,
  `album` text NOT NULL,
  `duration` integer NOT NULL,
  `s3_key` text NOT NULL,
  `last_played` datetime NOT NULL,
  `play_count` integer NOT NULL DEFAULT 0,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`youtube_id`)
);
-- Create index "idx_songs_play_count" to table: "songs"
CREATE INDEX `idx_songs_play_count` ON `songs` (`play_count`);
-- Create index "idx_songs_last_played" to table: "songs"
CREATE INDEX `idx_songs_last_played` ON `songs` (`last_played`);
-- Create "playlists" table
CREATE TABLE `playlists` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `creator` text NOT NULL,
  `name` text NOT NULL,
  `description` text NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL
);
-- Create index "idx_playlists_name_unique" to table: "playlists"
CREATE UNIQUE INDEX `idx_playlists_name_unique` ON `playlists` (`name`);
-- Create "playlist_songs" table
CREATE TABLE `playlist_songs` (
  `playlist_id` integer NOT NULL,
  `youtube_id` text NOT NULL,
  `position` integer NOT NULL,
  `created_at` datetime NOT NULL,
  PRIMARY KEY (`playlist_id`, `youtube_id`),
  CONSTRAINT `fk_playlist_songs_playlist` FOREIGN KEY (`playlist_id`) REFERENCES `playlists` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_playlist_songs_song` FOREIGN KEY (`youtube_id`) REFERENCES `songs` (`youtube_id`) ON DELETE CASCADE
);
-- Create index "idx_playlist_songs_position" to table: "playlist_songs"
CREATE INDEX `idx_playlist_songs_position` ON `playlist_songs` (`playlist_id`, `position`);
