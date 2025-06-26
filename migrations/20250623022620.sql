-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Create "new_playlists" table
CREATE TABLE `new_playlists` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `name` text NOT NULL,
  `description` text NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL
);
-- Copy rows from old table "playlists" to new temporary table "new_playlists"
INSERT INTO `new_playlists` (`id`, `name`, `description`, `created_at`, `updated_at`) SELECT `id`, `name`, `description`, `created_at`, `updated_at` FROM `playlists`;
-- Drop "playlists" table after copying rows
DROP TABLE `playlists`;
-- Rename temporary table "new_playlists" to "playlists"
ALTER TABLE `new_playlists` RENAME TO `playlists`;
-- Create index "idx_playlists_name_unique" to table: "playlists"
CREATE UNIQUE INDEX `idx_playlists_name_unique` ON `playlists` (`name`);
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
