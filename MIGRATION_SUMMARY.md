# SQLite to PostgreSQL Migration Summary

This document summarizes all the changes made to migrate the Go Radio v2 application from SQLite to PostgreSQL.

## Changes Made

### 1. Dependencies Updated
- **go.mod**: Replaced `modernc.org/sqlite v1.28.0` with `github.com/lib/pq v1.10.9`
- **go.sum**: Updated automatically via `go mod tidy`

### 2. Configuration Changes
- **internal/config/config.go**: 
  - Updated `DatabaseConfig` struct to include PostgreSQL connection parameters
  - Added environment variables: `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, `POSTGRES_SSLMODE`
  - Removed `SQLITE_DB_PATH` configuration

### 3. Database Connection Updates
- **cmd/server/main.go**:
  - Changed import from `_ "modernc.org/sqlite"` to `_ "github.com/lib/pq"`
  - Updated database connection string to PostgreSQL format
  - Increased connection pool settings (MaxOpenConns: 1→25, MaxIdleConns: 1→5)
  - Removed SQLite-specific connection pool limitations

- **cmd/download/main.go**:
  - Updated import and database connection similar to server

### 4. SQL Query Updates
- **internal/repositories/song_repository.go**:
  - Changed all parameter placeholders from `?` to `$1`, `$2`, etc.
  - Updated all SQL queries to use PostgreSQL syntax

- **internal/repositories/playlist_repository.go**:
  - Changed all parameter placeholders from `?` to `$1`, `$2`, etc.
  - Updated all SQL queries to use PostgreSQL syntax

### 5. Database Schema Migration
- **atlas.hcl**: Updated to use PostgreSQL connection string
- **migrations/20250101000000_postgres_schema.sql**: Created new PostgreSQL-compatible schema
  - Changed `INTEGER PRIMARY KEY AUTOINCREMENT` to `SERIAL PRIMARY KEY`
  - Changed `text` to `VARCHAR(255)` or `TEXT`
  - Changed `datetime` to `TIMESTAMP`
  - Updated index and constraint syntax for PostgreSQL

### 6. Docker Configuration
- **docker-compose.yml**:
  - Added PostgreSQL service with health checks
  - Updated backend service to use PostgreSQL environment variables
  - Added dependency on PostgreSQL service
  - Removed SQLite volume mount
  - Updated volume name from `go_radio_data` to `postgres_data`

- **Dockerfile**:
  - Added PostgreSQL client tools for debugging
  - Removed SQLite-specific file copying and directory creation
  - Removed SQLite database file copy

### 7. Deployment Configuration
- **fly.toml**: Updated environment variables to use PostgreSQL configuration
- **Makefile**: Updated database backup/restore commands to use PostgreSQL tools
- **.github/workflows/deploy-environments.yml**: Updated backup command to use `pg_dump`

### 8. Documentation Updates
- **README.md**: Completely updated to reflect PostgreSQL usage
  - Updated tech stack section
  - Updated environment variables table
  - Updated setup instructions
  - Updated architecture diagram

### 9. Cleanup
- **Removed files**:
  - `migrations/20250623021206.sql` (old SQLite schema)
  - `migrations/20250623022620.sql` (old SQLite migration)
  - `data/radio.db` (SQLite database file)

- **Updated files**:
  - `.dockerignore`: Removed SQLite-specific patterns

## Environment Variables

### New PostgreSQL Variables
```bash
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=go_radio
POSTGRES_SSLMODE=disable
```

### Removed Variables
```bash
SQLITE_DB_PATH=./data/radio.db  # No longer needed
```

## Database Schema Changes

### Songs Table
```sql
-- Before (SQLite)
CREATE TABLE songs (
  youtube_id text NOT NULL,
  title text NOT NULL,
  -- ... other fields
);

-- After (PostgreSQL)
CREATE TABLE songs (
  youtube_id VARCHAR(255) NOT NULL,
  title TEXT NOT NULL,
  -- ... other fields
);
```

### Playlists Table
```sql
-- Before (SQLite)
CREATE TABLE playlists (
  id integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  -- ... other fields
);

-- After (PostgreSQL)
CREATE TABLE playlists (
  id SERIAL PRIMARY KEY,
  -- ... other fields
);
```

## Connection String Format

### Before (SQLite)
```go
db, err := sql.Open("sqlite", "./data/radio.db")
```

### After (PostgreSQL)
```go
dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
    cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
    cfg.Database.Password, cfg.Database.DBName, cfg.Database.SSLMode)
db, err := sql.Open("postgres", dsn)
```

## Query Parameter Changes

### Before (SQLite)
```sql
SELECT * FROM songs WHERE youtube_id = ?
INSERT INTO songs VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
```

### After (PostgreSQL)
```sql
SELECT * FROM songs WHERE youtube_id = $1
INSERT INTO songs VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
```

## Deployment Changes

### Local Development
```bash
# Start with PostgreSQL
make dev-compose

# Database operations
make db-backup    # Uses pg_dump
make db-restore   # Uses psql
```

### Production
- PostgreSQL service must be available
- Environment variables must be configured for PostgreSQL connection
- Database migrations will create PostgreSQL-compatible schema

## Testing the Migration

1. **Start the application**:
   ```bash
   make dev-compose
   ```

2. **Verify database connection**:
   - Check application logs for successful database connection
   - Verify migrations ran successfully

3. **Test functionality**:
   - Create playlists
   - Add songs
   - Test radio playback
   - Verify WebSocket connections

## Rollback Considerations

If rollback is needed:
1. Restore from SQLite backup (if available)
2. Revert all code changes
3. Restore old migration files
4. Update dependencies back to SQLite

## Benefits of PostgreSQL Migration

1. **Better concurrency**: Supports multiple simultaneous connections
2. **ACID compliance**: Full transaction support
3. **Advanced features**: JSON support, full-text search, etc.
4. **Better performance**: Optimized for complex queries
5. **Production ready**: Better suited for production deployments
6. **Scalability**: Can handle larger datasets and more concurrent users 