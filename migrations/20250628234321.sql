-- Set comment to schema: "public"
COMMENT ON SCHEMA "public" IS 'Radio application database schema';
-- Create "playlists" table
CREATE TABLE "public"."playlists" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" text NOT NULL,
  "description" text NOT NULL,
  "created_at" timestamp NOT NULL,
  "updated_at" timestamp NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_playlists_name_unique" to table: "playlists"
CREATE UNIQUE INDEX "idx_playlists_name_unique" ON "public"."playlists" ("name");
-- Create "songs" table
CREATE TABLE "public"."songs" (
  "youtube_id" text NOT NULL,
  "title" text NOT NULL,
  "artist" text NOT NULL,
  "album" text NOT NULL,
  "duration" integer NOT NULL,
  "s3_key" text NOT NULL,
  "last_played" timestamp NOT NULL,
  "play_count" integer NOT NULL DEFAULT 0,
  "created_at" timestamp NOT NULL,
  "updated_at" timestamp NOT NULL,
  PRIMARY KEY ("youtube_id")
);
-- Create index "idx_songs_last_played" to table: "songs"
CREATE INDEX "idx_songs_last_played" ON "public"."songs" ("last_played");
-- Create index "idx_songs_play_count" to table: "songs"
CREATE INDEX "idx_songs_play_count" ON "public"."songs" ("play_count");
-- Create "playlist_songs" table
CREATE TABLE "public"."playlist_songs" (
  "playlist_id" uuid NOT NULL,
  "youtube_id" text NOT NULL,
  "position" integer NOT NULL,
  "created_at" timestamp NOT NULL,
  PRIMARY KEY ("playlist_id", "youtube_id"),
  CONSTRAINT "fk_playlist_songs_playlist" FOREIGN KEY ("playlist_id") REFERENCES "public"."playlists" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "fk_playlist_songs_song" FOREIGN KEY ("youtube_id") REFERENCES "public"."songs" ("youtube_id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_playlist_songs_position" to table: "playlist_songs"
CREATE INDEX "idx_playlist_songs_position" ON "public"."playlist_songs" ("playlist_id", "position");
