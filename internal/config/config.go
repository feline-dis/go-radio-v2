package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	AWS      AWSConfig
	JWT      JWTConfig
	Database DatabaseConfig
	Logging  LoggingConfig
	Metrics  MetricsConfig
	Admin    AdminConfig
	YouTube  YouTubeConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type AWSConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
}

type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type LoggingConfig struct {
	Level string
}

type MetricsConfig struct {
	Enabled bool
	Port    string
}

type AdminConfig struct {
	Username string
	Password string
}

type YouTubeConfig struct {
	APIKey string
}

// Load attempts to load environment variables from .env file
// and falls back to system environment variables if not found
func Load() *Config {
	// Try to load .env file from different possible locations
	envFiles := []string{
		".env",                      // root directory
		"../.env",                   // one level up
		filepath.Join("..", ".env"), // using filepath for cross-platform compatibility
	}

	var envLoaded bool
	for _, envFile := range envFiles {
		if err := godotenv.Load(envFile); err == nil {
			log.Printf("Loaded environment from %s", envFile)
			envLoaded = true
			break
		}
	}

	if !envLoaded {
		log.Println("No .env file found, using system environment variables")
	}

	return &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			ReadTimeout:  getDurationEnv("READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getDurationEnv("WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  getDurationEnv("IDLE_TIMEOUT", 60*time.Second),
		},
		AWS: AWSConfig{
			Region:          getEnv("AWS_REGION", "us-east-1"),
			AccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
			BucketName:      getEnv("S3_BUCKET_NAME", ""),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", ""),
			Expiration: getDurationEnv("JWT_EXPIRATION", 24*time.Hour),
		},
		Database: DatabaseConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnv("POSTGRES_PORT", "5432"),
			User:     getEnv("POSTGRES_USER", "postgres"),
			Password: getEnv("POSTGRES_PASSWORD", "postgres"),
			DBName:   getEnv("POSTGRES_DB", "go_radio"),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
		},
		Logging: LoggingConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
		Metrics: MetricsConfig{
			Enabled: getBoolEnv("ENABLE_METRICS", true),
			Port:    getEnv("METRICS_PORT", "9090"),
		},
		Admin: AdminConfig{
			Username: getEnv("ADMIN_USERNAME", "admin"),
			Password: getEnv("ADMIN_PASSWORD", "admin"),
		},
		YouTube: YouTubeConfig{
			APIKey: getEnv("YOUTUBE_API_KEY", ""),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
