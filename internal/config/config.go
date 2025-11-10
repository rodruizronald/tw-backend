// Package config provides application configuration management using environment variables.
// It supports loading configuration from .env files and environment variables with sensible defaults.
package config

import (
	"github.com/kelseyhightower/envconfig"
)

// Config holds all configuration for the application
type Config struct {
	Gin      GinConfig
	Server   ServerConfig
	Logger   LoggerConfig
	Database DatabaseConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port string `envconfig:"PORT" default:"8080"`
}

// LoggerConfig holds debug-related configuration
type LoggerConfig struct {
	LogLevel string `envconfig:"LOG_LEVEL" default:"debug"` // Options: debug, info, warn, error
}

// GinConfig holds Gin-related configuration
type GinConfig struct {
	Mode string `envconfig:"GIN_MODE" default:"debug"` // Options: debug, release, test
}

// DatabaseConfig holds the configuration for the database connection
type DatabaseConfig struct {
	Host     string `envconfig:"DB_HOST" default:"localhost"`
	Port     int    `envconfig:"DB_PORT" default:"5432"`
	User     string `envconfig:"DB_USER" default:"postgres"`
	Password string `envconfig:"DB_PASSWORD" default:"postgres"`
	DBName   string `envconfig:"DB_NAME" default:"postgres"`
	SSLMode  string `envconfig:"DB_SSL_MODE" default:"disable"`
}

// Load reads application configuration from environment variables
// and returns a populated Config struct.
func Load() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	return &cfg, err
}
