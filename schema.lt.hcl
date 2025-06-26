schema "main" {
  comment = "Radio application database schema"
}

table "songs" {
  schema = schema.main
  column "youtube_id" {
    type = text
    null = false
  }
  column "title" {
    type = text
    null = false
  }
  column "artist" {
    type = text
    null = false
  }
  column "album" {
    type = text
    null = false
  }
  column "duration" {
    type = integer
    null = false
  }
  column "s3_key" {
    type = text
    null = false
  }
  column "last_played" {
    type = datetime
    null = false
  }
  column "play_count" {
    type = integer
    default = 0
    null = false
  }
  column "created_at" {
    type = datetime
    null = false
  }
  column "updated_at" {
    type = datetime
    null = false
  }
  primary_key {
    columns = [column.youtube_id]
  }
  index "idx_songs_play_count" {
    columns = [column.play_count]
  }
  index "idx_songs_last_played" {
    columns = [column.last_played]
  }
}

table "playlists" {
  schema = schema.main
  column "id" {
    type = integer
    null = false
    auto_increment = true
  }
  column "name" {
    type = text
    null = false
  }
  column "description" {
    type = text
    null = false
  }
  column "created_at" {
    type = datetime
    null = false
  }
  column "updated_at" {
    type = datetime
    null = false
  }
  primary_key {
    columns = [column.id]
  }
  index "idx_playlists_name_unique" {
    columns = [column.name]
    unique = true
  }
}

table "playlist_songs" {
  schema = schema.main
  column "playlist_id" {
    type = integer
    null = false
  }
  column "youtube_id" {
    type = text
    null = false
  }
  column "position" {
    type = integer
    null = false
  }
  column "created_at" {
    type = datetime
    null = false
  }
  primary_key {
    columns = [column.playlist_id, column.youtube_id]
  }
  foreign_key "fk_playlist_songs_playlist" {
    columns = [column.playlist_id]
    ref_columns = [table.playlists.column.id]
    on_delete = CASCADE
  }
  foreign_key "fk_playlist_songs_song" {
    columns = [column.youtube_id]
    ref_columns = [table.songs.column.youtube_id]
    on_delete = CASCADE
  }
  index "idx_playlist_songs_position" {
    columns = [column.playlist_id, column.position]
  }
} 