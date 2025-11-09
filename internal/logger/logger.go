// Package logger provides centralized logging configuration and setup for the application.
// It handles log level configuration, formatting, and provides factory methods for creating loggers.
package logger

import (
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/rodruizronald/tw-backend/internal/config"
)

// New creates and configures a new logger instance based on the provided configuration.
func New(cfg *config.LoggerConfig) *logrus.Logger {
	log := logrus.New()

	// Set log level
	level := parseLogLevel(cfg.LogLevel)
	log.SetLevel(level)

	// Configure formatter
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})

	return log
}

// parseLogLevel converts string log level to logrus.Level
func parseLogLevel(level string) logrus.Level {
	switch strings.ToLower(level) {
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn", "warning":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	case "panic":
		return logrus.PanicLevel
	default:
		return logrus.InfoLevel // Safe default
	}
}

// WithComponent adds a component field to an existing logger.
func WithComponent(baseLogger *logrus.Logger, component string) logrus.FieldLogger {
	return baseLogger.WithField("component", component)
}
