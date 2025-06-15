package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Log is the global logger instance.
var Log zerolog.Logger

func init() {
	// Configure global settings
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.TimestampFieldName = "time"
	zerolog.CallerFieldName = "caller"
	zerolog.MessageFieldName = "message"
	zerolog.LevelFieldName = "level"

	// Create the logger with enriched context
	Log = zerolog.New(os.Stderr).With().
		Timestamp().
		Caller().
		Str("service", "draftstore").
		Logger()
}

// GetServiceLogger returns a logger with service-specific context
func GetServiceLogger(serviceName string) zerolog.Logger {
	return Log.With().
		Str("component", serviceName).
		Logger()
}

// GetHandlerLogger returns a logger for HTTP/gRPC handlers
func GetHandlerLogger(handlerType, method, path string) zerolog.Logger {
	return Log.With().
		Str("component", "handler").
		Str("type", handlerType).
		Str("method", method).
		Str("path", path).
		Logger()
}

// LogStartup logs application startup information
func LogStartup(component string, config map[string]interface{}) {
	Log.Info().
		Str("component", component).
		Interface("config", config).
		Msg("Application starting")
}

// LogShutdown logs application shutdown information
func LogShutdown(component string, duration time.Duration) {
	Log.Info().
		Str("component", component).
		Dur("uptime", duration).
		Msg("Application shutting down")
}

// LogStateChange logs important state changes with before/after values
func LogStateChange(operation string, resourceType string, resourceID string, before, after interface{}) {
	Log.Info().
		Str("operation", operation).
		Str("resource_type", resourceType).
		Str("resource_id", resourceID).
		Interface("before", before).
		Interface("after", after).
		Msg("State changed")
}
